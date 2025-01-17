package main

import (
	context "context"
	"crypto/tls"
	"fmt"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func (atlas *Atlas) connect() {
	tgt := fmt.Sprintf("%s:%d", titan_grpc, titan_grpc_port)
	cfg := &tls.Config{
		Certificates: []tls.Certificate{*atlas.TLSCert},
		RootCAs:      X509pool,
	}
	crd := credentials.NewTLS(cfg)
	if conn, err := grpc.NewClient(tgt, 
		grpc.WithTransportCredentials(crd), 
		grpc.WithUnaryInterceptor(atlasUnaryClientInterceptor), 
		grpc.WithStreamInterceptor(atlasStreamClientInterceptor),
	); err == nil {
		atlas.titan = NewTitanClient(conn)
	}
}

type atlasClientStream struct {
	grpc.ClientStream
}

func atlasUnaryClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	strt := time.Now()
	ctx = addMeta(ctx, ctxValues(ctx, []string{"cmid"}))
	if err := invoker(ctx, method, req, reply, cc, opts...); err != nil {
		Log("atlas", "unary_clnt_int", method, "command failed", time.Since(strt), nil, err)
		return err
	} else {
		// Log("atlas", "unary_clnt_int", method, "command succeeded", time.Since(strt), nil, nil)
		return nil
	}
}
func atlasStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	strt := time.Now()
	ctx = addMeta(ctx, ctxValues(ctx, []string{"cmid"}))
	if s, err := streamer(ctx, desc, cc, method, opts...); err != nil {
		Log("atlas", "strm_clnt_int", method, "command failed", time.Since(strt), nil, err)
		return nil, err
	} else {
		// Log("atlas", "strm_clnt_int", method, "command succeeded", time.Since(strt), nil, nil)
		return &atlasClientStream{s}, nil
	}
}

func (w *atlasClientStream) RecvMsg(m any) error {
	return w.ClientStream.RecvMsg(m)
}

func (w *atlasClientStream) SendMsg(m any) error {
	return w.ClientStream.SendMsg(m)
}


func (atlas *Atlas) getClaims(stop chan any, seq int64) chan *Claim {
	chn, err := db_select[Claim](atlas.pools["atlas"], "atlas", "atlas.claims", nil, "", "", stop)
	if err != nil {
		close(chn)
	}
	return chn
}
func (atlas *Atlas) getESP1(stop chan any, seq int64) chan *ESP1PharmNDC {
	return strm_recv_srvr[ESP1PharmNDC]("atlas", "esp1", seq, atlas.titan.GetESP1Pharms, stop);
}
func (atlas *Atlas) getEntities(stop chan any, seq int64) chan *Entity {
	return strm_recv_srvr[Entity]("atlas", "ents", seq, atlas.titan.GetEntities, stop);
}
func (atlas *Atlas) getLedger(stop chan any, seq int64) chan *Eligibility {
	return strm_recv_srvr[Eligibility]("atlas", "elig", seq, atlas.titan.GetEligibilityLedger, stop);
}
func (atlas *Atlas) getNDCs(stop chan any, seq int64) chan *NDC {
	return strm_recv_srvr[NDC]("atlas", "ndcs", seq, atlas.titan.GetNDCs, stop);
}
func (atlas *Atlas) getPharms(stop chan any, seq int64) chan *Pharmacy {
	return strm_recv_srvr[Pharmacy]("atlas", "phms", seq, atlas.titan.GetPharmacies, stop);
}
func (atlas *Atlas) getSPIs(stop chan any, seq int64) chan *SPI {
	return strm_recv_srvr[SPI]("atlas", "spis", seq, atlas.titan.GetSPIs, stop);
}
