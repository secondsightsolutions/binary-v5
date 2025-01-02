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

func (s *titanServer) GetClaims(req *SyncReq, strm grpc.ServerStreamingServer[Claim]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	pool := titan.pools["citus"]
	tbln := "public.submission_rows"
	dbm  := new_dbmap[Claim]()
	dbm.column("chnm", "chain_name", 				"COALESCE(chain_name, '')")
	dbm.column("cnfm", "claim_conforms_flag", 		"COALESCE(claim_conforms_flag, true)")
	dbm.column("seq",  "created_at",  				"COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)   *1000000, 0), 0)")
	dbm.column("doc",  "created_at",  				"COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)   *1000000, 0), 0)")
	dbm.column("dop",  "formatted_dop",  			"COALESCE(TRUNC(EXTRACT(EPOCH FROM formatted_dop)*1000000, 0), 0)")
	dbm.column("dos",  "formatted_dos",  			"COALESCE(TRUNC(EXTRACT(EPOCH FROM formatted_dos)*1000000, 0), 0)")
	dbm.column("hdop", "date_prescribed", 			"COALESCE(date_prescribed, '')")
	dbm.column("hdos", "date_of_service", 			"COALESCE(date_of_service, '')")
	dbm.column("hfrx", "formatted_rx_number", 		"COALESCE(formatted_rx_number, '')")
	dbm.column("hrxn", "rx_number", 				"COALESCE(rx_number, '')")
	dbm.column("i340", "id_340b", 					"SPLIT_PART(COALESCE(id_340b, ''), '-', 1)")
	dbm.column("manu", "manufacturer", 				"COALESCE(manufacturer, '')")
	dbm.column("ndc",  "ndc",  						"REPLACE(COALESCE(ndc, ''), '-', '')")
	dbm.column("netw", "network", 					"COALESCE(network, '')")
	dbm.column("prnm", "product_name", 				"COALESCE(product_name, '')")
	dbm.column("qty",  "quantity",  				"COALESCE(quantity, 0)")
	dbm.column("shrt", "short_id", 					"COALESCE(short_id, '')")
	dbm.column("spid", "service_provider_id", 		"COALESCE(service_provider_id, '')")
	dbm.column("prid", "prescriber_id", 			"COALESCE(prescriber_id, '')")
	dbm.column("elig", "eligible_at_submission",	"COALESCE(eligible_at_submission, true)")
	dbm.column("susp", "suspended_submission", 		"COALESCE(suspended_submission, false)")
	dbm.column("ihph", "in_house_pharmacy_ids", 	"array_to_string(in_house_pharmacy_ids, ',')")
	dbm.table(pool, tbln)

	manu := metaManu(strm.Context())
	whr  := fmt.Sprintf("manufacturer = '%s' AND COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)*1000000, 0), 0) > %d", manu, req.Last)
	sync_to_client(pool, "titan", manu, tbln, whr, dbm, strm)
	return nil
}
func (s *titanServer) GetSPIs(req *SyncReq, strm grpc.ServerStreamingServer[SPI]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	pool := titan.pools["esp"]
	tbln := "public.ncpdp_providers"
	dbm  := new_dbmap[SPI]()
	dbm.column("ncp", "ncpdp_provider_id", 			"COALESCE(ncpdp_provider_id, '')")
	dbm.column("npi", "national_provider_id", 		"COALESCE(national_provider_id, '')")
	dbm.column("dea", "dea_registration_id", 		"COALESCE(dea_registration_id, '')")
	dbm.column("sto", "store_number", 				"COALESCE(store_number, '')")
	dbm.column("lbn", "legal_business_name", 		"COALESCE(legal_business_name, '')")
	dbm.column("cde", "status_code_340b", 			"COALESCE(status_code_340b, '')")
	dbm.column("chn", "chain_name", 				"COALESCE(chain_name, '')")
	dbm.column("nam", "name", 						"COALESCE(name, '')")
	dbm.table(pool, tbln)
	
	whr := fmt.Sprintf("COALESCE(id, 0) > %d", req.Last)
	sync_to_client(pool, "titan", manu, tbln, whr, dbm, strm)
	return nil
}
func (s *titanServer) GetNDCs(req *SyncReq, strm grpc.ServerStreamingServer[NDC]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	pool := titan.pools["esp"]
	tbln := "public.ndcs"
	dbm  := new_dbmap[NDC]()
	dbm.column("ndc", "item",  						"COALESCE(REPLACE(item, '-', ''), '')")
	dbm.column("name", "product_name", 				"COALESCE(product_name, '')")
	dbm.column("netw", "network", 					"COALESCE(network, '')")
	dbm.table(pool, tbln)
	
	manu := metaManu(strm.Context())
	whr := fmt.Sprintf("manufacturer_name = '%s' AND COALESCE(id, 0) > %d", manu, req.Last)
	sync_to_client(pool, "titan", manu, tbln, whr, dbm, strm)
	return nil
}
func (s *titanServer) GetEntities(req *SyncReq, strm grpc.ServerStreamingServer[Entity]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	pool := titan.pools["esp"]
	tbln := "public.covered_entities"
	dbm := new_dbmap[Entity]()
	dbm.column("i340", "id_340b",  					"COALESCE(id_340b, '')")
	dbm.column("state", "state", 					"COALESCE(state, '')")
	dbm.column("strt", "participating_start_date",	"COALESCE(TRUNC(EXTRACT(EPOCH FROM participating_start_date::timestamp) *1000000, 0), 0)")
	dbm.column("term", "term_date",  				"COALESCE(TRUNC(EXTRACT(EPOCH FROM term_date::timestamp)                *1000000, 0), 0)")
	dbm.table(pool, tbln)

	whr := fmt.Sprintf("COALESCE(id, 0) > %d", req.Last)
	sync_to_client(pool, "titan", manu, tbln, whr, dbm, strm)
	return nil
}
func (s *titanServer) GetPharmacies(req *SyncReq, strm grpc.ServerStreamingServer[Pharmacy]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	pool := titan.pools["esp"]
	tbln := "public.contracted_pharmacies"
	dbm  := new_dbmap[Pharmacy]()
	dbm.column("chnm", "chain_name",  			"COALESCE(chain_name, '')")
	dbm.column("i340", "id_340b",  				"COALESCE(id_340b, '')")
	dbm.column("phid", "pharmacy_id", 			"COALESCE(pharmacy_id, '')")
	dbm.column("dea",  "dea_id",   				"COALESCE(dea_id, '')")
	dbm.column("npi",  "national_provider_id",	"COALESCE(national_provider_id, '')")
	dbm.column("ncp",  "ncpdp_provider_id",   	"COALESCE(ncpdp_provider_id, '')")
	dbm.column("deas", "dea",  					"array_to_string(dea, ',')")
	dbm.column("npis", "npi",  					"array_to_string(npi, ',')")
	dbm.column("ncps", "ncpdp",  				"array_to_string(ncpdp, ',')")
	dbm.column("state", "pharmacy_state", 		"COALESCE(pharmacy_state, '')")
	dbm.table(pool, tbln)

	whr := fmt.Sprintf("COALESCE(id, 0) > %d", req.Last)
	sync_to_client(pool, "titan", manu, tbln, whr, dbm, strm)
	return nil
}
func (s *titanServer) GetESP1Pharms(req *SyncReq, strm grpc.ServerStreamingServer[ESP1PharmNDC]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	pool := titan.pools["citus"]
	tbln := "public.esp1_providers"
	dbm  := new_dbmap[ESP1PharmNDC]()
	dbm.column("spid", "service_provider_id",	"service_provider_id")
	dbm.column("ndc",  "ndc",  					"ndc")
	dbm.column("strt", "start", 				"COALESCE(TRUNC(EXTRACT(EPOCH FROM start::timestamp)*1000000, 0), 0)")
	dbm.column("term", "term", 					"COALESCE(TRUNC(EXTRACT(EPOCH FROM term::timestamp) *1000000, 0), 0)")
	dbm.table(pool, tbln)

	manu := metaManu(strm.Context())
	whr := fmt.Sprintf("manufacturer = '%s' AND COALESCE(TRUNC(EXTRACT(EPOCH FROM updated_at)*1000000, 0), 0) > %d", manu, req.Last)
	sync_to_client(pool, "titan", manu, tbln, whr, dbm, strm)
	return nil
}
func (s *titanServer) GetEligibilityLedger(req *SyncReq, strm grpc.ServerStreamingServer[Eligibility]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	pool := titan.pools["citus"]
	tbln := "public.eligibility_ledger"
	dbm  := new_dbmap[Eligibility]()
	dbm.column("id",   "id",   					"id")
	dbm.column("i340", "id_340b", 				"id_340b")
	dbm.column("phid", "pharmacy_id", 			"pharmacy_id")
	dbm.column("manu", "manufacturer", 			"manufacturer")
	dbm.column("netw", "network", 				"network")
	dbm.column("strt", "start_at", 				"COALESCE(TRUNC(EXTRACT(EPOCH FROM start_at)*1000000, 0), 0)")
	dbm.column("term", "end_at", 				"COALESCE(TRUNC(EXTRACT(EPOCH FROM end_at)  *1000000, 0), 0)")
	dbm.table(pool, tbln)

	manu := metaManu(strm.Context())
	whr := fmt.Sprintf("manufacturer = '%s' AND COALESCE(id, 0) > %d", manu, req.Last)
	sync_to_client(pool, "titan", manu, tbln, whr, dbm, strm)
	return nil
}
func (s *titanServer) GetAuths(req *SyncReq, strm grpc.ServerStreamingServer[Auth]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	pool := titan.pools["titan"]
	tbln := "titan.auth"
	dbm  := new_dbmap[Auth]()
	dbm.table(pool, tbln)
	manu := metaManu(strm.Context())
	whr := fmt.Sprintf("manu = '%s'", manu)
	sync_to_client(pool, "titan", manu, tbln, whr, dbm, strm)
	return nil
}

func (s *titanServer) Rebates(strm grpc.ClientStreamingServer[TitanRebate, Res]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	sync_fm_client(titan.pools["titan"], "titan", manu, "titan.rebates", strm)
	return nil
}
func (s *titanServer) Scrubs(strm grpc.ClientStreamingServer[Scrub, Res]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	sync_fm_client(titan.pools["titan"], "titan", manu, "titan.scrubs", strm)
	return nil
}
func (s *titanServer) ClaimsUsed(strm grpc.ClientStreamingServer[ClaimUse, Res]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	sync_fm_client(titan.pools["titan"], "titan", manu, "titan.claim_uses", strm)
	return nil
}
func (s *titanServer) RebateClaims(strm grpc.ClientStreamingServer[RebateClaim, Res]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	sync_fm_client(titan.pools["titan"], "titan", manu, "titan.rebate_claims", strm)
	return nil
}

