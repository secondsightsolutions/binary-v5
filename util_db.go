package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func db_pool(appl, host, port, name, user, pass string, tls bool) *pgxpool.Pool {
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
		//log(appl, "db_pool", "cannot create connection pool - host=%s port=%s user=%s tls=%v", time.Since(strt), err, host, port, user, tls)
		Log(appl, "db_pool", "", "cannot create connection pool - host=%s port=%s user=%s tls=%v", time.Since(strt), nil, err, host, port, user, tls)
		return nil
	} else {
		return pool
	}
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

func db_select_cust[T any](pool *pgxpool.Pool, appl string, dbm *dbmap, qry string, stop chan any) (chan *T, error) {
	if dbm == nil {
		panic("db_select_cust needs dbmap")
	}
	chn  := make(chan *T, 1000)
	obj  := new(T)
	rfl  := &rflt{}
	dbfs := dbm.mapped()
	ctx  := context.Background()
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		if rows, err := tx.Query(ctx, qry); err == nil {
			go func() {
				defer tx.Commit(ctx)
				defer rows.Close()
				for {
					select {
					case <-stop:
						Log(appl, "db_select_cust", "", "got stop signal, returning", 0, nil, nil)
						tx.Rollback(ctx)
						close(chn)
						return
					default:
						if rows.Next() {
							obj = new(T)
							if vals, err := rows.Values(); err == nil {
								for i, val := range vals {
									rfl.setFieldValue(obj, dbfs[i].fld, val)
								}
								chn <-obj
							} else {
								Log(appl, "db_select_cust", "", "getting row values failed", 0, nil, err)
								fmt.Println(qry)
								tx.Rollback(ctx)
								close(chn)
								return
							}
						} else {
							if err := rows.Err(); err != nil {
								Log(appl, "db_select_cust", "", "getting next row failed", 0, nil, err)
								fmt.Println(qry)
								dbm.Print()
							}
							close(chn)
							return
						}
					}
				}
			}()
			return chn, nil
		} else {
			Log(appl, "db_select_cust", "", "query failed, NOT closing channel", 0, nil, err)
			tx.Rollback(ctx)
			return nil, err
		}
	} else {
		return nil, err
	}
}
func db_select[T any](pool *pgxpool.Pool, appl, tbln string, dbm *dbmap, where, sort string, stop chan any) (chan *T, error) {
	if dbm == nil {
		dbm = new_dbmap[T]()
		dbm.table(pool, tbln)
	}
	chn := make(chan *T, 1000)
	obj := new(T)
	rfl := &rflt{}
	qry := dyn_select(tbln, where, sort, dbm)
	cnt := 0
	ctx := context.Background()
	// dbm.Print()
	// fmt.Println(qry)
	if dbm.tbln == "" {
		panic("dbmap not initialized with database table")
	}
	dbfs := dbm.mapped()
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		if rows, err := tx.Query(ctx, qry); err == nil {
			go func() {
				defer tx.Commit(ctx)
				defer rows.Close()
				for {
					select {
					case <-stop:
						Log(appl, "db_select", tbln, "got stop signal, returning", 0, nil, nil)
						tx.Rollback(ctx)
						close(chn)
						return
					default:
						if rows.Next() {
							cnt++
							obj = new(T)
							if vals, err := rows.Values(); err == nil {
								for i, val := range vals {
									rfl.setFieldValue(obj, dbfs[i].fld, val)
								}
								chn <-obj
							} else {
								Log(appl, "db_select", tbln, "getting row values failed", 0, nil, err)
								fmt.Println(qry)
								dbm.Print()
								tx.Rollback(ctx)
								close(chn)
								return
							}
						} else {
							if err := rows.Err(); err != nil {
								Log(appl, "db_select", tbln, "getting next row failed", 0, nil, err)
								fmt.Println(qry)
								dbm.Print()
								tx.Rollback(ctx)
							}
							close(chn)
							return
						}
					}
				}
			}()
			return chn, nil
		} else {
			Log(appl, "db_select", tbln, "query failed, NOT closing channel", 0, nil, err)
			tx.Rollback(ctx)
			return nil, err
		}
	} else {
		return nil, err
	}
}

