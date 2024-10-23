package main

import (
	context "context"
	"crypto/tls"
	"fmt"
	"io"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func (srv *Server) connect() {
	tgt := fmt.Sprintf("%s:%d", svch, svcp)
	cfg := &tls.Config{
		Certificates: []tls.Certificate{TLSCert},
		RootCAs:      X509pool,
	}
	crd := credentials.NewTLS(cfg)
	if conn, err := grpc.NewClient(tgt, grpc.WithTransportCredentials(crd)); err == nil {
		srv.svc = NewBinaryV5SvcClient(conn)
	}
}

func ping() {
}

func (srv *Server) getClaims(stop chan any, batch, retry int) []any {
	return server_load(stop, &srv.done, batch, retry, "claims", srv.svc.GetClaims)
}
func (srv *Server) getESP1(stop chan any, batch, retry int) []any {
	return server_load(stop, &srv.done, batch, retry, "esp1", srv.svc.GetESP1Pharms)
}
func (srv *Server) getEntities(stop chan any, batch, retry int) []any {
	return server_load(stop, &srv.done, batch, retry, "entities", srv.svc.GetEntities)
}
func (srv *Server) getLedger(stop chan any, batch, retry int) []any {
	return server_load(stop, &srv.done, batch, retry, "ledger", srv.svc.GetEligibilityLedger)
}
func (srv *Server) getNDCs(stop chan any, batch, retry int) []any {
	return server_load(stop, &srv.done, batch, retry, "ndcs", srv.svc.GetNDCs)
}
func (srv *Server) getPharms(stop chan any, batch, retry int) []any {
	return server_load(stop, &srv.done, batch, retry, "pharms", srv.svc.GetPharmacies)
}
func (srv *Server) getSPIs(stop chan any, batch, retry int) []any {
	return server_load(stop, &srv.done, batch, retry, "esp1", srv.svc.GetSPIs)
}

func server_load[T any](stop chan any, done *bool, batch, retry int, name string, f func(context.Context, *Req, ...grpc.CallOption)(grpc.ServerStreamingClient[T], error)) []any {
	if *done {
		return nil
	}
	title := "server_load"
	again := func(strt time.Time, wait int, max *int, msg string, err error) bool {
		if *max >= retry {
			return false
		}
		*max++
		log("server", title, "%s, will retry in %d seconds", time.Since(strt), err, msg, retry)
		select {
		case <-stop:
			return false
		case <-time.After(time.Duration(wait) * time.Second):
			return true
		}
	}
	req := &Req{Auth: auth, Ver: vers, Manu: manu}
	cnt := 3

	// Stay in this outer loop until either we successfully read all rows from server, or we are stopped.
	for {
		list := make([]any, 0)
		strt := time.Now()
		c,fn := context.WithCancel(context.Background())

		log("server", title, "%s: creating stream", time.Since(strt), nil, name)
		if strm, err := f(c, req); err == nil {
			// Stay in this inner loop to read the stream - each time through we'll watch to see if we're being stopped.
			for {
				select {
				case <-stop:	// We've been shut down from above! Must return.
					fn()		// Cancel the context. This should let the other side know.
					*done = true
					return nil

				default:		// Not stopped yet. Read another row.
					if obj, err := strm.Recv(); err == nil {
						list = append(list, obj)
						if len(list)%batch == 0 {
							log("server", title, "%s: loaded %d rows", time.Since(strt), nil, name, len(list))
						}
					} else if err == io.EOF {
						if len(list)%batch != 0 {
							log("server", title, "%s: loaded %d rows - done", time.Since(strt), nil, name, len(list))
						}
						fn()
						return list
					} else {
						if again(strt, 5, &cnt, fmt.Sprintf("got an error after reading %d rows", len(list)), err) {
							break	// Breaks from inner loop, back into outer loop where we'll try again (immediately).
						} else {
							fn()
							*done = true
							return nil
						}	
					}
				}
			}
		} else {
			if !again(strt, retry, &cnt, "connection to server failed", err) {	// Log, sleep, and wake up after timeout or being stopped.
				fn()
				*done = true
				return nil
			}
		}
	}
}
