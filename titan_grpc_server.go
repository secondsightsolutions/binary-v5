package main

import (
	context "context"
	"fmt"

	grpc "google.golang.org/grpc"
)

type titanServer struct {
	UnimplementedTitanServer
}

func (s *titanServer) Ping(context.Context, *Req) (*Res, error) {
	return &Res{}, nil
}

func (s *titanServer) GetClaims(req *SyncReq, strm grpc.ServerStreamingServer[Claim]) error {
	cols := map[string]string{
		// "clid":  "COALESCE(id, '')",
		"chnm": "COALESCE(chain_name, '')",
		"cnfm": "COALESCE(claim_conforms_flag, true)",
		"seq":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)   *1000000, 0), 0)",
		"doc":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)   *1000000, 0), 0)",
		"dop":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM formatted_dop)*1000000, 0), 0)",
		"dos":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM formatted_dos)*1000000, 0), 0)",
		"hdop": "COALESCE(date_prescribed, '')",
		"hdos": "COALESCE(date_of_service, '')",
		"hfrx": "COALESCE(formatted_rx_number, '')",
		"hrxn": "COALESCE(rx_number, '')",
		"i340": "SPLIT_PART(COALESCE(id_340b, ''), '-', 1)",
		// "lauth": "COALESCE(rbt_hdos_auth, '')",
		// "lownr": "COALESCE(rbt_hdos_owner, '')",
		// "lscid": "COALESCE(rbt_rrid, -1)",
		"manu": "COALESCE(manufacturer, '')",
		"ndc":  "REPLACE(COALESCE(ndc, ''), '-', '')",
		"netw": "COALESCE(network, '')",
		"prnm": "COALESCE(product_name, '')",
		"qty":  "COALESCE(quantity, 0)",
		"shrt": "COALESCE(short_id, '')",
		"spid": "COALESCE(service_provider_id, '')",
		"prid": "COALESCE(prescriber_id, '')",
		"elig": "COALESCE(eligible_at_submission, true)",
		"susp": "COALESCE(suspended_submission, false)",
		"ihph": "COALESCE(in_house_pharmacy_ids, '{}')",
	}
	manu := metaManu(strm.Context())
	whr  := fmt.Sprintf("manufacturer = '%s' AND COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)*1000000, 0), 0) > %d", manu, req.Last)
	sync_to_client(titan.pools["citus"], "titan", manu, "submission_rows", whr, cols, strm)
	return nil
}
func (s *titanServer) GetSPIs(req *SyncReq, strm grpc.ServerStreamingServer[SPI]) error {
	cols := map[string]string{
		"ncp": "COALESCE(ncpdp_provider_id, '')",
		"npi": "COALESCE(national_provider_id, '')",
		"dea": "COALESCE(dea_registration_id, '')",
		"sto": "COALESCE(store_number, '')",
		"lbn": "COALESCE(legal_business_name, '')",
		"cde": "COALESCE(status_code_340b, '')",
		"chn": "COALESCE(chain_name, '')",
	}
	whr := fmt.Sprintf("COALESCE(id, 0) > %d", req.Last)
	sync_to_client(titan.pools["esp"], "titan", manu, "ncpdp_providers", whr, cols, strm)
	return nil
}
func (s *titanServer) GetNDCs(req *SyncReq, strm grpc.ServerStreamingServer[NDC]) error {
	cols := map[string]string{
		"ndc":  "COALESCE(REPLACE(item, '-', ''), '')",
		"name": "COALESCE(product_name, '')",
		"netw": "COALESCE(network, '')",
	}
	manu := metaManu(strm.Context())
	whr := fmt.Sprintf("manufacturer_name = '%s' AND COALESCE(id, 0) > %d", manu, req.Last)
	sync_to_client(titan.pools["esp"], "titan", manu, "ndcs", whr, cols, strm)
	return nil
}
func (s *titanServer) GetEntities(req *SyncReq, strm grpc.ServerStreamingServer[Entity]) error {
	cols := map[string]string{
		"i340":  "COALESCE(id_340b, '')",
		"state": "COALESCE(state, '')",
		"strt":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM participating_start_date::timestamp) *1000000, 0), 0)",
		"term":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM term_date::timestamp)                *1000000, 0), 0)",
	}
	whr := fmt.Sprintf("COALESCE(id, 0) > %d", req.Last)
	sync_to_client(titan.pools["esp"], "titan", manu, "covered_entities", whr, cols, strm)
	return nil
}
func (s *titanServer) GetPharmacies(req *SyncReq, strm grpc.ServerStreamingServer[Pharmacy]) error {
	cols := map[string]string{
		"chnm":  "COALESCE(chain_name, '')",
		"i340":  "COALESCE(id_340b, '')",
		"phid":  "COALESCE(pharmacy_id, '')",
		"dea":   "COALESCE(dea_id, '')",
		"npi":   "COALESCE(national_provider_id, '')",
		"ncp":   "COALESCE(ncpdp_provider_id, '')",
		"deas":  "COALESCE(dea, '{}')",
		"npis":  "COALESCE(npi, '{}')",
		"ncps":  "COALESCE(ncpdp, '{}')",
		"state": "COALESCE(pharmacy_state, '')",
	}
	whr := fmt.Sprintf("COALESCE(id, 0) > %d", req.Last)
	sync_to_client(titan.pools["esp"], "titan", manu, "contracted_pharmacies", whr, cols, strm)
	return nil
}
func (s *titanServer) GetESP1Pharms(req *SyncReq, strm grpc.ServerStreamingServer[ESP1PharmNDC]) error {
	cols := map[string]string{
		"spid": "service_provider_id",
		"ndc":  "ndc",
		"strt": "COALESCE(TRUNC(EXTRACT(EPOCH FROM start::timestamp)*1000000, 0), 0)",
		"term": "COALESCE(TRUNC(EXTRACT(EPOCH FROM term::timestamp) *1000000, 0), 0)",
	}
	manu := metaManu(strm.Context())
	whr := fmt.Sprintf("manufacturer = '%s' AND COALESCE(TRUNC(EXTRACT(EPOCH FROM updated_at)*1000000, 0), 0) > %d", manu, req.Last)
	sync_to_client(titan.pools["citus"], "titan", manu, "esp1_providers", whr, cols, strm)
	return nil
}
func (s *titanServer) GetEligibilityLedger(req *SyncReq, strm grpc.ServerStreamingServer[Eligibility]) error {
	cols := map[string]string{
		"id":   "id",
		"i340": "id_340b",
		"phid": "pharmacy_id",
		"manu": "manufacturer",
		"netw": "network",
		"strt": "COALESCE(TRUNC(EXTRACT(EPOCH FROM start_at)*1000000, 0), 0)",
		"term": "COALESCE(TRUNC(EXTRACT(EPOCH FROM end_at)  *1000000, 0), 0)",
	}
	manu := metaManu(strm.Context())
	whr := fmt.Sprintf("manufacturer = '%s' AND COALESCE(id, 0) > %d", manu, req.Last)
	sync_to_client(titan.pools["citus"], "titan", manu, "eligibility_ledger", whr, cols, strm)
	return nil
}

