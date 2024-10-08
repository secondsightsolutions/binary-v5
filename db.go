package main

import (
	context "context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	grpc "google.golang.org/grpc"
)

type DB struct {
	tx pgx.Tx
}

func NewDB(ctx context.Context, pool *pgxpool.Pool) (*DB, error) {
	db := &DB{}
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		db.tx = tx
		return db, nil
	} else {
		return nil, err
	}
}

func (db *DB) query(ctx context.Context, pool *pgxpool.Pool, qry string) (chan map[string]any, error) {
	rchn := make(chan map[string]any, 1000)
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		if rows, err := tx.Query(ctx, qry); err == nil {
			flds := rows.FieldDescriptions()
			go func() {
				for rows.Next() {
					res := make(map[string]any)
					if vals, err := rows.Values(); err == nil {
						for i, fld := range flds {
							res[fld.Name] = vals[i]
						}
					}
					rchn <-res
				}
				rows.Close()
				tx.Commit(ctx)
				close(rchn)
			}()
		} else {
			tx.Rollback(ctx)
			return nil, err
		}
	} else {
		return nil, err
	}
	return rchn, nil
}

func (db *DB) insert(ctx context.Context, tbln string, row map[string]any) (error) {
	return nil
}
func (db *DB) insert1(ctx context.Context, tbln string, row map[string]any) (int64, error) {
	return 0, nil
}

func db_read[T any](strm grpc.ServerStreamingServer[T], pool *pgxpool.Pool, qry string, create func(map[string]any)any) error {
    ctx,f := context.WithCancel(strm.Context())
	defer f()
    if db, err := NewDB(ctx, nil); err == nil {
        if chn, err := db.query(ctx, pool, qry); err == nil {
            for obj := range chn {
                if err != nil {
                    continue
                }
                spi := create(obj)
                if err = strm.SendMsg(spi); err != nil {
                    f()
                }
            }
        } else {
            return err
        }
    } else {
        return err
    }
    return nil
}
func db_insert[T,R any](strm grpc.ClientStreamingServer[T,R], tbln string, create func(any)map[string]any) error {
    ctx,f := context.WithCancel(strm.Context())
    defer f()
    if db, err := NewDB(ctx, nil); err == nil {
        defer db.tx.Commit(ctx) // If an error occurs on db.insert(), that function will roll back the tx, making this line safe. 
        for {
            if obj, err := strm.Recv(); err == nil {
                row := create(obj)
                if err := db.insert(ctx, tbln, row); err != nil {
                    return err
                }
            } else {
                if err := db.insert(ctx, tbln, nil); err != nil {
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
func db_insert1(ctx context.Context, tbln string, obj any, create func(any)map[string]any) (int64, error) {
    ctx,f := context.WithCancel(ctx)
    defer f()
    if db, err := NewDB(ctx, nil); err == nil {
        row := create(obj)
        if id, err := db.insert1(ctx, tbln, row); err == nil {
            return id, nil
        } else {
            return -1, err
        }
    } else {
        return -1, err
    }
}

func toStr(obj any) string {
	switch val := obj.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case int16:
		return fmt.Sprintf("%d", val)
	case int32:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float32:
		return fmt.Sprintf("%f", val)
	case float64:
		return fmt.Sprintf("%f", val)
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	default:
		fmt.Printf("toStr() rcvd type %T\n", obj)
		return ""
	}
}
func toStrList(obj any) []string {
	pgx.RowToStructByName[]()
    return []string{}
}
func toU64(obj any) int64 {
    switch val := obj.(type) {
	case time.Time:
		return val.UnixMicro()
	default:
		fmt.Printf("toU64() rcvd type %T\n", obj)
		return 0
	}
}
func toI64(obj any) int64 {
    switch val := obj.(type) {
	case string:
		if i64, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i64
		} else {
			fmt.Printf("toI64() string [%s] cannot be parsed into int64", val)
			return 0
		}
	case int:
		return int64(val)
	case int16:
		return int64(val)
	case int32:
		return int64(val)
	case int64:
		return val
	case float32:
		return int64(val)
	case float64:
		return int64(val)
	case time.Time:
		return val.Unix()
	default:
		fmt.Printf("toI64() rcvd type %T\n", obj)
		return 0
	}
}
func toF64(obj any) float64 {
    switch val := obj.(type) {
	case string:
		if f64, err := strconv.ParseFloat(val, 64); err == nil {
			return f64
		} else {
			fmt.Printf("toF64() string [%s] cannot be parsed into float64", val)
			return 0
		}
	case int:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case time.Time:
		fmt.Printf("toF64() time [%v] cannot be parsed into float64", val)
		return 0
	default:
		fmt.Printf("toI64() rcvd type %T\n", obj)
		return 0
	}
}
func toBool(obj any) bool {
    switch val := obj.(type) {
	case bool:
		return val
	case string:
		if tf, err := strconv.ParseBool(val); err == nil {
			return tf
		} else {
			fmt.Printf("toBool() string [%s] cannot be parsed into bool", val)
			return false
		}
	case int:
		return val != 0
	case int16:
		return val != 0
	case int32:
		return val != 0
	case int64:
		return val != 0
	case float32:
		return val != 0
	case float64:
		return val != 0
	case time.Time:
		fmt.Printf("toBool() time [%v] cannot be parsed into bool", val)
		return false
	default:
		fmt.Printf("toBool() rcvd type %T\n", obj)
		return false
	}
}