func db_insert_run[T any](wg *sync.WaitGroup, pool *pgxpool.Pool, appl, tbln string, dbm *dbmap, fm <-chan *T, batch int, idcol string, replace, multitx bool, cnt, seq *int64, err *error, elapsed *time.Duration, stop chan any) {
	go func() {
		defer wg.Done()
		_cnt, _seq, _err := db_insert[T](pool, appl, tbln, dbm, fm, batch, idcol, replace, multitx, elapsed, stop)
		if cnt != nil {
			*cnt = _cnt
		}
		if seq != nil{
			*seq = _seq
		}
		if err != nil {
			*err = _err
		}
	}()
}

func db_insert[T any](pool *pgxpool.Pool, appl, tbln string, dbm *dbmap, fm <-chan *T, batch int, idcol string, replace, multitx bool, elapsed *time.Duration, stop chan any) (int64, int64, error) {
	if dbm == nil {
		dbm = new_dbmap[T]()
		dbm.table(pool, tbln)
	}
	ctx := context.Background()
	rfl := &rflt{}
	lst := []any{}
	cnt := int64(0)
	cur := int64(0)	// current max within the batch (once we write the batch to disk, it's our new max).
	max := int64(0)	// update to max seq found in the batch, *after* the batch is successfully inserted.
	if dbm.tbln == "" {
		panic("dbmap not initialized with database table")
	}

	if replace {
		if !strings.HasPrefix(tbln, "titan") && !strings.HasPrefix(tbln, "atlas") {
			panic("cannot delete rows from " + tbln)
		}
		// If we're replacing, then first delete all. Very important that all is done in same transaction!
		strt := time.Now()
		if cnt, err := db_exec(ctx, pool, fmt.Sprintf("DELETE FROM %s", tbln)); err == nil {
			Log(appl, "db_insert", tbln, "delete rows succeeded", time.Since(strt), map[string]any{"cnt": cnt}, nil)
		} else {
			Log(appl, "db_insert", tbln, "delete rows failed", time.Since(strt), nil, err)
			return cnt, max, err
		}
	}

	var tx  pgx.Tx
	var err error

	if multitx {
		if tx == nil {
			if tx, err = pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err != nil {
				Log(appl, "db_insert", tbln, "create transaction failed", 0, nil, err)
				return cnt, max, err
			}
		}
	}
	
	commit := func() {
		if tx != nil {
			tx.Commit(context.Background())
		}
	}
	insert := func(lst []any) error {
		if len(lst) == 0 {
			return nil
		}
		strt := time.Now()
		if err := db_insert_batch(context.Background(), tx, pool, tbln, dbm, lst, true); err != nil {
			Log(appl, "db_insert", tbln, "insert batch rows failed", time.Since(strt), map[string]any{"cnt": cnt, "seq": max}, err)
			if tx != nil {
				tx.Rollback(context.Background())
			}
			return err
		} else {
			Log(appl, "db_insert", tbln, "inserted %d (%d)", time.Since(strt), map[string]any{"cnt": cnt, "seq": max}, err, len(lst), cnt)
		}
		if elapsed != nil {
			*elapsed += time.Since(strt)
		}
		return nil
	}
	defer commit()

	for {
		select {
		case <-stop:
			return cnt, max, nil

		case obj,ok := <-fm:
			if !ok {
				goto done
			}
			if idcol != "" {
				rfl.setFieldValue(obj, idcol, cnt)	// Sets Rbid to a unique value within this scrub insert.
			}
			cnt++
			seqn := rfl.getFieldValueAsInt64(obj, "Seq")
			if seqn > cur {
				cur = seqn	// current max within the batch
			}
			lst = append(lst, obj)
			if len(lst) == batch {
				if err := insert(lst); err != nil {
					return cnt, max, err
				}
				if cur > max {	// If the max in this batch exceeds the max so far (it should), then use it (cuz we hopefully sort by Seq on input).
					max = cur
				}
				lst = []any{}
			}
		}
	}
	done:
	err = insert(lst)
	return cnt, max, err
}

