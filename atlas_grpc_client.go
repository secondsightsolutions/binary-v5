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
	if conn, err := grpc.NewClient(tgt, grpc.WithTransportCredentials(crd)); err == nil {
		atlas.titan = NewTitanClient(conn)
	}
}

func ping() {
}

func (atlas *Atlas) getClaims(stop chan any) []*Claim {
	cols := map[string]string{
		"shrt": "",
		"i340": "",
		"ndc" : "",
		"spid": "",
		"prid": "",
		"hrxn": "",
		"hfrx": "",
		"hdos": "",
		"hdop": "",
		"doc" : "",
		"dos" : "",
		"dop" : "",
		"netw": "",
		"prnm": "",
		"chnm": "",
		"elig": "",
		"susp": "",
		"cnfm": "",
		"qty" : "",
		"manu": "",
		"ihph": "",
	}
	whr := ""
	return read_db[Claim](atlas.pools["atlas"], "atlas", "atlas.claims", cols, whr, stop)
}
func (atlas *Atlas) getESP1(stop chan any) []*ESP1PharmNDC {
	return recv_fm(atlas.pools["atlas"], "atlas", "esp1", atlas.titan.GetESP1Pharms, stop)
}
func (atlas *Atlas) getEntities(stop chan any) []*Entity {
	return recv_fm(atlas.pools["atlas"], "atlas", "ents", atlas.titan.GetEntities, stop)
}
func (atlas *Atlas) getLedger(stop chan any) []*Eligibility {
	return recv_fm(atlas.pools["atlas"], "atlas", "elig", atlas.titan.GetEligibilityLedger, stop)
}
func (atlas *Atlas) getNDCs(stop chan any) []*NDC {
	return recv_fm(atlas.pools["atlas"], "atlas", "ndcs", atlas.titan.GetNDCs, stop)
}
func (atlas *Atlas) getPharms(stop chan any) []*Pharmacy {
	return recv_fm(atlas.pools["atlas"], "atlas", "phms", atlas.titan.GetPharmacies, stop)
}
func (atlas *Atlas) getSPIs(stop chan any) []*SPI {
	return recv_fm(atlas.pools["atlas"], "atlas", "spis", atlas.titan.GetSPIs, stop)
}

func (atlas *Atlas) sync(stop chan any) {
	pool := atlas.pools["atlas"]

	sync_fm_server(pool, "atlas", "atlas.claims",        "claims",          atlas.titan.GetClaims,    stop)
	sync_fm_server(pool, "atlas", "atlas.proc",          "",            	atlas.titan.GetProcs,     stop)
	sync_fm_server(pool, "atlas", "atlas.auth",          "",            	atlas.titan.GetAuths,     stop)
	sync_fm_server(pool, "atlas", "atlas.proc_auth",     "",            	atlas.titan.GetProcAuths, stop)
	sync_to_server(pool, "atlas", "atlas.scrubs",        "scrubs",			atlas.titan.Scrubs,       stop)
	sync_to_server(pool, "atlas", "atlas.rebates",       "rebates",		    atlas.titan.Rebates,      stop)
	sync_to_server(pool, "atlas", "atlas.claims_used",   "claim_uses",		atlas.titan.ClaimsUsed,   stop)
	sync_to_server(pool, "atlas", "atlas.rebate_claims", "rebate_meta",	    atlas.titan.RebateClaims, stop)
	sync_to_server(pool, "atlas", "atlas.rebate_meta",   "rebate_claims",	atlas.titan.RebateMetas,  stop)
}
