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

func (atlas *Atlas) connect() {
	tgt := fmt.Sprintf("%s:%d", svch, svcp)
	cfg := &tls.Config{
		Certificates: []tls.Certificate{TLSCert},
		RootCAs:      X509pool,
	}
	crd := credentials.NewTLS(cfg)
	if conn, err := grpc.NewClient(tgt, grpc.WithTransportCredentials(crd)); err == nil {
		atlas.titan = NewTitanClient(conn)
	}
}

func ping() {
}

func (atlas *Atlas) getESP1(stop chan any, batch int) []any {
	return read_stream(stop, &atlas.done, batch, "esp1", atlas.opts.auth, atlas.titan.GetESP1Pharms)
}
func (atlas *Atlas) getEntities(stop chan any, batch int) []any {
	return read_stream(stop, &atlas.done, batch, "entities", atlas.opts.auth, atlas.titan.GetEntities)
}
func (atlas *Atlas) getLedger(stop chan any, batch int) []any {
	return read_stream(stop, &atlas.done, batch, "ledger", atlas.opts.auth, atlas.titan.GetEligibilityLedger)
}
func (atlas *Atlas) getNDCs(stop chan any, batch int) []any {
	return read_stream(stop, &atlas.done, batch, "ndcs", atlas.opts.auth, atlas.titan.GetNDCs)
}
func (atlas *Atlas) getPharms(stop chan any, batch int) []any {
	return read_stream(stop, &atlas.done, batch, "pharms", atlas.opts.auth, atlas.titan.GetPharmacies)
}
func (atlas *Atlas) getSPIs(stop chan any, batch int) []any {
	return read_stream(stop, &atlas.done, batch, "spis", atlas.opts.auth, atlas.titan.GetSPIs)
}

func read_stream[T any](stop chan any, done *bool, batch int, name, auth string, f func(context.Context, *Req, ...grpc.CallOption) (grpc.ServerStreamingClient[T], error)) []any {
	if *done {
		return nil
	}
	title := "read_stream"
	req := &Req{Auth: auth, Vers: vers, Manu: manu}

	// Stay in this outer loop until either we successfully read all rows from server, or we are stopped.
	for {
		outer:
		list := make([]any, 0)
		strt := time.Now()
		c, fn := context.WithCancel(context.Background())

		if strm, err := f(c, req); err == nil {
			log("atlas", title, "%s: created stream", time.Since(strt), nil, name)
			// Stay in this inner loop to read the stream - each time through we'll watch to see if we're being stopped.
			for {
				select {
				case <-stop: // We've been shut down from above! Must return.
					*done = true
					fn() // Cancel the context. This should let the other side know.
					return nil

				default: // Not stopped yet. Read another row.
					if obj, err := strm.Recv(); err == nil {
						list = append(list, obj)
						if len(list)%batch == 0 {
							log("atlas", title, "%s: loaded %d rows", time.Since(strt), nil, name, len(list))
						}
					} else if err == io.EOF {
						if len(list)%batch != 0 {
							log("atlas", title, "%s: loaded %d rows - done", time.Since(strt), nil, name, len(list))
						}
						fn()
						return list
					} else {
						log("atlas", title, "%s: got an error after reading %d rows, restarting stream read", time.Since(strt), err, name, len(list))
						goto outer
					}
				}
			}
		} else {
			log("atlas", title, "%s: error connecting to titan service", time.Since(strt), err, name)
			select {
			case <-stop:
				*done = true
				fn()
				return nil
			case <-time.After(time.Duration(10)*time.Second):
			}
		}
	}
}
