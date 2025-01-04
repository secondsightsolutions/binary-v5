package main

import (
	"context"
	"fmt"
	"io"
	"time"

	grpc "google.golang.org/grpc"
)

// Server side - Server functions that receive a request and push stream data down to clients.
func strm_send_clnt[T any](appl, name string, strm grpc.ServerStreamingServer[T], fm chan *T, stop chan any) (int64, int64, error) {
	strt := time.Now()
	cnt := int64(0)
	max := int64(0)
	rfl := &rflt{}
	for {
		select {
		case <-stop: // We've been shut down from above! Must return.
			Log(appl, "strm_send_clnt", name, "received stop signal, returning", time.Since(strt), map[string]any{"cnt": cnt, "seq": max}, nil)
			return cnt, max, nil

		case obj, ok := <-fm:
			if !ok {
				//Log(appl, "strm_send_clnt", name, "channel drained, returning", time.Since(strt), map[string]any{"cnt": cnt, "seq": max}, nil)
				return cnt, max, nil
			}
			if err := strm.Send(obj); err == nil {
				cnt++
				seq := rfl.getFieldValueAsInt64(obj, "Seq")
				if seq > max {
					max = seq
				}
			} else {
				Log(appl, "strm_send_clnt", name, "error sending, returning", time.Since(strt), map[string]any{"cnt": cnt, "seq": max}, err)
				return cnt, max, err
			}
		}
	}
}

// Server side - server functions that have stream data pushed up.
func strm_recv_clnt[T, R any](appl, name string, strm grpc.ClientStreamingServer[T, R], stop chan any) chan *T {
	chn := make(chan *T, 1000)
	go func() {
		// Stay in this loop until either we successfully read all rows from client, or we are stopped.
		strt := time.Now()
		cnt := 0
		for {
			select {
			case <-stop: // We've been shut down from above! Must return.
				Log(appl, "strm_recv_clnt", name, "received stop signal, returning", time.Since(strt), map[string]any{"cnt": cnt}, nil)
				close(chn)
				return

			default: // Not stopped yet. Read another row.
				if obj, err := strm.Recv(); err == nil {
					chn <- obj
					cnt++
				} else if err == io.EOF {
					//Log(appl, "strm_recv_clnt", name, "stream closed, returning", time.Since(strt), map[string]any{"cnt": cnt}, nil)
					close(chn)
					return
				} else {
					Log(appl, "strm_recv_clnt", name, "error reading, returning", time.Since(strt), map[string]any{"cnt": cnt}, err)
					close(chn)
					return
				}
			}
		}
	}()
	return chn
}

