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
		log("titan", "db_pool", "cannot create connection pool - host=%s port=%s user=%s tls=%v", time.Since(strt), err, host, port, user, tls)
		return nil
	} else {
		return pool
	}
}

func db_strm_select[T any](strm grpc.ServerStreamingServer[T], pool *pgxpool.Pool, tbln string, cols map[string]string, where string) error {
	mk := func(tbln string, cols map[string]string, where string) string {
		var sb bytes.Buffer
		sb.WriteString("SELECT ")
		cnt := 0
		for k,v := range cols {
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
	ctx := strm.Context()
	qry := mk(tbln, cols, where)
	now := time.Now()
	cnt := 0
	//log("service", "db_strm_select", qry, 0, nil)
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		if rows, err := tx.Query(ctx, qry); err == nil {
			for rows.Next() {
				cnt++
				// if cnt%100000 == 0 {
				// 	log("titan", "db_strm_select", "%s: read %d rows", time.Since(now), nil, tbln, cnt)
				// }
				pgx.RowToAddrOfStructByNameLax[T](rows)
				if obj, err := pgx.RowToAddrOfStructByNameLax[T](rows); err == nil {
					strm.SendMsg(obj)
				} else {
					tx.Rollback(ctx)
					log("titan", "db_strm_select", "%s: cannot map returned data to object", 0, err, tbln)
					return err
				}
			}
			rows.Close()
			tx.Commit(ctx)
		} else {
			tx.Rollback(ctx)
			log("titan", "db_strm_select", "%s: error in query [%s]", 0, err, tbln, qry)
			return err
		}
	} else {
		log("titan", "db_strm_select", "%s: cannot create transaction", 0, err, tbln)
		return err
	}
	log("titan", "db_strm_select", "%s: read %d rows", time.Since(now), nil, tbln, cnt)
    return nil
}

func db_strm_insert[T,R any](strm grpc.ClientStreamingServer[T,R], pool *pgxpool.Pool, tbln string, colMap map[string]string, batch int) error {
	mk := func(list []any) string {
		cols := make([]string, 0, len(colMap))
		for col := range colMap {
			cols = append(cols, col)
		}
		var sb bytes.Buffer
		sb.WriteString(fmt.Sprintf("INSERT INTO %s ", tbln))
		sb.WriteString(fmt.Sprintf("( %s )", strings.Join(cols, ", ")))
		sb.WriteString(" VALUES ")
		for _, obj := range list {
			sb.WriteString(" ( ")
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
				sb.WriteString(sv)
				if i < len(cols)-1 {
					sb.WriteString(", ")
				}
			}
			sb.WriteString(" ) ")
		}
		return sb.String()
	}
	ctx := strm.Context()
	list := []any{}
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		for {
			if obj, err := strm.Recv(); err == nil {
				list = append(list, obj)
				if len(list) == batch {
					qry := mk(list)
					list = []any{}
					if _, err := tx.Exec(ctx, qry); err != nil {
						tx.Rollback(ctx)
						return err
					}
				}
			} else if err == io.EOF {
				if len(list) > 0 {
					qry := mk(list)
					if _, err := tx.Exec(ctx, qry); err != nil {
						tx.Rollback(ctx)
						return err
					}
				}
			}
		}
	}
	return nil
}

func db_insert(ctx context.Context, obj any, pool *pgxpool.Pool, tbln string, colMap map[string]string, rtrn string) (int64, error) {
	mk := func(obj any) string {
		cols := make([]string, 0, len(colMap))
		for col := range colMap {
			cols = append(cols, col)
		}
		var sb bytes.Buffer
		sb.WriteString(fmt.Sprintf("INSERT INTO %s ", tbln))
		sb.WriteString(fmt.Sprintf("( %s )", strings.Join(cols, ", ")))
		sb.WriteString(" VALUES ")
		sb.WriteString(" ( ")
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
			sb.WriteString(sv)
			if i < len(cols)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString(" ) ")
		if rtrn != "" {
			sb.WriteString("RETURNING " + rtrn)
		}
		return sb.String()
	}
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		defer tx.Commit(ctx)
		qry := mk(obj)
		if rtrn != "" {
			if row := tx.QueryRow(ctx, qry); err != nil {
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