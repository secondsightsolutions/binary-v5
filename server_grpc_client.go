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
	tgt := fmt.Sprintf("%s:%s", host, port)
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
	return server_load(stop, batch, retry, "claims", srv.svc.GetClaims)
}
func (srv *Server) getESP1(stop chan any, batch, retry int) []any {
	return server_load(stop, batch, retry, "esp1", srv.svc.GetESP1Pharms)
}
func (srv *Server) getEntities(stop chan any, batch, retry int) []any {
	return server_load(stop, batch, retry, "entities", srv.svc.GetEntities)
}
func (srv *Server) getLedger(stop chan any, batch, retry int) []any {
	return server_load(stop, batch, retry, "ledger", srv.svc.GetEligibilityLedger)
}
func (srv *Server) getNDCs(stop chan any, batch, retry int) []any {
	return server_load(stop, batch, retry, "ndcs", srv.svc.GetNDCs)
}
func (srv *Server) getPharms(stop chan any, batch, retry int) []any {
	return server_load(stop, batch, retry, "pharms", srv.svc.GetPharmacies)
}
func (srv *Server) getSPIs(stop chan any, batch, retry int) []any {
	return server_load(stop, batch, retry, "esp1", srv.svc.GetSPIs)
}

func server_load[T any](stop chan any, batch, retry int, name string, f func(context.Context, *Req, ...grpc.CallOption)(grpc.ServerStreamingClient[T], error)) []any {
	title := fmt.Sprintf("[%-15s] load",  name)
	again := func(strt time.Time, wait int, msg string, err error) bool {
		log("server", title, "%s, will retry in %d seconds", time.Since(strt), err, msg, retry)
		select {
		case <-stop:
			return false
		case <-time.After(time.Duration(wait) * time.Second):
			return true
		}
	}
	req := &Req{Auth: auth, Ver: vers, Manu: manu}

	// Stay in this outer loop until either we successfully read all rows from server, or we are stopped.
	for {
		list := make([]any, 0)
		strt := time.Now()
		c,fn := context.WithCancel(context.Background())

		log("server", title, "creating stream", time.Since(strt), nil)
		if strm, err := f(c, req); err == nil {
			// Stay in this inner loop to read the stream - each time through we'll watch to see if we're being stopped.
			for {
				select {
				case <-stop:	// We've been shut down from above! Must return.
					fn()		// Cancel the context. This should let the other side know.
					return nil

				default:		// Not stopped yet. Read another row.
					if obj, err := strm.Recv(); err == nil {
						list = append(list, obj)
						if len(list)%batch == 0 {
							log("server", title, "loaded %d rows", time.Since(strt), nil, len(list))
						}
					} else if err == io.EOF {
						if len(list)%batch != 0 {
							log("server", fmt.Sprintf("[%-15s] done",  name), "loaded %d rows", time.Since(strt), nil, len(list))
						}
						fn()
						return list
					} else {
						if again(strt, 0, fmt.Sprintf("got an error after reading %d rows", len(list)), err) {
							break	// Breaks from inner loop, back into outer loop where we'll try again (immediately).
						} else {
							fn()
							return nil
						}	
					}
				}
			}
		} else {
			if !again(strt, retry, "connection to server failed", err) {	// Log, sleep, and wake up after timeout or being stopped.
				fn()
				return nil
			}
		}
	}
}