// set locks

func (s *titanServer) NewScrub(ctx context.Context, scr *Scrub) (*Res, error) {
	_, err := db_insert_one(ctx, titan.pools["titan"], "titan.scrubs", nil, scr, "")
	return &Res{}, err
}
func (s *titanServer) ScrubDone(ctx context.Context, m *Metrics) (*Res, error) {
	cols := map[string]string{
		"rbt_total":   "GetRbtTotal",
		"rbt_valid":   "GetRbtValid",
		"rbt_matched": "GetRbtMatched",
		"rbt_nomatch": "GetRbtNomatch",
		"rbt_passed":  "GetRbtPassed",
		"rbt_failed":  "GetRbtFailed",
		"clm_total":   "GetClmTotal",
		"clm_valid":   "GetClmValid",
		"clm_matched": "GetClmMatched",
		"clm_nomatch": "GetClmNomatch",
		"clm_invalid": "GetClmInvalid",
		"spi_exact":   "GetSpiExact",
		"spi_cross":   "GetSpiCross",
		"spi_stack":   "GetSpiStack",
		"spi_chain":   "GetSpiChain",
		"dos_equ_doc": "GetDosEquDoc",
		"dos_bef_doc": "GetDosBefDoc",
		"dos_equ_dof": "GetDosEquDof",
		"dos_bef_dof": "GetDosBefDof",
		"dos_aft_dof": "GetDosAftDof",
	}
	_, _, _, manu, scid := getMetaGRPC(ctx)
	whr := map[string]string{
		"manu": manu,
		"scid": scid,
	}
	return &Res{}, db_update(ctx, m, titan.pools["titan"], "titan.scrubs", cols, whr)
}

func (s *titanServer) Rebates(strm grpc.ClientStreamingServer[TitanRebate, Res]) error {
	sync_fm_client(titan.pools["titan"], "titan", manu, "titan.rebates", strm)
	return nil
}
func (s *titanServer) ClaimsUsed(strm grpc.ClientStreamingServer[ClaimUse, Res]) error {
	sync_fm_client(titan.pools["titan"], "titan", manu, "titan.claim_uses", strm)
	return nil
}
func (s *titanServer) RebateClaims(strm grpc.ClientStreamingServer[RebateClaim, Res]) error {
	sync_fm_client(titan.pools["titan"], "titan", manu, "titan.rebate_claims", strm)
	return nil
}

func titanValidate(ctx context.Context) error {
	return nil
}
