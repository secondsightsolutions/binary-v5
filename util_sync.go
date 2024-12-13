package main

import (
	context "context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	grpc "google.golang.org/grpc"
)

func db_sync_get(pool *pgxpool.Pool, appl, coln string) (int64, error) {
	if val, err := db_select_col(context.Background(), pool, fmt.Sprintf("SELECT %s FROM %s.sync", coln, appl)); err == nil {
		if seqn, ok := val.(int64); ok {
			return seqn, nil
		} else {
			return 0, err
		}
	} else {
		return 0, err
	}
}
func db_sync_set(pool *pgxpool.Pool, appl, coln string, seq int64) error {
	qry := fmt.Sprintf("INSERT INTO %s.sync ( pkey, %s ) VALUES ( 1, %d ) ON CONFLICT (pkey) DO UPDATE SET %s = EXCLUDED.%s ", appl, coln, seq, coln, coln)
	if _, err := db_exec(context.Background(), pool, qry); err != nil {
		return err
	}
	return nil
}

func sync_fm_server[T any](pool *pgxpool.Pool, appl, tbln, coln string, colmap map[string]string, f func(context.Context, *SyncReq, ...grpc.CallOption) (grpc.ServerStreamingClient[T], error), stop chan any) {
	strt := time.Now()
	name := tbln
	if coln != "" {
		// If there's a column name, it's the column name in the sync table, which means we continually add to the target table.
		if seqn, err := db_sync_get(pool, appl, coln); err == nil {
			chn := strm_recv_srvr(appl, name, seqn, f, stop)
			if cnt, seq, err := db_insert(pool, appl, tbln, colmap, chn, 5000, "", false); err == nil {
				if cnt > 0 {
					if err := db_sync_set(pool, appl, coln, seq); err != nil {
						log_sync(appl, "sync_fm_server", tbln, manu, "saving seqn failed", seqn, cnt, seq, err, time.Since(strt))
					}
				}
				log_sync(appl, "sync_fm_server", tbln, manu, "", seqn, cnt, seq, err, time.Since(strt))
			} else {
				log_sync(appl, "sync_fm_server", tbln, manu, "db insert failed", seqn, cnt, seq, err, time.Since(strt))
			}
		} else {
			log_sync(appl, "sync_fm_server", tbln, manu, "read from seq table failed", 0, 0, 0, err, time.Since(strt))
		}
	} else {
		// No column name, so it becomes a clean replacement (delete all rows, then insert from scratch)
		chn := strm_recv_srvr(appl, name, 0, f, stop)
		if cnt, seq, err := db_insert(pool, appl, tbln, nil, chn, 5000, "", true); err == nil {
			log_sync(appl, "sync_fm_server", tbln, manu, "", 0, cnt, seq, err, time.Since(strt))
		} else {
			log_sync(appl, "sync_fm_server", tbln, manu, "db replace failed", 0, cnt, seq, err, time.Since(strt))
		}
	}
	
}
func sync_to_server[T, R any](pool *pgxpool.Pool, appl, tbln, coln string, f2c map[string]string, f func(context.Context, ...grpc.CallOption) (grpc.ClientStreamingClient[T, R], error), stop chan any) {
	strt := time.Now()
	if seqn, err := db_sync_get(pool, appl, coln); err == nil {
		whr := fmt.Sprintf("seq > %d ", seqn)
		if chn, err := db_select[T](pool, tbln, f2c, whr, "", stop); err == nil {
			if cnt, seq, err := strm_send_srvr(appl, tbln, f, chn, stop); err == nil {
				if cnt > 0 {
					if err := db_sync_set(pool, appl, coln, seq); err != nil {
						log_sync(appl, "sync_to_server", tbln, manu, "saving seqn failed", seqn, cnt, seq, err, time.Since(strt))
					}
				}
				log_sync(appl, "sync_to_server", tbln, manu, "", seqn, cnt, seq, err, time.Since(strt))
			} else {
				log_sync(appl, "sync_to_server", tbln, manu, "send on stream failed", seqn, cnt, seq, err, time.Since(strt))
			}
		} else {
			log_sync(appl, "sync_to_server", tbln, manu, "read from source table failed", seqn, 0, 0, err, time.Since(strt))
		}
	} else {
		log_sync(appl, "sync_to_server", tbln, manu, "read from seq table failed", 0, 0, 0, err, time.Since(strt))
	}
}
func sync_fm_client[T,R any](pool *pgxpool.Pool, appl, manu, tbln string, strm grpc.ClientStreamingServer[T,R]) {
	strt := time.Now()
	stop := make(chan any, 1)
	chn  := strm_recv_clnt(appl, tbln, strm, stop);
	if cnt, seq, err := db_insert(pool, appl, tbln, nil, chn, 5000, "", false); err == nil {
		log_sync(appl, "sync_fm_client", tbln, manu, "", 0, cnt, seq, err, time.Since(strt))
	} else {
		log_sync(appl, "sync_fm_client", tbln, manu, "db insert failed", 0, cnt, seq, err, time.Since(strt))
	}
}
func sync_to_client[T any](pool *pgxpool.Pool, appl, manu, tbln, whr string, f2c map[string]string, strm grpc.ServerStreamingServer[T]) {
	strt := time.Now()
	stop := make(chan any, 1)
	if chn, err := db_select[T](pool, tbln, f2c, whr, "", stop); err == nil {
		if cnt, seq, err := strm_send_clnt(appl, tbln, strm, chn, stop); err == nil {
			log_sync(appl, "sync_to_client", tbln, manu, "", 0, cnt, seq, err, time.Since(strt))
		} else {
			log_sync(appl, "sync_to_client", tbln, manu, "send on stream failed", 0, cnt, seq, err, time.Since(strt))
		}
	} else {
		log_sync(appl, "sync_to_client", tbln, manu, "read from source table failed", 0, 0, 0, err, time.Since(strt))
	}
}
