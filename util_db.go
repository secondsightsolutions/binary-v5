package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	reflect "reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	grpc "google.golang.org/grpc"
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

func db_select_one[T any](ctx context.Context, pool *pgxpool.Pool, tbln string, cols map[string]string, where string) (*T, error) {
	qry := dyn_select(tbln, cols, where)
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		if rows, err := tx.Query(ctx, qry); err == nil {
			if rows.Next() {
				pgx.RowToAddrOfStructByNameLax[T](rows)
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

// func db_select[T any](ctx context.Context, pool *pgxpool.Pool, tbln string, cols map[string]string, where string) ([]*T, error) {
// 	qry := dyn_select(tbln, cols, where)
// 	now := time.Now()
// 	lst := []*T{}
// 	cnt := 0
// 	log(appl, "db_select", qry, 0, nil)
// 	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
// 		if rows, err := tx.Query(ctx, qry); err == nil {
// 			for rows.Next() {
// 				cnt++
// 				// if cnt%100000 == 0 {
// 				// 	log(appl, "db_select", "%s: read %d rows", time.Since(now), nil, tbln, cnt)
// 				// }
// 				pgx.RowToAddrOfStructByNameLax[T](rows)
// 				if obj, err := pgx.RowToAddrOfStructByNameLax[T](rows); err == nil {
// 					lst = append(lst, obj)
// 				} else {
// 					tx.Rollback(context.Background())
// 					log(appl, "db_select", "%s: cannot map returned data to object", time.Since(now), err, tbln)
// 					return nil, err
// 				}
// 			}
// 			rows.Close()
// 			tx.Commit(ctx)
// 			log(appl, "db_select", "%s: read %d rows", time.Since(now), nil, tbln, cnt)
//     		return lst, nil
// 		} else {
// 			tx.Rollback(context.Background())
// 			log(appl, "db_select", "%s: error in query [%s]", time.Since(now), err, tbln, qry)
// 			return nil, err
// 		}
// 	} else {
// 		log(appl, "db_select", "%s: cannot create transaction", time.Since(now), err, tbln)
// 		return nil, err
// 	}
// }

func db_select_strm_to_client[T any](strm grpc.ServerStreamingServer[T], pool *pgxpool.Pool, tbln string, cols map[string]string, where string) (int, error) {
	ctx := strm.Context()
	qry := dyn_select(tbln, cols, where)
	cnt := 0
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		if rows, err := tx.Query(ctx, qry); err == nil {
			for rows.Next() {
				cnt++
				pgx.RowToAddrOfStructByNameLax[T](rows)
				if obj, err := pgx.RowToAddrOfStructByNameLax[T](rows); err == nil {
					strm.SendMsg(obj)
				} else {
					tx.Rollback(context.Background())
					return cnt, err
				}
			}
			rows.Close()
			tx.Commit(ctx)
		} else {
			tx.Rollback(context.Background())
			return cnt, err
		}
	} else {
		return cnt, err
	}
	return cnt, nil
}

func db_select_strm_to_server[T, R any](strm grpc.ClientStreamingClient[T, R], pool *pgxpool.Pool, tbln string, cols map[string]string, where string) (int, error) {
	ctx := strm.Context()
	qry := dyn_select(tbln, cols, where)
	cnt := 0
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		if rows, err := tx.Query(ctx, qry); err == nil {
			for rows.Next() {
				cnt++
				pgx.RowToAddrOfStructByNameLax[T](rows)
				if obj, err := pgx.RowToAddrOfStructByNameLax[T](rows); err == nil {
					strm.SendMsg(obj)
				} else {
					tx.Rollback(context.Background())
					return cnt, err
				}
			}
			rows.Close()
			tx.Commit(ctx)
		} else {
			tx.Rollback(context.Background())
			return cnt, err
		}
	} else {
		return cnt, err
	}
	return cnt, nil
}

func db_insert_strm_fm_client[T, R any](strm grpc.ClientStreamingServer[T, R], pool *pgxpool.Pool, tbln string, colMap map[string]string, batch int) (int, error) {
	ctx := strm.Context()
	lst := []any{}
	cnt := 0
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		defer tx.Commit(context.Background())
		for {
			if obj, err := strm.Recv(); err == nil {
				cnt++
				lst = append(lst, obj)
				if len(lst) == batch {
					if err := db_insert_batch(ctx, tx, pool, tbln, colMap, lst, true); err != nil {
						tx.Rollback(ctx)
						return cnt, err
					}
					lst = []any{}
				}
			} else if err == io.EOF {
				if len(lst) > 0 {
					if len(lst) == batch {
						if err := db_insert_batch(context.Background(), tx, pool, tbln, colMap, lst, true); err != nil {
							tx.Rollback(context.Background())
							return cnt, err
						}
					}
				}
				return cnt, nil
			} else {
				tx.Rollback(context.Background())
				return cnt, err
			}
		}
	} else {
		return cnt, err
	}
}
func db_insert_strm_fm_server[T any](strm grpc.ServerStreamingClient[T], pool *pgxpool.Pool, tbln string, colMap map[string]string, batch int) (int, error) {
	ctx := strm.Context()
	lst := []any{}
	cnt := 0
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		defer tx.Commit(context.Background())
		for {
			if obj, err := strm.Recv(); err == nil {
				cnt++
				lst = append(lst, obj)
				if len(lst) == batch {
					if err := db_insert_batch(ctx, tx, pool, tbln, colMap, lst, true); err != nil {
						tx.Rollback(context.Background())
						return cnt, err
					}
					lst = []any{}
				}
			} else if err == io.EOF {
				if len(lst) > 0 {
					if err := db_insert_batch(context.Background(), tx, pool, tbln, colMap, lst, true); err != nil {
						tx.Rollback(context.Background())
						return cnt, err
					}
				}
				return cnt, nil
			} else {
				tx.Rollback(context.Background())
				return cnt, err
			}
		}
	} else {
		return cnt, err
	}
}

