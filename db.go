package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	grpc "google.golang.org/grpc"
)

func db_pool(host, port, name, user, pass, tag string, tls bool) *pgxpool.Pool {
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
		log("service", "db_pool", "cannot create connection pool - host=%s port=%s user=%s tls=%v", time.Since(strt), err, host, port, user, tls)
		return nil
	} else {
		return pool
	}
}

func db_strm_insert[T,R any](strm grpc.ClientStreamingServer[T,R], pool *pgxpool.Pool, tbln string, cols []string, toSlice func(any) []any, cnt int) error {
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

func db_strm_select[T any](strm grpc.ServerStreamingServer[T], pool *pgxpool.Pool, tbln string, cols map[string]string, where string) error {
	mk := func(tbln string, cols map[string]string, where string) string {
		var sb bytes.Buffer
		sb.WriteString("SELECT ")
		cnt := 0
		for k,v := range cols {
			sb.WriteString(k + " " + v)
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
	cnt := 0
	log("service", "db_strm_select", qry, 0, nil)
	if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
		if rows, err := tx.Query(ctx, qry); err == nil {
			for rows.Next() {
				cnt++
				if cnt%5000 == 0 {
					log("service", "db_strm_select", "%s: read %d rows", 0, nil, tbln, cnt)
				}
				if obj, err := pgx.RowToAddrOfStructByNameLax[T](rows); err == nil {
					strm.SendMsg(obj)
				} else {
					tx.Rollback(ctx)
					log("service", "db_strm_select", "%s: error1", 0, err, tbln)
					return err
				}
			}
			rows.Close()
			tx.Commit(ctx)
		} else {
			tx.Rollback(ctx)
			log("service", "db_strm_select", "%s: error2", 0, err, tbln)
			return err
		}
	} else {
		log("service", "db_strm_select", "%s: error3", 0, err, tbln)
		return err
	}
    return nil
}

func db_updates(ctx context.Context, pool *pgxpool.Pool, file string) error {
	count := func(path string) int {
        if fd, err := os.Open(path); err == nil {
            defer fd.Close()
            buf := make([]byte, 32*1024)
            count := 0
            lineSep := []byte{'\n'}
            var last byte
            for {
                c, err := fd.Read(buf)
                count += bytes.Count(buf[:c], lineSep)
                if c != 0 {
                    last = buf[c-1]
                }
                switch {
                case err == io.EOF:
                    if last != lineSep[0] {
                        count++
                    }
                    return count
                case err != nil:
                    return count
                default:
                }
            }
        }
        return 0
    }
	readl := func(br *bufio.Reader) (string, error) {
        if line, _, err := br.ReadLine(); err == nil {
            str := string(line)
            return str, nil
        } else {
            return "", err
        }
    }

	cnt   := 0
	lines := count(file)
	strt  := time.Now()
	
    if fd, err := os.Open(file); err == nil {
        defer fd.Close()
        br := bufio.NewReader(fd)
		if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
			defer tx.Commit(ctx)
			for {
				if upd, err := readl(br); err == nil {
					if tag, err := tx.Exec(ctx, upd); err == nil {
						// Only allow one row to be updated!
						if tag.RowsAffected() > 1 {
							tx.Rollback(ctx)
							log("server", "db_updates", "file=%s update (%d of %d) updated %d rows, rolling back", time.Since(strt), err, file, cnt, lines, tag.RowsAffected())
							return err
						}
					} else {
						tx.Rollback(ctx)
						log("server", "db_updates", "file=%s update (%d of %d) failed, rolling back", time.Since(strt), err, file, cnt, lines)
						return err
					}

				} else if err == io.EOF {
					return nil
				} else {
					log("server", "db_updates", "file=%s update (%d of %d) failed to read next line, rolling back", time.Since(strt), err, file, cnt, lines)
					tx.Rollback(ctx)
					return err
				}
			}
		} else {
			log("server", "db_updates", "file=%s failed to start transaction", time.Since(strt), err, file)
			return err
		}
    } else {
		log("server", "db_updates", "file=%s failed to open file", time.Since(strt), err, file)
        return err
    }
}
func db_copyfrom(ctx context.Context, pool *pgxpool.Pool, tbln, file string, size int) error {
	count := func(path string) int {
        if fd, err := os.Open(path); err == nil {
            defer fd.Close()
            buf := make([]byte, 32*1024)
            count := 0
            lineSep := []byte{'\n'}
            var last byte
            for {
                c, err := fd.Read(buf)
                count += bytes.Count(buf[:c], lineSep)
                if c != 0 {
                    last = buf[c-1]
                }
                switch {
                case err == io.EOF:
                    if last != lineSep[0] {
                        count++
                    }
                    return count
                case err != nil:
                    return count
                default:
                }
            }
        }
        return 0
    }
	readl := func(br *bufio.Reader, csep string) ([]any, error) {
        if line, _, err := br.ReadLine(); err == nil {
            str := string(line)
            if len(str) > 0 {
                toks := strings.Split(str, csep)
                vals := make([]any, 0, len(toks))
                for _, tok := range toks {
                    tok  = strings.ReplaceAll(tok, " ",  "")
                    tok  = strings.ReplaceAll(tok, "\t", "")
                    vals = append(vals, tok)
                }
                return vals, nil
            } else {
                return []any{}, nil
            }
        } else {
            return nil, err
        }
    }

	cnt   := 0
	lines := count(file)
	rows  := make([][]any, 0, size)
	tblI  := pgx.Identifier{tbln}
	hdrs := []string{}

	saveBatch := func(tx pgx.Tx) error {
		if len(rows) == 0 {
			return nil
		}
		strt := time.Now()
		csrc := pgx.CopyFromRows(rows)
		if _, err := tx.CopyFrom(ctx, tblI, hdrs, csrc); err == nil {
			cnt += len(rows)
			log("server", "db_copyfrom", "tbln=%s file=%s uploaded %d rows (%d of %d)", time.Since(strt), nil, tbln, file, len(rows), cnt, lines)
			rows = make([][]any, 0, cnt)
			return nil
		} else {
			log("server", "db_copyfrom", "tbln=%s file=%s upload %d rows (%d of %d) failed", time.Since(strt), err, tbln, file, len(rows), cnt, lines)
			tx.Rollback(ctx)
			return err
		}
	}

    if fd, err := os.Open(file); err == nil {
        defer fd.Close()
        br := bufio.NewReader(fd)
		row0, err := readl(br, ",")
        if err != nil {
            return err
        }
        for _, col := range row0 {
			hdrs = append(hdrs, col.(string))
        }
		if tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}); err == nil {
			defer tx.Commit(ctx)
			for {
				if toks, err := readl(br, ","); err == nil {
					rows = append(rows, toks)
					if len(rows) == size {
						if err := saveBatch(tx); err != nil {
							return err
						}
					}
				} else if err == io.EOF {
					if err := saveBatch(tx); err != nil {
						return err
					}
					return nil
				} else {
					tx.Rollback(ctx)
					return err
				}
			}
		} else {
			return err
		}
    } else {
        return err
    }
}

// func db_insert1(ctx context.Context, pool *pgxpool.Pool, tbln string, row map[string]any) (int64, error) {
// 	return 0, nil
// }

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