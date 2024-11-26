package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func db_pool(host, port, name, user, pass string, tls bool) *pgxpool.Pool {
	strt := time.Now()
	cs := []string{}
	cs = append(cs, fmt.Sprintf("%s=%s", "host", host))
	cs = append(cs, fmt.Sprintf("%s=%s", "port", port))
	cs = append(cs, fmt.Sprintf("%s=%s", "user", user))
	cs = append(cs, fmt.Sprintf("%s=%s", "password", pass))
	cs = append(cs, fmt.Sprintf("%s=%s", "dbname", name))
	cs = append(cs, fmt.Sprintf("%s=%s", "default_query_exec_mode", "exec"))
	cs = append(cs, fmt.Sprintf("%s=%d", "pool_min_conns", 2))
	cs = append(cs, fmt.Sprintf("%s=%d", "pool_max_conns", 8))
	if tls {
		cs = append(cs, fmt.Sprintf("%s=%s", "sslmode", "require"))
	}
	if pool, err := pgxpool.New(context.Background(), strings.Join(cs, " ")); err != nil {
		log(appl, "db_pool", "cannot create connection pool - host=%s port=%s user=%s tls=%v", time.Since(strt), err, host, port, user, tls)
		return nil
	} else {
		return pool
	}
}

func read_db[T any](pool *pgxpool.Pool, appl, tbln string, cols map[string]string,  where string, stop chan any) []*T {
	strt := time.Now()
	lst := make([]*T, 0)
	if chn, err := db_select[T](pool, tbln, cols, where, "", stop); err == nil {
		for obj := range chn {
			lst = append(lst, obj)
		}
	} else {
		log(appl, "read_db", "%s read db failed - failed to read rows from table", time.Since(strt), err, tbln)
	}
	return lst
}

func db_select_col(ctx context.Context, pool *pgxpool.Pool, qry string) (any, error) {
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		defer tx.Commit(context.Background())
		if rows, err := tx.Query(ctx, qry); err == nil {
			defer rows.Close()
			if rows.Next() {
				if vals, err := rows.Values(); err == nil {
					if len(vals) > 0 {
						return vals[0], nil
					} else {
						return nil, fmt.Errorf("missing column")
					}
				} else {
					return nil, err
				}
			} else {
				return nil, nil
			}
		} else {
			tx.Rollback(context.Background())
			return nil, err
		}
	} else {
		return nil, err
	}
}
func db_select_one[T any](ctx context.Context, pool *pgxpool.Pool, tbln string, cols map[string]string, where, sort string) (*T, error) {
	qry := dyn_select(tbln, cols, where, sort)
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		if rows, err := tx.Query(ctx, qry); err == nil {
			if rows.Next() {
				if obj, err := pgx.RowToAddrOfStructByNameLax[T](rows); err == nil {
					tx.Commit(ctx)
					return obj, err
				} else {
					tx.Rollback(context.Background())
					return nil, err
				}
			} else {
				rows.Close()
				tx.Commit(ctx)
				return nil, nil
			}
		} else {
			tx.Rollback(context.Background())
			return nil, err
		}
	} else {
		return nil, err
	}
}

func db_select[T any](pool *pgxpool.Pool, tbln string, cols map[string]string, where, sort string, stop chan any) (chan *T, error) {
	chn := make(chan *T, 1000)
	qry := dyn_select(tbln, cols, where, sort)
	cnt := 0
	ctx := context.Background()
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		if rows, err := tx.Query(ctx, qry); err == nil {
			go func() {
				defer tx.Commit(ctx)
				defer rows.Close()
				for {
					select {
					case <-stop:
						tx.Rollback(ctx)
						close(chn)
						return
					default:
						if rows.Next() {
							cnt++
							if obj, err := pgx.RowToAddrOfStructByNameLax[T](rows); err == nil {
								chn <-obj
							} else {
								tx.Rollback(ctx)
								close(chn)
								return
							}
						} else {
							close(chn)
							return
						}
					}
				}
			}()
			return chn, nil
		} else {
			tx.Rollback(ctx)
			return nil, err
		}
	} else {
		return nil, err
	}
}