func db_insert_batch(ctx context.Context, tx pgx.Tx, pool *pgxpool.Pool, tbln string, colMap map[string]string, objs []any, ignoreConflicts bool) error {
	if tx == nil {
		var err error
		if tx, err = pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err != nil {
			return err
		}
		defer tx.Commit(ctx)
	}
	qry := dyn_insert(pool, tbln, colMap, objs, ignoreConflicts)
	_, err := tx.Exec(ctx, qry)
	if err != nil {
		tx.Rollback(ctx)
	}
	return err
}
func db_insert_one(ctx context.Context, pool *pgxpool.Pool, tbln string, colMap map[string]string, obj any, rtrn string) (int64, error) {
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		defer tx.Commit(ctx)
		qry := dyn_insert(pool, tbln, colMap, []any{obj}, false)
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
		sb.WriteString(fmt.Sprintf("UPDATE %s SET ", tbln))
		for i, col := range cols {
			fn := colMap[col]
			vs := reflect.ValueOf(&obj).MethodByName(fn).Call([]reflect.Value{})
			v0 := vs[0]
			sv := "''"
			if v0.Kind() == reflect.String {
				sv = fmt.Sprintf("'%s'", v0.String())
			} else if v0.CanFloat() {
				sv = fmt.Sprintf("%f", v0.Float())
			} else if v0.CanInt() {
				sv = fmt.Sprintf("%d", v0.Int())
			} else if v0.CanUint() {
				sv = fmt.Sprintf("%d", v0.Uint())
			} else if v0.Kind() == reflect.Bool {
				sv = fmt.Sprintf("'%v'", v0.Bool())
			}
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

func db_max_seq(ctx context.Context, pool *pgxpool.Pool, tbln, coln string) (int64, error) {
	if rows, err := pool.Query(ctx, fmt.Sprintf("SELECT COALESCE(MAX(%s), 0) seq FROM %s", coln, tbln)); err == nil {
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

type dbFldMap struct {
	flds []string
	cols []string
	f2cs map[string]string // fields to columns
	c2fs map[string]string // columns to fields
}

func newDbFldMap(pool *pgxpool.Pool, tbln string, f2c map[string]string, obj any) *dbFldMap {
	dfm := &dbFldMap{
		flds: []string{},
		cols: []string{},
		f2cs: map[string]string{},
		c2fs: map[string]string{},
	}
	objV := reflect.ValueOf(obj)
	if objV.Kind() == reflect.Pointer { // Pointers to structs (ours are) must be dereferenced.
		objV = objV.Elem()
	}
	for _, vf := range reflect.VisibleFields(objV.Type()) {
		if vf.IsExported() {
			dfm.flds = append(dfm.flds, vf.Name)
		}
	}
	// What are the actual column names on this table?
	qry := fmt.Sprintf("SELECT * FROM %s LIMIT 1", tbln)
	if rows, err := pool.Query(context.Background(), qry); err == nil {
		colDes := rows.FieldDescriptions()
		for _, desc := range colDes {
			dfm.cols = append(dfm.cols, desc.Name)
		}
		rows.Close()
	}
	// We have our two lists - struct fields and database columns. Join them.
	// This first pass of associations is really only to deal with case differences.
	// Unless there are custom mappings - this is where we use them as overriding mappings.
	for _, fld := range dfm.flds {
		// First look to see if this field has a custom mapping (must loop to allow for case)
		cust := false
		for cfld, ccol := range f2c { // Look at each custom fld => col mapping (like rxn => rx_number)
			if strings.EqualFold(cfld, fld) { // We see that we have fld Rxn
				dfm.c2fs[ccol] = fld // rx_number => Rxn
				dfm.f2cs[fld] = ccol // Rxn => rx_number (using fld, not cfld here. The caller may have messed up the case)
				cust = true
				break
			}
		}
		// Hopefully most/all are direct matches between fields and columns (case notwithstanding).
		if !cust {
			for _, col := range dfm.cols {
				if strings.EqualFold(fld, col) { // The test is case-insensitive.
					dfm.c2fs[col] = fld // Save the true case.
					dfm.f2cs[fld] = col // Save the true case.
				}
			}
		}
	}
	return dfm
}

//	func (dfm *dbFldMap) print() {
//		fmt.Println()
//		fmt.Println("All flds")
//		for _, fld := range dfm.flds {
//			fmt.Println(fld)
//		}
//		fmt.Println()
//		fmt.Println("All cols")
//		for _, col := range dfm.cols {
//			fmt.Println(col)
//		}
//		fmt.Println()
//		fmt.Println("fld => col")
//		for fld, col := range dfm.f2cs {
//			fmt.Printf("%-10s => %-10s\n", fld, col)
//		}
//		fmt.Println()
//		fmt.Println("col => fld")
//		for col, fld := range dfm.c2fs {
//			fmt.Printf("%-10s => %-10s\n", col, fld)
//		}
//		fmt.Println()
//		fmt.Println("orphan flds")
//		for _, fld := range dfm.flds {
//			if col := dfm.f2cs[fld]; col == "" {
//				fmt.Println(fld)
//			}
//		}
//		fmt.Println()
//		fmt.Println("orphan cols")
//		for _, col := range dfm.cols {
//			if fld := dfm.c2fs[col]; fld == "" {
//				fmt.Println(col)
//			}
//		}
//		fmt.Println()
//	}
func (dfm *dbFldMap) reflect(obj any) *reflect.Value {
	objV := reflect.ValueOf(obj)
	if objV.Kind() == reflect.Pointer {
		objV = objV.Elem()
	}
	return &objV
}
func (dfm *dbFldMap) getFieldValueByColumn(objV *reflect.Value, col string) string {
	return dfm.getFieldValue(objV, dfm.c2fs[col])
}
func (dfm *dbFldMap) getFieldValue(objV *reflect.Value, fldn string) string {
	fld := objV.FieldByName(fldn)
	fv := "''"
	if fld.Kind() == reflect.String {
		fv = fmt.Sprintf("'%s'", fld.String())
	} else if fld.CanFloat() {
		fv = fmt.Sprintf("%f", fld.Float())
	} else if fld.CanInt() {
		fv = fmt.Sprintf("%d", fld.Int())
	} else if fld.CanUint() {
		fv = fmt.Sprintf("%d", fld.Uint())
	} else if fld.Kind() == reflect.Bool {
		fv = fmt.Sprintf("'%v'", fld.Bool())
	} else if fld.Kind() == reflect.Slice {
		var sb bytes.Buffer
		sb.WriteString("'{")
		for a := 0; a < fld.Len(); a++ {
			if sb.Len() > 2 {
				sb.WriteString(",")
			}
			e := fld.Index(a).Interface()
			switch val := e.(type) {
			case string:
				sb.WriteString(val)
			case int:
				sb.WriteString(fmt.Sprintf("%d", val))
			case int32:
				sb.WriteString(fmt.Sprintf("%d", val))
			case int64:
				sb.WriteString(fmt.Sprintf("%d", val))
			case bool:
				sb.WriteString(fmt.Sprintf("%v", val))
			default:
				sb.WriteString(" ")
			}

		}
		sb.WriteString("}'")
		fv = sb.String()
	}
	return fv
}
func dyn_select(tbln string, cols map[string]string, where string) string {
	var sb bytes.Buffer
	sb.WriteString("SELECT ")
	cnt := 0
	for k, v := range cols {
		sb.WriteString(v + " " + k)
		if cnt < len(cols)-1 {
			sb.WriteString(", ")
		}
		cnt++
	}
	sb.WriteString(" FROM " + tbln)
	if where != "" {
		sb.WriteString(" WHERE " + where)
	}
	return sb.String()
}
func dyn_insert(pool *pgxpool.Pool, tbln string, colMap map[string]string, objs []any, ignoreConflicts bool) string {
	dfm := newDbFldMap(pool, tbln, colMap, objs[0])

	// Start building the insert query.
	var sb bytes.Buffer
	sb.WriteString(fmt.Sprintf("INSERT INTO %s ", tbln))
	sb.WriteString(fmt.Sprintf("( %s )", strings.Join(dfm.cols, ", ")))
	sb.WriteString(" VALUES ")

	for i, obj := range objs {
		sb.WriteString(" ( ")
		objV := dfm.reflect(obj)
		for j, colN := range dfm.cols {
			fv := dfm.getFieldValueByColumn(objV, colN)
			sb.WriteString(fv)
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

func pingDB(app, name string, pool *pgxpool.Pool) {
	strt := time.Now()
	if pool == nil {
		log(app, "pingDB", "DB pool %s not initialized yet", time.Since(strt), nil, name)
		return
	}
	if err := pool.Ping(context.Background()); err == nil {
		log(app, "pingDB", "DB pool %s connected to database server", time.Since(strt), nil, name)
	} else {
		log(app, "pingDB", "DB pool %s cannot ping database server", time.Since(strt), err, name)
	}
}
