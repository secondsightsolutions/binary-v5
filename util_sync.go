package main

import (
	context "context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	grpc "google.golang.org/grpc"
)

//	func db_sync_get(pool *pgxpool.Pool, appl, coln string) (int64, error) {
//		if val, err := db_select_col(context.Background(), pool, fmt.Sprintf("SELECT %s FROM %s.sync", coln, appl)); err == nil {
//			if seqn, ok := val.(int64); ok {
//				return seqn, nil
//			} else {
//				return 0, err
//			}
//		} else {
//			return 0, err
//		}
//	}
// func db_sync_set(pool *pgxpool.Pool, appl, coln string, seq int64) error {
// 	qry := fmt.Sprintf("INSERT INTO %s.sync ( pkey, %s ) VALUES ( 1, %d ) ON CONFLICT (pkey) DO UPDATE SET %s = EXCLUDED.%s ", appl, coln, seq, coln, coln)
// 	if _, err := db_exec(context.Background(), pool, qry); err != nil {
// 		return err
// 	}
// 	return nil
// }

func sync_fm_server[T any](pool *pgxpool.Pool, appl, tbln string, f func(context.Context, *SyncReq, ...grpc.CallOption) (grpc.ServerStreamingClient[T], error), stop chan any) {
	strt := time.Now()
	name := tbln
	dbm  := new_dbmap[T]()
	dbm.table(pool, tbln)
	if seqn, err := db_max(pool, tbln, "seq"); err == nil {
		chn := strm_recv_srvr(appl, name, seqn, f, stop)
		cnt, seq, err := db_insert(pool, appl, tbln, dbm, chn, 5000, "", false)
		Log(appl, "sync_fm_server", tbln, "sync completed", time.Since(strt), map[string]any{"manu": manu, "cnt": cnt, "seq": seq}, err)
	} else {
		Log(appl, "sync_fm_server", tbln, "reading seqn failed", time.Since(strt), map[string]any{"manu": manu}, err)
	}
}
func sync_to_server[T, R any](pool *pgxpool.Pool, appl, tbln string, f func(context.Context, ...grpc.CallOption) (grpc.ClientStreamingClient[T, R], error), stop chan any) {
	strt := time.Now()
	dbm  := new_dbmap[T]()
	dbm.table(pool, tbln)
	if seqn, err := db_max(pool, tbln, "seq"); err == nil {
		whr := fmt.Sprintf("seq > %d ", seqn)
		if chn, err := db_select[T](pool, appl, tbln, dbm, whr, "", stop); err == nil {
			cnt, seq, err := strm_send_srvr(appl, tbln, f, chn, stop)
			Log(appl, "sync_to_server", tbln, "sync completed", time.Since(strt), map[string]any{"manu": manu, "cnt": cnt, "seq": seq}, err)
		} else {
			Log(appl, "sync_to_server", tbln, "reading source table failed", time.Since(strt), map[string]any{"manu": manu}, err)
		}
	} else {
		Log(appl, "sync_to_server", tbln, "reading seqn failed", time.Since(strt), map[string]any{"manu": manu}, err)
	}
}
func sync_fm_client[T, R any](pool *pgxpool.Pool, appl, manu, tbln string, strm grpc.ClientStreamingServer[T, R]) (int64, int64, error) {
	strt := time.Now()
	stop := make(chan any, 1)
	dbm := new_dbmap[T]()
	dbm.table(pool, tbln)
	chn := strm_recv_clnt(appl, tbln, strm, stop)
	cnt, seq, err := db_insert(pool, appl, tbln, dbm, chn, 5000, "", false)
	Log(appl, "sync_fm_client", tbln, "sync completed", time.Since(strt), map[string]any{"manu": manu, "cnt": cnt, "seq": seq}, err)
	return cnt, seq, err
}
func sync_to_client[T any](pool *pgxpool.Pool, appl, manu, tbln, whr string, dbm *dbmap, strm grpc.ServerStreamingServer[T]) (int64, int64, error) {
	strt := time.Now()
	stop := make(chan any, 1)
	if chn, err := db_select[T](pool, appl, tbln, dbm, whr, "", stop); err == nil {
		cnt, seq, err := strm_send_clnt(appl, tbln, strm, chn, stop)
		Log(appl, "sync_to_client", tbln, "sync completed", time.Since(strt), map[string]any{"manu": manu, "cnt": cnt, "seq": seq}, err)
		return cnt, seq, err
	} else {
		Log(appl, "sync_to_client", tbln, "sync completed", time.Since(strt), map[string]any{"manu": manu}, err)
		return 0, 0, err
	}
}