func db_insert_batch(ctx context.Context, tx pgx.Tx, pool *pgxpool.Pool, tbln string, dbm *dbmap, objs []any, ignoreConflicts bool) error {
	if tx == nil {
		var err error
		if tx, err = pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err != nil {
			return err
		}
		defer tx.Commit(ctx)
	}
	if dbm.tbln == "" {
		panic("dbmap not initialized with database table")
	}
	qry := dyn_insert(tbln, dbm, objs, ignoreConflicts)
	//fmt.Println(qry)
	_, err := tx.Exec(ctx, qry)
	if err != nil {
		tx.Rollback(ctx)
		dbm.Print()
		fmt.Println(qry)
	}
	return err
}
func db_insert_one[T any](ctx context.Context, pool *pgxpool.Pool, tbln string, dbm *dbmap, obj any, rtrn string) (int64, error) {
	if dbm == nil {
		dbm = new_dbmap[T]()
		dbm.table(pool, tbln)
	}
	// dbm.Print()
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		qry := dyn_insert(tbln, dbm, []any{obj}, false)
		var err error
		if rtrn != "" {
			qry += " RETURNING " + rtrn
			//fmt.Println(qry)
			if row := tx.QueryRow(ctx, qry); row != nil {
				var id int64
				if err = row.Scan(&id); err == nil {
					if err = tx.Commit(ctx); err == nil {
						return id, nil
					}
				}
			}
		} else {
			if _, err = tx.Exec(ctx, qry); err == nil {
				if err = tx.Commit(ctx); err == nil {
					return 0, nil
				}
			}
		}
		fmt.Println(qry)
		tx.Rollback(ctx)
		return 0, err
	} else {
		return 0, err
	}
}