func db_insert[T any](pool *pgxpool.Pool, appl, tbln string, cols map[string]string, fm chan *T, batch int) (int64, int64, error) {
	var dfm *dbFldMap
	ctx := context.Background()
	rfl := &rflt{}
	lst := []any{}
	cnt := int64(0)
	cur := int64(0)	// current max within the batch (once we write the batch to disk, it's our new max).
	max := int64(0)	// update to max seq found in the batch, *after* the batch is successfully inserted.
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		defer tx.Commit(context.Background())
		for obj := range fm {
			if dfm == nil {		// Assume the dfm can be the same across all objs in list (same object type).
				dfm = newDbFldMap(pool, tbln, cols, obj)
				if len(dfm.unqC) > 0 || len(dfm.unqF) > 0 {
					log(appl, "db_insert", "tbln=%s uniq-cols:%s uniq-flds: %s", 0, nil, tbln, strings.Join(dfm.unqC, ","), strings.Join(dfm.unqF, ","))
				}
			}
			cnt++
			seqn := rfl.getFieldValueAsInt64(obj, "Seq")
			if seqn > cur {
				cur = seqn	// current max within the batch
			}
			lst = append(lst, obj)
			if len(lst) == batch {
				if err := db_insert_batch(ctx, tx, pool, tbln, dfm, lst, true); err != nil {
					tx.Rollback(ctx)
					return cnt, max, err
				}
				if cur > max {	// If the max in this batch exceeds the max so far (it should), then use it (cuz we hopefully sort by Seq on input).
					max = cur
				}
				lst = []any{}
			}
		}
		if len(lst) > 0 {
			if err := db_insert_batch(context.Background(), tx, pool, tbln, dfm, lst, true); err != nil {
				tx.Rollback(context.Background())
				return cnt, max, err
			}
			if cur > max {	// If the max in this batch exceeds the max so far (it should), then use it (cuz we hopefully sort by Seq on input).
				max = cur
			}
		}
	} else {
		return cnt, max, err
	}
	return cnt, max, nil
}

func db_insert_batch(ctx context.Context, tx pgx.Tx, pool *pgxpool.Pool, tbln string, dfm *dbFldMap, objs []any, ignoreConflicts bool) error {
	if tx == nil {
		var err error
		if tx, err = pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err != nil {
			return err
		}
		defer tx.Commit(ctx)
	}
	qry := dyn_insert(pool, tbln, dfm, objs, ignoreConflicts)
	_, err := tx.Exec(ctx, qry)
	if err != nil {
		tx.Rollback(ctx)
	}
	return err
}
func db_insert_one(ctx context.Context, pool *pgxpool.Pool, tbln string, dfm *dbFldMap, obj any, rtrn string) (int64, error) {
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		defer tx.Commit(ctx)
		qry := dyn_insert(pool, tbln, dfm, []any{obj}, false)
		if rtrn != "" {
			qry += " RETURNING " + rtrn
			if row := tx.QueryRow(ctx, qry); row != nil {
				var id int64
				if err := row.Scan(&id); err != nil {
					tx.Rollback(ctx)
					return 0, err
				} else {
					return id, nil
				}
			} else {
				return 0, err
			}
		} else {
			if _, err := tx.Exec(ctx, qry); err != nil {
				tx.Rollback(ctx)
				return 0, err
			} else {
				return 0, nil
			}
		}
	} else {
		return 0, err
	}
}

