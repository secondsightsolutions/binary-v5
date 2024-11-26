package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	grpc "google.golang.org/grpc"
)

func recv_fm[T any](pool *pgxpool.Pool, appl, name string, f func(context.Context, *SyncReq, ...grpc.CallOption) (grpc.ServerStreamingClient[T], error), stop chan any) []*T {
	chn := strm_recv_srvr(appl, name, 0, f, stop);
	lst := make([]*T, 0)
	for obj := range chn {
		lst = append(lst, obj)
	}
	return lst
}

// Server side - Server functions that receive a request and push stream data down to clients.
func strm_send_clnt[T any](appl, name string, strm grpc.ServerStreamingServer[T], fm chan *T, stop chan any) (int64, int64, error) {
	strt := time.Now()
	cnt  := int64(0)
	max  := int64(0)
	rfl  := &rflt{}
	for {
		select {
		case <-stop:	// We've been shut down from above! Must return.
			return cnt, max, nil

		case obj, ok := <-fm:
			if !ok {
				return cnt, max, nil
			}
			if err := strm.Send(obj); err == nil {
				cnt++
				seq := rfl.getFieldValueAsInt64(obj, "Seq")
				if seq > max {
					max = seq
				}
			} else {
				log(appl, "strm_send_clnt", "%s: got an error after sending %d rows", time.Since(strt), err, name, cnt)
				return cnt, max, err
			}
		}
	}
}

// Server side - server functions that have stream data pushed up.
func strm_recv_clnt[T,R any](appl, name string, strm grpc.ClientStreamingServer[T, R], stop chan any) chan *T {
	chn := make(chan *T, 1000)
	go func() {
		// Stay in this loop until either we successfully read all rows from client, or we are stopped.
		strt := time.Now()
		cnt  := 0
		for {
			select {
			case <-stop: 	// We've been shut down from above! Must return.
				close(chn)
				return

			default: 		// Not stopped yet. Read another row.
				if obj, err := strm.Recv(); err == nil {
					chn <-obj
					cnt++
				} else if err == io.EOF {
					close(chn)
					return
				} else {
					log(appl, "strm_recv_clnt", "%s: got an error after reading %d rows", time.Since(strt), err, name, cnt)
					close(chn)
					return
				}
			}
		}
	}()
	return chn
}

// Client side - client streams up to server
func strm_send_srvr[T,R any](appl, name string, f func(context.Context, ...grpc.CallOption) (grpc.ClientStreamingClient[T, R], error), fm chan *T, stop chan any) (int64, int64, error) {
	dur := time.Duration(500) * time.Millisecond
	sleep := func() bool {
		select {
		case <-stop:
			return false
		case <-time.After(dur):
			if dur < time.Duration(32) * time.Second {
				dur *= 2
			}
			return true
		}
	}
	var obj *T
	var ok  bool
	var max int64
	for {
		outer:
		c, fn := context.WithCancel(metaGRPC())
		if strm, err := f(c); err == nil {
			// Stay in this loop until either we successfully pushed all rows to server, or we are stopped.
			strt := time.Now()
			cnt  := int64(0)
			rfl  := &rflt{}
			for {
				select {
				case <-stop:	// We've been shut down from above! Must return.
					fn()
					return cnt, max, fmt.Errorf("stopped")

				default:
					if obj == nil {
						if obj, ok = <-fm; !ok {
							strm.CloseAndRecv()
							strm.CloseSend()
							fn()
							return cnt, max, nil
						}
					}
					if err := strm.Send(obj); err == nil {
						cnt++
						seq := rfl.getFieldValueAsInt64(obj, "Seq")
						if seq > max {
							max = seq
						}
						obj = nil
					} else {
						log(appl, "strm_send_srvr", "%s: got an error after sending %d rows, will retry", time.Since(strt), err, name, cnt)
						if !sleep() {
							strm.CloseAndRecv()
							strm.CloseSend()
							fn()
							return cnt, max, fmt.Errorf("stopped")
						}
						goto outer
					}
				}
			}
		}
	}
}

// Client side - server streams down to client
func strm_recv_srvr[T any](appl, name string, seq int64, f func(context.Context, *SyncReq, ...grpc.CallOption) (grpc.ServerStreamingClient[T], error), stop chan any) chan *T {
	chn := make(chan *T, 1000)
	dur := time.Duration(500) * time.Millisecond
	req := &SyncReq{Last: seq}
	sleep := func() bool {
		select {
		case <-stop:
			return false
		case <-time.After(dur):
			if dur < time.Duration(32) * time.Second {
				dur *= 2
			}
			return true
		}
	}
	go func() {
		for {
			outer:
			c, fn := context.WithCancel(metaGRPC())
			if strm, err := f(c, req); err == nil {
				// Stay in this loop until either we successfully pushed all rows to server, or we are stopped.
				strt := time.Now()
				cnt  := 0
				for {
					select {
					case <-stop:
						fn()
						close(chn)
						return

					default: // Not stopped yet. Read another row.
						if obj, err := strm.Recv(); err == nil {
							chn <-obj
						} else if err == io.EOF {
							fn()
							close(chn)
							return
						} else {
							log(appl, "strm_recv_srvr", "%s: got an error after reading %d rows, will retry", time.Since(strt), err, name, cnt)
							if !sleep() {
								fn()
								close(chn)
								return
							}
							goto outer
						}
					}
				}
			}
		}
	}()
	return chn
}