// Client side - client streams up to server
func strm_send_srvr[T, R any](appl, name string, f func(context.Context, ...grpc.CallOption) (grpc.ClientStreamingClient[T, R], error), fm chan *T, stop chan any) (int64, int64, error) {
	var obj *T
	var ok bool
	dur := time.Duration(250) * time.Millisecond
	strt := time.Now()
	rfl := &rflt{}
	cnt := int64(0)
	max := int64(0)
	erc := 0
connect:
	if dur < (32 * time.Second) {
		dur *= 2
	}
	c, fn := context.WithCancel(metaGRPC(nil))
	if strm, err := f(c); err == nil {
		// Stay in this loop until either we successfully pushed all rows to server, or we are stopped.
		for {
			select {
			case <-stop: // We've been shut down from above! Must return.
				Log(appl, "strm_send_srvr", name, "received stop signal, returning", time.Since(strt), map[string]any{"cnt": cnt, "seq": max, "erc": erc}, nil)
				strm.CloseAndRecv()
				strm.CloseSend()
				fn()
				return cnt, max, fmt.Errorf("stopped")

			default:
				if obj == nil { // Pulled off input queue on last pass, could not send to server. This must go before pulling next from input queue.
					if obj, ok = <-fm; !ok { // The input queue has closed. Nothing more to send. We're done.
						strm.CloseAndRecv()
						strm.CloseSend()
						//Log(appl, "strm_send_srvr", name, "channel drained, returning", time.Since(strt), map[string]any{"cnt": cnt, "seq": max, "erc": erc}, nil)
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
					// We failed to send the object. Close the stream and go back to the top for a reconnect/resend.
					erc++
					strm.CloseAndRecv()
					strm.CloseSend()
					fn()
					if stopped := sleep(dur, stop); stopped {
						Log(appl, "strm_send_srvr", name, "received stop signal, returning", time.Since(strt), map[string]any{"cnt": cnt, "seq": max, "erc": erc}, nil)
						return 0, max, err
					}
					goto connect
				}
			}
		}
	} else {
		erc++
		if stopped := sleep(dur, stop); stopped {
			fn()
			Log(appl, "strm_send_srvr", name, "received stop signal, returning", time.Since(strt), map[string]any{"cnt": cnt, "seq": max, "erc": erc}, err)
			return 0, max, err
		}
		goto connect
	}
}

// Client side - server streams down to client
func strm_recv_srvr[T any](appl, name string, seq int64, f func(context.Context, *SyncReq, ...grpc.CallOption) (grpc.ServerStreamingClient[T], error), stop chan any) chan *T {
	chn := make(chan *T, 1000)
	go func() {
		rfl := &rflt{}
		cnt := 0
	connect:
		req := &SyncReq{Last: seq}
		strt := time.Now()
		c, fn := context.WithCancel(metaGRPC(nil))
		if strm, err := f(c, req); err == nil {
			for {
				select {
				case <-stop:
					Log(appl, "strm_recv_srvr", name, "received stop signal, returning", time.Since(strt), map[string]any{"cnt": cnt, "seq": seq}, nil)
					fn()
					close(chn)
					return

				default: // Not stopped yet. Read another row.
					if obj, err := strm.Recv(); err == nil {
						cnt++
						seq = rfl.getFieldValueAsInt64(obj, "Seq")
						chn <- obj
					} else if err == io.EOF {
						//Log(appl, "strm_recv_srvr", name, "stream read completed", time.Since(strt), map[string]any{"cnt": cnt, "seq": seq}, nil)
						fn()
						close(chn)
						return
					} else {
						// We detected an error while reading from the stream.
						Log(appl, "strm_recv_srvr", name, "stream read failed, will retry", time.Since(strt), map[string]any{"cnt": cnt, "seq": seq}, err)
						fn()
						time.Sleep(time.Duration(5) * time.Second)
						goto connect
					}
				}
			}
		} else {
			Log(appl, "strm_recv_srvr", name, "stream connect failed, will retry", time.Since(strt), map[string]any{"cnt": cnt, "seq": seq}, err)
			fn()
			select {
			case <-stop:
				Log(appl, "strm_recv_srvr", name, "received stop signal, returning", time.Since(strt), map[string]any{"cnt": cnt, "seq": seq}, nil)
				fn()
				close(chn)
				return
			case <-time.After(time.Duration(5) * time.Second):
				goto connect
			}
		}
	}()
	return chn
}

// -----------------

// Server side - server functions that have stream data pushed up.
func strm_fmto_clnt[T, R any](appl, name string, strm grpc.BidiStreamingServer[T, R], stop chan any) (<-chan *T, chan<- *R) {
	chnT := make(chan *T, 1000)
	chnR := make(chan *R, 1000)
	// Read from the client.
	go func() {
		// Stay in this loop until either we successfully read all rows from client, or we are stopped.
		strt := time.Now()
		cnt := 0
		for {
			select {
			case <-stop: // We've been shut down from above! Must return.
				Log(appl, "strm_fmto_clnt", name, "reader received stop signal, returning", time.Since(strt), map[string]any{"cnt": cnt}, nil)
				close(chnT)
				return

			default: // Not stopped yet. Read another row.
				if obj, err := strm.Recv(); err == nil {
					chnT <- obj
					cnt++
				} else if err == io.EOF {
					Log(appl, "strm_fmto_clnt", name, "reader stream closed, returning", time.Since(strt), map[string]any{"cnt": cnt}, nil)
					close(chnT)
					return
				} else {
					Log(appl, "strm_fmto_clnt", name, "reader error reading, returning", time.Since(strt), map[string]any{"cnt": cnt}, err)
					close(chnT)
					return
				}
			}
		}
	}()
	// Send down to the client.
	go func() {
		strt := time.Now()
		cnt := 0
		for {
			select {
			case <-stop: // We've been shut down from above! Must return.
				Log(appl, "strm_fmto_clnt", name, "sender received stop signal, returning", time.Since(strt), map[string]any{"cnt": cnt}, nil)
				return

			case obj, ok := <-chnR:
				if !ok {
					Log(appl, "strm_fmto_clnt", name, "sender stream closed, returning", time.Since(strt), map[string]any{"cnt": cnt}, nil)
					return
				}
				if err := strm.Send(obj); err != nil {
					Log(appl, "strm_fmto_clnt", name, "sender error sending, returning", time.Since(strt), map[string]any{"cnt": cnt}, err)
					return
				}
			}
		}
	}()
	return chnT, chnR
}