func db_update(ctx context.Context, obj any, pool *pgxpool.Pool, tbln string, colMap map[string]string, where map[string]string) error {
	mk := func(obj any) string {
		cols := make([]string, 0, len(colMap))
		for col := range colMap {
			cols = append(cols, col)
		}
		var sb bytes.Buffer
		rfl := &rflt{}
		sb.WriteString(fmt.Sprintf("UPDATE %s SET ", tbln))
		for i, col := range cols {
			fn := colMap[col]
			sv := fmt.Sprintf("'%s'", rfl.getString(obj, fn))
			sb.WriteString(col + " = " + sv)
			if i < len(cols)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString(" WHERE ")
		cnt := 0
		for col, val := range where {
			sb.WriteString(col + " = '" + val + "'")
			cnt++
			if cnt < len(where) {
				sb.WriteString(", ")
			}
		}
		return sb.String()
	}
	if len(where) == 0 {
		return fmt.Errorf("missing WHERE")
	}
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		defer tx.Commit(ctx)
		upd := mk(obj)
		if tag, err := tx.Exec(ctx, upd); err == nil {
			// Only allow one row to be updated!
			if tag.RowsAffected() > 1 {
				tx.Rollback(ctx)
				return err
			}
		} else {
			tx.Rollback(ctx)
			return err
		}
	} else {
		return err
	}
	return nil
}

func db_exec(ctx context.Context, pool *pgxpool.Pool, qry string) (int64, error) {
	tag, err := pool.Exec(ctx, qry)
	return tag.RowsAffected(), err
}

func db_last_seq(ctx context.Context, pool *pgxpool.Pool, tbln, coln string) (int64, error) {
	if rows, err := pool.Query(ctx, fmt.Sprintf("SELECT COALESCE(%s, 0) seq FROM %s", coln, tbln)); err == nil {
		defer rows.Close()
		var seq int64
		if rows.Next() {
			err := rows.Scan(&seq)
			return seq, err
		} else {
			return 0, nil
		}
	} else {
		return 0, err
	}
}
func db_max_seq(pool *pgxpool.Pool, tbln, coln string) (int64, error) {
	if rows, err := pool.Query(context.Background(), fmt.Sprintf("SELECT COALESCE(MAX(%s), 0) seq FROM %s", coln, tbln)); err == nil {
		defer rows.Close()
		var seq int64
		if rows.Next() {
			err := rows.Scan(&seq)
			return seq, err
		} else {
			return 0, nil
		}
	} else {
		return 0, err
	}
}
func db_count(ctx context.Context, pool *pgxpool.Pool, frmWhr string) (int64, error) {
	if rows, err := pool.Query(ctx, fmt.Sprintf("SELECT COUNT(*) count %s ", frmWhr)); err == nil {
		defer rows.Close()
		var cnt int64
		if rows.Next() {
			err := rows.Scan(&cnt)
			return cnt, err
		} else {
			return 0, nil
		}
	} else {
		return 0, err
	}
}

type dbFldMap struct {
	flds []string
	cols []string
	f2cs map[string]string // fields to columns
	c2fs map[string]string // columns to fields
	unqF []string
	unqC []string
}

