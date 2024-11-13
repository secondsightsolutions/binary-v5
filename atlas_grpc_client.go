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

func (atlas *Atlas) getESP1(stop chan any) []any {
	return read_stream(stop, &atlas.done, "esp1", atlas.opts.auth, atlas.titan.GetESP1Pharms)
}
func (atlas *Atlas) getEntities(stop chan any) []any {
	return read_stream(stop, &atlas.done, "entities", atlas.opts.auth, atlas.titan.GetEntities)
}
func (atlas *Atlas) getLedger(stop chan any) []any {
	return read_stream(stop, &atlas.done, "ledger", atlas.opts.auth, atlas.titan.GetEligibilityLedger)
}
func (atlas *Atlas) getNDCs(stop chan any) []any {
	return read_stream(stop, &atlas.done, "ndcs", atlas.opts.auth, atlas.titan.GetNDCs)
}
func (atlas *Atlas) getPharms(stop chan any) []any {
	return read_stream(stop, &atlas.done, "pharms", atlas.opts.auth, atlas.titan.GetPharmacies)
}
func (atlas *Atlas) getSPIs(stop chan any) []any {
	return read_stream(stop, &atlas.done, "spis", atlas.opts.auth, atlas.titan.GetSPIs)
}

type last_claim struct {
	Doc int64
}
type seqs struct {
	Scrubs       int64
	Rebates      int64
	ClaimUses    int64
	RebateMeta   int64
	RebateClaims int64
}

func sync_to[T, R any](pool, tbln string, last int64, f func(context.Context, ...grpc.CallOption) (grpc.ClientStreamingClient[T, R], error)) {
	strt := time.Now()
	if strm, err := f(context.Background()); err == nil {
		whr := fmt.Sprintf(" WHERE seq > %d ", last)
		if cnt, err := db_select_strm_to_server[T, R](strm, atlas.pools[pool], tbln, nil, whr); err == nil {
			log("atlas", "sync", "%s upload completed (%d rows)", time.Since(strt), nil, tbln, cnt)
		} else {
			log("atlas", "sync", "%s upload failed", time.Since(strt), err, tbln)
		}
		strm.CloseSend()
	} else {
		log("atlas", "sync", "%s upload failed", time.Since(strt), err, tbln)
	}
}
func (atlas *Atlas) sync() {
	strt := time.Now()
	cols := map[string]string{
		"doc": "COALESCE(MAX(doc), 0)",
	}
	if obj, err := db_select_one[last_claim](context.Background(), atlas.pools["atlas"], "atlas.claims", cols, ""); err == nil {
		if strm, err := atlas.titan.SyncClaims(context.Background(), &SyncReq{Manu: manu, Last: obj.Doc}); err == nil {
			if cnt, err := db_insert_strm_fm_server(strm, atlas.pools["atlas"], "atlas", "atlas.claims", nil, 5000); err == nil {
				log("atlas", "sync", "claims download completed (%d rows)", time.Since(strt), nil, cnt)
			} else {
				log("atlas", "sync", "claims download failed", time.Since(strt), err)
			}
		} else {
			log("atlas", "sync", "claims download failed", time.Since(strt), err)
		}
	} else {
		log("atlas", "sync", "failed to read last claim time", time.Since(strt), err)
		return
	}
	if obj, err := db_select_one[seqs](context.Background(), atlas.pools["atlas"], "atlas.sync", cols, ""); err == nil {
		if obj == nil {
			obj = &seqs{}
		}
		sync_to("atlas", "atlas.rebates",       obj.Rebates,      atlas.titan.Rebates)
		sync_to("atlas", "atlas.claims_used",   obj.ClaimUses,    atlas.titan.ClaimsUsed)
		sync_to("atlas", "atlas.rebate_claims", obj.RebateClaims, atlas.titan.RebateClaims)
		sync_to("atlas", "atlas.rebate_meta",   obj.RebateMeta,   atlas.titan.RebateMetas)
	} else {
		log("atlas", "sync", "failed to read last seq values", time.Since(strt), err)
		return
	}
}

func read_stream[T any](stop chan any, done *bool, name, auth string, f func(context.Context, *Req, ...grpc.CallOption) (grpc.ServerStreamingClient[T], error)) []any {
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
					} else if err == io.EOF {
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
			case <-time.After(time.Duration(10) * time.Second):
			}
		}
	}
}
