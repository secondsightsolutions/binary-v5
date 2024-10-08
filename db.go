package main

import (
	"bytes"
	"context"
	"io"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	grpc "google.golang.org/grpc"
)

func db_read[T any](strm grpc.ServerStreamingServer[T], pool *pgxpool.Pool, tbln string, cols map[string]string, where string) error {
	mk := func(tbln string, cols map[string]string, where string) string {
		var sb bytes.Buffer
		sb.WriteString("SELECT ")
		cnt := 0
		for k,v := range cols {
			sb.WriteString(k + " " + v)
			if cnt < len(cols)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString(" FROM " + tbln)
		if where != "" {
			sb.WriteString(" WHERE " + where)
		}
		return sb.String()
	}
	ctx := strm.Context()
	qry := mk(tbln, cols, where)
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		if rows, err := tx.Query(ctx, qry); err == nil {
			for rows.Next() {
				if obj, err := pgx.RowToAddrOfStructByNameLax[T](rows); err == nil {
					strm.SendMsg(obj)
				} else {
					tx.Rollback(ctx)
					return err
				}
			}
			rows.Close()
			tx.Commit(ctx)
		} else {
			tx.Rollback(ctx)
			return err
		}
	} else {
		return err
	}
    return nil
}

func db_insert[T,R any](strm grpc.ClientStreamingServer[T,R], pool *pgxpool.Pool, tbln string, cols []string, toSlice func(any) []any, cnt int) error {
	ctx  := strm.Context()
	tblI := pgx.Identifier{tbln}
	rows := make([][]any, 0, cnt)
    if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		defer tx.Commit(ctx)
        for {
            if obj, err := strm.Recv(); err == nil {
				rows = append(rows, toSlice(obj))
				if len(rows) == cnt {
					csrc := pgx.CopyFromRows(rows)
					if _, err := tx.CopyFrom(ctx, tblI, cols, csrc); err == nil {
						rows = make([][]any, 0, cnt)
					} else {
						tx.Rollback(ctx)
						return err
					}
				}
            } else {
                if err != io.EOF {
					tx.Rollback(ctx)
					return err
				}
                break
            }
        }
		return nil
	} else {
		return err
	}
}

// func db_insert1(ctx context.Context, pool *pgxpool.Pool, tbln string, row map[string]any) (int64, error) {
// 	return 0, nil
// }

func pingDB(app, name string, pool *pgxpool.Pool) {
	if pool == nil {
		log(app, "pingDB", "DB pool %s not initialized yet", name)
		return
	}
	if err := pool.Ping(context.Background()); err == nil {
		log(app, "pingDB", "DB pool %s connected to database server", name)
	} else {
		log(app, "pingDB", "DB pool %s cannot ping database server: %s", name, err.Error())
	}
}