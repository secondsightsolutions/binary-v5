package main

import (
	"crypto/tls"
	"fmt"

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
	); err == nil {
		atlas.titan = NewTitanClient(conn)
	}
}

func (atlas *Atlas) getClaims(stop chan any, seq int64) chan *Claim {
	chn, err := db_select[Claim](atlas.pools["atlas"], "atlas", "atlas.claims", nil, "", "", stop)
	if err != nil {
		close(chn)
	}
	return chn
}
func (atlas *Atlas) getESP1(stop chan any, seq int64) chan *ESP1PharmNDC {
	return strm_recv_srvr[ESP1PharmNDC]("atlas", "esp1", seq, atlas.X509cert, atlas.titan.GetESP1Pharms, stop);
}
func (atlas *Atlas) getEntities(stop chan any, seq int64) chan *Entity {
	return strm_recv_srvr[Entity]("atlas", "ents", seq, atlas.X509cert, atlas.titan.GetEntities, stop);
}
func (atlas *Atlas) getLedger(stop chan any, seq int64) chan *Eligibility {
	return strm_recv_srvr[Eligibility]("atlas", "elig", seq, atlas.X509cert, atlas.titan.GetEligibilityLedger, stop);
}
func (atlas *Atlas) getNDCs(stop chan any, seq int64) chan *NDC {
	return strm_recv_srvr[NDC]("atlas", "ndcs", seq, atlas.X509cert, atlas.titan.GetNDCs, stop);
}
func (atlas *Atlas) getPharms(stop chan any, seq int64) chan *Pharmacy {
	return strm_recv_srvr[Pharmacy]("atlas", "phms", seq, atlas.X509cert, atlas.titan.GetPharmacies, stop);
}
func (atlas *Atlas) getSPIs(stop chan any, seq int64) chan *SPI {
	return strm_recv_srvr[SPI]("atlas", "spis", seq, atlas.X509cert, atlas.titan.GetSPIs, stop);
}