func db_update(ctx context.Context, obj any, tx pgx.Tx, pool *pgxpool.Pool, tbln string, dbm *dbmap, where map[string]string) error {
	mk := func(obj any) string {
		var sb bytes.Buffer
		rfl := &rflt{}	// Empty type that has the reflection convenience functions.
		dbfs := []*dbfld{}
		for _, dbf := range dbm.dbfs {
			if dbf.upd {
				dbfs = append(dbfs, dbf)
			}
		}
		sb.WriteString(fmt.Sprintf("UPDATE %s SET ", tbln))
		for j, dbf := range dbfs {
			fv := rfl.getFieldValueAsString(obj, dbf.fld)
			fv  = strings.ReplaceAll(fv, "'", "")
			cv := dbm.getColumnValueAsString(dbf.col, fv)
			sb.WriteString("\"" + dbf.col + "\"" + " = " + cv)
			if j < len(dbfs)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString(" WHERE ")
		cnt := 0
		for colN, val := range where {
			sb.WriteString("\"" + colN + "\"" + " = " + "'" + val + "'")
			cnt++
			if cnt < len(where) {
				sb.WriteString(", ")
			}
		}
		return sb.String()
	}
	if len(where) == 0 {
		panic("missing WHERE clause")
	}
	if tx == nil {
		var err error
		if tx, err = pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err != nil {
			return err
		}
		defer tx.Commit(ctx)
	}
	if dbm.tbln == "" {
		panic("dbmap not initialized with database table")
	}
	upd := mk(obj)
	//fmt.Println(upd)
	if tag, err := tx.Exec(ctx, upd); err == nil {
		// Only allow one row to be updated!
		if tag.RowsAffected() > 1 {
			tx.Rollback(ctx)
			return err
		}
	} else {
		tx.Rollback(ctx)
		fmt.Println(upd)
		return err
	}
	return nil
}

func db_exec(ctx context.Context, pool *pgxpool.Pool, qry string) (int64, error) {
	tag, err := pool.Exec(ctx, qry)
	return tag.RowsAffected(), err
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

func dyn_select(tbln string, where, sort string, dbm *dbmap) string {
	var sb bytes.Buffer
	sb.WriteString("SELECT ")
	dbfs := dbm.mapped()
	for i, dbf := range dbfs {
		if dbf.qry != "" {
			sb.WriteString(dbf.qry + " " + dbf.fld)
		} else {
			sb.WriteString(fmt.Sprintf("\"%s\"", dbf.col) + " " + dbf.fld)
		}
		if i < len(dbfs)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(" FROM " + tbln)
	if where != "" {
		sb.WriteString(" WHERE " + where)
	}
	if sort != "" {
		sb.WriteString(" ORDER BY " + sort)
	}
	return sb.String()
}
func dyn_insert[T any](tbln string, dbm *dbmap, objs []T, ignoreConflicts bool) string {
	cols := []string{}
	colq := []string{}
	dbfs := dbm.mapped_upd()
	for _, dbf := range dbfs {
		colq = append(colq, fmt.Sprintf("\"%s\"", dbf.col))
		cols = append(cols, dbf.col)
	}
	// Start building the insert query.
	var sb bytes.Buffer
	sb.WriteString(fmt.Sprintf("INSERT INTO %s ", tbln))
	sb.WriteString(fmt.Sprintf("( %s )", strings.Join(colq, ", ")))
	sb.WriteString(" VALUES ")

	rfl := &rflt{}	// Empty type that has the reflection convenience functions.
	for i, obj := range objs {
		sb.WriteString(" ( ")
		for j, colN := range cols {
			dbf := dbm.find(colN, true, true)
			fv  := rfl.getFieldValueAsString(obj, dbf.fld)
			fv  = strings.ReplaceAll(fv, "'", "")
			cv  := dbm.getColumnValueAsString(colN, fv)
			sb.WriteString(cv)
			if j < len(cols)-1 {
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
func db_max(pool *pgxpool.Pool, tbln, coln string) (int64, error) {
	if rows, err := pool.Query(context.Background(), fmt.Sprintf("SELECT COALESCE(MAX(%s), 0) max FROM %s", coln, tbln)); err == nil {
		defer rows.Close()
		var max int64
		if rows.Next() {
			err := rows.Scan(&max)
			return max, err
		} else {
			return 0, nil
		}
	} else {
		return 0, err
	}
}

func ping_db(appl, name string, pool *pgxpool.Pool) {
	strt := time.Now()
	if pool == nil {
		Log(appl, "ping_db", name, "pool not defined", time.Since(strt), nil, nil)
		//log2(appl, "ping_db", "pool not defined", name, "", nil, time.Since(strt))
		return
	}
	if err := pool.Ping(context.Background()); err == nil {
		Log(appl, "ping_db", name, "pool connected", time.Since(strt), nil, nil)
		//log2(appl, "ping_db", "pool connected", name, "", nil, time.Since(strt))
	} else {
		Log(appl, "ping_db", name, "pool not connected", time.Since(strt), nil, err)
		//log2(appl, "ping_db", "pool not connected", name, "", err, time.Since(strt))
	}
}


//func db_select_col(ctx context.Context, pool *pgxpool.Pool, qry string) (any, error) {
// 	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
// 		defer tx.Commit(context.Background())
// 		if rows, err := tx.Query(ctx, qry); err == nil {
// 			defer rows.Close()
// 			if rows.Next() {
// 				if vals, err := rows.Values(); err == nil {
// 					if len(vals) > 0 {
// 						return vals[0], nil
// 					} else {
// 						return nil, fmt.Errorf("missing column")
// 					}
// 				} else {
// 					return nil, err
// 				}
// 			} else {
// 				return nil, nil
// 			}
// 		} else {
// 			tx.Rollback(context.Background())
// 			return nil, err
// 		}
// 	} else {
// 		return nil, err
// 	}
// }
	
// type dbFldMap struct {
// 	list []*dbfld
// 	flds []string			// all fields on object (exported fields)
// 	cols []string			// all columns on table
// 	wcol []string			// all columns that have a matching field, and insertable/updatable (not auto-incrementing, etc.)
// 	f2cs map[string]string 	// fields to columns
// 	c2fs map[string]string 	// columns to fields
// 	c2tp map[string]string 	// columns to column types
// 	c2nl map[string]bool   	// columns to nullable
// 	c2df map[string]string 	// columns to default values
// 	c2qs map[string]string  // custom queries for db columns (ie, COALESCE(...))
// }

// func newDbFldMap(pool *pgxpool.Pool, tbln string, f2c, c2q map[string]string, obj any) *dbFldMap {
// 	// f2c: maps fields to columns in the case where they are a case-insensitive match (like rxn => rx_number).
// 	// Note that this is not the same as the SQL column select mappings (not used here).
// 	// There is another layer of mapping that we need when defining the columns in a SELECT list.
// 	// For example, in the above case of (rxn=>rx_number) in f2c, our actual SELECT column might need to be:
// 	// SELECT COALESCE(rx_number, '') rxn
// 	// So there are really two mapping layers:
// 	// SELECT rx_number               rxn -- this is handled via f2c (rxn=>rx_number)
// 	// SELECT COALESCE(rx_number, '') rxn -- not handled here. A SELECT-ism, and will be handled when issuing SELECT call
// 	dfm := &dbFldMap{
// 		flds: []string{},			// full list of reflect-able fields on obj
// 		cols: []string{},			// full list of columns on table
// 		wcol: []string{},			// all writable columns on table with matching field (no default values, not auto-increment)
// 		f2cs: map[string]string{},	// fields to columns (proper column name in database) (case-insens matching and f2c)
// 		c2fs: map[string]string{},	// columns to fields (mapped using case-insensitive matching and f2c)
// 		c2tp: map[string]string{},	// columns to database column types (text, integer, timestamp, etc)
// 		c2df: map[string]string{},	// columns to default values (now(), '', etc)
// 		c2qs: c2q,					// columns to custom select (COALESCE(...), etc)
// 		c2nl: map[string]bool{},	// columns to whether they are nullable
// 	}
// 	rfl := &rflt{}
// 	dfm.flds = rfl.fields(obj)

// 	// What are the actual column names on this table?
// 	schm := ""
// 	_tbl := tbln
// 	toks := strings.Split(tbln, ".")
// 	if len(toks) > 1 {
// 		schm = toks[0]
// 		_tbl = toks[1]
// 	}
// 	wcol := []string{}
// 	qry := fmt.Sprintf("select column_name, data_type, is_nullable, column_default from information_schema.columns where table_schema = '%s' and table_name = '%s';", schm, _tbl)
// 	if rows, err := pool.Query(context.Background(), qry); err == nil {
// 		for rows.Next() {
// 			var coln, colt, null string
// 			var dflt sql.NullString
// 			rows.Scan(&coln, &colt, &null, &dflt)
// 			dfm.cols = append(dfm.cols, coln)
// 			dfm.c2tp[coln] = colt
// 			dfm.c2nl[coln] = null == "YES"
// 			if dflt.Valid {
// 				if dflt.String == "''::text" {
// 					dfm.c2df[coln] = "''"
// 				} else if dflt.String == "'{}'::text[]" {
// 					dfm.c2df[coln] = "'{}'"
// 				} else {
// 					dfm.c2df[coln] = dflt.String
// 				}
// 				if !strings.HasPrefix(dflt.String, "nextval") {	// Autoincrementing columns should not be in the insertable column list.
// 					wcol = append(wcol, coln)
// 				}
// 			} else {
// 				wcol = append(wcol, coln) // A column without a default value is always insertable.
// 			}
// 		}
// 		rows.Close()
// 	}
// 	// We have our two lists - struct fields and database columns. Join them.
// 	// This first pass of associations is really only to deal with case differences.
// 	// Unless there are custom mappings - this is where we use them as overriding mappings.
// 	for _, fld := range dfm.flds {
// 		// First look to see if this field has a custom mapping (must loop to allow for case)
// 		cust := false
// 		for cfld, ccol := range f2c { 			// Look at each custom fld => col mapping (like rxn => rx_number)
// 			if strings.EqualFold(cfld, fld) { 	// We see that we have fld Rxn
// 				dfm.c2fs[ccol] = fld 			// rx_number => Rxn
// 				dfm.f2cs[fld]  = ccol 			// Rxn => rx_number (using fld, not cfld here. The caller may have messed up the case)
// 				cust  = true
// 				break
// 			}
// 		}
// 		// Hopefully most/all are direct matches between fields and columns (case notwithstanding).
// 		if !cust {
// 			for _, col := range dfm.cols {			// The full list of columns
// 				if strings.EqualFold(fld, col) { 	// The test is case-insensitive.
// 					dfm.c2fs[col] = fld 			// Save the true case.
// 					dfm.f2cs[fld] = col 			// Save the true case.
// 					break
// 				}
// 			}
// 		}
// 	}
// 	for _, col := range wcol {
// 		if _,ok := dfm.c2fs[col];ok {
// 			dfm.wcol = append(dfm.wcol, col)
// 		}
// 	}
// 	return dfm
// }
// func (dfm *dbFldMap) getColumnValueAsString(coln, colv string) string {
// 	// types: integer, bigint, text, boolean, numeric, ARRAY, date, time, timestamp, timestamp without time zone
// 	if colv == "" {
// 		if dflt, ok := dfm.c2df[coln]; ok {
// 			return dflt
// 		} else if null, ok := dfm.c2nl[coln]; ok {
// 			if null {
// 				return "NULL"
// 			}
// 		}
// 	}
// 	if ct := dfm.c2tp[coln]; ct == "ARRAY" {
// 		// Special handling if the column type is an array
// 		// The colv will already be comma-separated, so just need to wrap
// 		return fmt.Sprintf("'{%s}'", colv)
// 	}
// 	return fmt.Sprintf("'%s'", colv)
// }
// func (dfm *dbFldMap) fields() []string {
// 	flds := []string{}
// 	for _, fld := range dfm.flds {
// 		if _, ok := dfm.f2cs[fld];ok {
// 			flds = append(flds, fld)
// 		}
// 	}
// 	return flds
// }

// func db_select_one[T any](ctx context.Context, pool *pgxpool.Pool, tbln string, cols map[string]string, where, sort string) (*T, error) {
// 	qry := dyn_select(tbln, cols, where, sort)
// 	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
// 		if rows, err := tx.Query(ctx, qry); err == nil {
// 			if rows.Next() {
// 				if obj, err := pgx.RowToAddrOfStructByNameLax[T](rows); err == nil {
// 					tx.Commit(ctx)
// 					return obj, err
// 				} else {
// 					tx.Rollback(context.Background())
// 					return nil, err
// 				}
// 			} else {
// 				rows.Close()
// 				tx.Commit(ctx)
// 				return nil, nil
// 			}
// 		} else {
// 			tx.Rollback(context.Background())
// 			return nil, err
// 		}
// 	} else {
// 		return nil, err
// 	}
// }
// func db_max_seq(pool *pgxpool.Pool, tbln, coln string) (int64, error) {
// 	if rows, err := pool.Query(context.Background(), fmt.Sprintf("SELECT COALESCE(MAX(%s), 0) seq FROM %s", coln, tbln)); err == nil {
// 		defer rows.Close()
// 		var seq int64
// 		if rows.Next() {
// 			err := rows.Scan(&seq)
// 			return seq, err
// 		} else {
// 			return 0, nil
// 		}
// 	} else {
// 		return 0, err
// 	}
// }
// func db_last_seq(ctx context.Context, pool *pgxpool.Pool, tbln, coln string) (int64, error) {
// 	if rows, err := pool.Query(ctx, fmt.Sprintf("SELECT COALESCE(%s, 0) seq FROM %s", coln, tbln)); err == nil {
// 		defer rows.Close()
// 		var seq int64
// 		if rows.Next() {
// 			err := rows.Scan(&seq)
// 			return seq, err
// 		} else {
// 			return 0, nil
// 		}
// 	} else {
// 		return 0, err
// 	}
// }