func newDbFldMap(pool *pgxpool.Pool, tbln string, f2c map[string]string, obj any) *dbFldMap {
	dfm := &dbFldMap{
		flds: []string{},
		cols: []string{},
		f2cs: map[string]string{},
		c2fs: map[string]string{},
		unqF: []string{},
		unqC: []string{},
	}
	rfl := &rflt{}
	dfm.flds = rfl.fields(obj)

	// What are the actual column names on this table?
	schm := ""
	_tbl := tbln
	toks := strings.Split(tbln, ".")
	if len(toks) > 1 {
		schm = toks[0]
		_tbl = toks[1]
	}
	qry := fmt.Sprintf("select column_name, data_type from information_schema.columns where table_schema = '%s' and table_name = '%s';", schm, _tbl)
	if rows, err := pool.Query(context.Background(), qry); err == nil {
		for rows.Next() {
			var coln, colt string
			rows.Scan(&coln, &colt)
			dfm.cols = append(dfm.cols, coln)
		}
		rows.Close()
	}
	// We have our two lists - struct fields and database columns. Join them.
	// This first pass of associations is really only to deal with case differences.
	// Unless there are custom mappings - this is where we use them as overriding mappings.
	for _, fld := range dfm.flds {
		found := false
		// First look to see if this field has a custom mapping (must loop to allow for case)
		cust := false
		for cfld, ccol := range f2c { // Look at each custom fld => col mapping (like rxn => rx_number)
			if strings.EqualFold(cfld, fld) { // We see that we have fld Rxn
				dfm.c2fs[ccol] = fld // rx_number => Rxn
				dfm.f2cs[fld] = ccol // Rxn => rx_number (using fld, not cfld here. The caller may have messed up the case)
				cust  = true
				found = true
				break
			}
		}
		// Hopefully most/all are direct matches between fields and columns (case notwithstanding).
		if !cust {
			for _, col := range dfm.cols {
				if strings.EqualFold(fld, col) { // The test is case-insensitive.
					dfm.c2fs[col] = fld // Save the true case.
					dfm.f2cs[fld] = col // Save the true case.
					found = true
					break
				}
			}
		}
		if !found && !strings.EqualFold(fld, "Seq") {	// Special column, ignore if only on one side.
			dfm.unqF = append(dfm.unqF, fld)
		}
	}
	for _, col := range dfm.cols {
		if _,ok := dfm.c2fs[col];!ok {
			dfm.unqC = append(dfm.unqC, col)
		}
	}
	return dfm
}

func dyn_select(tbln string, cols map[string]string, where, sort string) string {
	var sb bytes.Buffer
	sb.WriteString("SELECT ")
	seq := false
	if len(cols) > 0 {
		cnt := 0
		for k, v := range cols {
			sb.WriteString(v + " " + k)
			if cnt < len(cols)-1 {
				sb.WriteString(", ")
			}
			cnt++
			if strings.EqualFold(k, "seq") {
				seq = true
			}
		}
	} else {
		sb.WriteString(" * ")
	}
	sb.WriteString(" FROM " + tbln)
	if where != "" {
		sb.WriteString(" WHERE " + where)
	}
	if seq {
		sb.WriteString(" ORDER BY seq")
		if sort != "" {
			sb.WriteString(", " + sort)
		}
	} else {
		if sort != "" {
			sb.WriteString(" ORDER BY " + sort)
		}
	}
	return sb.String()
}
func dyn_insert(pool *pgxpool.Pool, tbln string, dfm *dbFldMap, objs []any, ignoreConflicts bool) string {
	//dfm := newDbFldMap(pool, tbln, colMap, objs[0])

	// Start building the insert query.
	var sb bytes.Buffer
	sb.WriteString(fmt.Sprintf("INSERT INTO %s ", tbln))
	sb.WriteString(fmt.Sprintf("( %s )", strings.Join(dfm.cols, ", ")))
	sb.WriteString(" VALUES ")

	rfl := &rflt{}	// Empty type that has the reflection convenience functions.
	for i, obj := range objs {
		sb.WriteString(" ( ")
		for j, colN := range dfm.cols {
			fv := rfl.getFieldValueAsString(obj, dfm.c2fs[colN])
			sb.WriteString("'")
			sb.WriteString(fv)
			sb.WriteString("'")
			if j < len(dfm.cols)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString(" ) ")
		if i < len(objs)-1 {
			sb.WriteString(", ")
		}
	}
	if ignoreConflicts {
		sb.WriteString(" ON CONFLICT DO NOTHING")
	}
	return sb.String()
}

func ping_db(app, name string, pool *pgxpool.Pool) {
	strt := time.Now()
	if pool == nil {
		log(app, "ping_db", "%-21s / %s", time.Since(strt), nil, "pool not defined", name)
		return
	}
	if err := pool.Ping(context.Background()); err == nil {
		log(app, "ping_db", "%-21s / %s", time.Since(strt), nil, "pool connected", name)
	} else {
		log(app, "ping_db", "%-21s / %s", time.Since(strt), err, "pool not connected", name)
	}
}
