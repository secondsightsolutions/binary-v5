package main

import (
	context "context"
	"fmt"

	grpc "google.golang.org/grpc"
)

type titanServer struct {
	UnimplementedTitanServer
}

func (s *titanServer) Ping(ctx context.Context, req *Req) (*Res, error) {
	if err := validate_client(ctx, titan.pools["titan"], "titan"); err != nil {
		return &Res{}, err
	}
	return &Res{}, nil
}

func titan_db_read[T any](tbln string, strm grpc.ServerStreamingServer[T], seq int64) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	pool := titan.pools["titan"]
	whr  := ""
	manu := metaManu(strm.Context())
	dbm  := new_dbmap[T]()
	dbm.table(pool, tbln)
	if seq > 0 {
		whr = fmt.Sprintf("seq > %d", seq)
	}
	if dbm.byCol("manu") != nil {
		if whr == "" {
			whr = fmt.Sprintf("manu = '%s'", manu)
		} else {
			whr = fmt.Sprintf("%s AND manu = '%s'", whr, manu)
		}
	}
	sync_to_client(pool, "titan", manu, tbln, whr, dbm, strm)
	return nil
}

func (s *titanServer) GetClaims(req *SyncReq, strm grpc.ServerStreamingServer[Claim]) error {
	return titan_db_read("titan.claims", strm, req.Last)
}
func (s *titanServer) GetSPIs(req *SyncReq, strm grpc.ServerStreamingServer[SPI]) error {
	return titan_db_read("titan.spis", strm, req.Last)
}
func (s *titanServer) GetNDCs(req *SyncReq, strm grpc.ServerStreamingServer[NDC]) error {
	return titan_db_read("titan.ndcs", strm, req.Last)
}
func (s *titanServer) GetEntities(req *SyncReq, strm grpc.ServerStreamingServer[Entity]) error {
	return titan_db_read("titan.entities", strm, req.Last)
}
func (s *titanServer) GetPharmacies(req *SyncReq, strm grpc.ServerStreamingServer[Pharmacy]) error {
	return titan_db_read("titan.pharmacies", strm, req.Last)
}
func (s *titanServer) GetESP1Pharms(req *SyncReq, strm grpc.ServerStreamingServer[ESP1PharmNDC]) error {
	return titan_db_read("titan.esp1", strm, req.Last)
}
func (s *titanServer) GetEligibilityLedger(req *SyncReq, strm grpc.ServerStreamingServer[Eligibility]) error {
	return titan_db_read("titan.eligibility", strm, req.Last)
}
func (s *titanServer) GetAuths(req *SyncReq, strm grpc.ServerStreamingServer[Auth]) error {
	return titan_db_read("titan.auth", strm, req.Last)
}


func (s *titanServer) Rebates(strm grpc.ClientStreamingServer[TitanRebate, Res]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	_,_, err := sync_fm_client(titan.pools["titan"], "titan", manu, "titan.rebates", strm)
	return err
}
func (s *titanServer) Scrubs(strm grpc.ClientStreamingServer[Scrub, Res]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	_,_, err := sync_fm_client(titan.pools["titan"], "titan", manu, "titan.scrubs", strm)
	return err
}
func (s *titanServer) ClaimsUsed(strm grpc.ClientStreamingServer[ClaimUse, Res]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	_,_, err := sync_fm_client(titan.pools["titan"], "titan", manu, "titan.claim_uses", strm)
	return err
}
func (s *titanServer) RebateClaims(strm grpc.ClientStreamingServer[RebateClaim, Res]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	_,_, err := sync_fm_client(titan.pools["titan"], "titan", manu, "titan.rebate_claims", strm)
	return err
}

