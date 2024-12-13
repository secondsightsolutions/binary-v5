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
	f2c := map[string]string{
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
		"ihph": "array_to_string(in_house_pharmacy_ids, ',')",
	}
	manu := metaManu(strm.Context())
	whr  := fmt.Sprintf("manufacturer = '%s' AND COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)*1000000, 0), 0) > %d", manu, req.Last)
	sync_to_client(titan.pools["citus"], "titan", manu, "public.submission_rows", whr, f2c, strm)
	return nil
}
func (s *titanServer) GetSPIs(req *SyncReq, strm grpc.ServerStreamingServer[SPI]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	f2c := map[string]string{
		"ncp": "COALESCE(ncpdp_provider_id, '')",
		"npi": "COALESCE(national_provider_id, '')",
		"dea": "COALESCE(dea_registration_id, '')",
		"sto": "COALESCE(store_number, '')",
		"lbn": "COALESCE(legal_business_name, '')",
		"cde": "COALESCE(status_code_340b, '')",
		"chn": "COALESCE(chain_name, '')",
		"nam": "COALESCE(name, '')",
	}
	whr := fmt.Sprintf("COALESCE(id, 0) > %d", req.Last)
	sync_to_client(titan.pools["esp"], "titan", manu, "public.ncpdp_providers", whr, f2c, strm)
	return nil
}
func (s *titanServer) GetNDCs(req *SyncReq, strm grpc.ServerStreamingServer[NDC]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	f2c := map[string]string{
		"ndc":  "COALESCE(REPLACE(item, '-', ''), '')",
		"name": "COALESCE(product_name, '')",
		"netw": "COALESCE(network, '')",
	}
	manu := metaManu(strm.Context())
	whr := fmt.Sprintf("manufacturer_name = '%s' AND COALESCE(id, 0) > %d", manu, req.Last)
	sync_to_client(titan.pools["esp"], "titan", manu, "public.ndcs", whr, f2c, strm)
	return nil
}
func (s *titanServer) GetEntities(req *SyncReq, strm grpc.ServerStreamingServer[Entity]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	f2c := map[string]string{
		"i340":  "COALESCE(id_340b, '')",
		"state": "COALESCE(state, '')",
		"strt":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM participating_start_date::timestamp) *1000000, 0), 0)",
		"term":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM term_date::timestamp)                *1000000, 0), 0)",
	}
	whr := fmt.Sprintf("COALESCE(id, 0) > %d", req.Last)
	sync_to_client(titan.pools["esp"], "titan", manu, "public.covered_entities", whr, f2c, strm)
	return nil
}
func (s *titanServer) GetPharmacies(req *SyncReq, strm grpc.ServerStreamingServer[Pharmacy]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	f2c := map[string]string{
		"chnm":  "COALESCE(chain_name, '')",
		"i340":  "COALESCE(id_340b, '')",
		"phid":  "COALESCE(pharmacy_id, '')",
		"dea":   "COALESCE(dea_id, '')",
		"npi":   "COALESCE(national_provider_id, '')",
		"ncp":   "COALESCE(ncpdp_provider_id, '')",
		"deas":  "array_to_string(dea, ',')",
		"npis":  "array_to_string(npi, ',')",
		"ncps":  "array_to_string(ncpdp, ',')",
		"state": "COALESCE(pharmacy_state, '')",
	}
	whr := fmt.Sprintf("COALESCE(id, 0) > %d", req.Last)
	sync_to_client(titan.pools["esp"], "titan", manu, "public.contracted_pharmacies", whr, f2c, strm)
	return nil
}
func (s *titanServer) GetESP1Pharms(req *SyncReq, strm grpc.ServerStreamingServer[ESP1PharmNDC]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	f2c := map[string]string{
		"spid": "service_provider_id",
		"ndc":  "ndc",
		"strt": "COALESCE(TRUNC(EXTRACT(EPOCH FROM start::timestamp)*1000000, 0), 0)",
		"term": "COALESCE(TRUNC(EXTRACT(EPOCH FROM term::timestamp) *1000000, 0), 0)",
	}
	manu := metaManu(strm.Context())
	whr := fmt.Sprintf("manufacturer = '%s' AND COALESCE(TRUNC(EXTRACT(EPOCH FROM updated_at)*1000000, 0), 0) > %d", manu, req.Last)
	sync_to_client(titan.pools["citus"], "titan", manu, "public.esp1_providers", whr, f2c, strm)
	return nil
}
func (s *titanServer) GetEligibilityLedger(req *SyncReq, strm grpc.ServerStreamingServer[Eligibility]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	f2c := map[string]string{
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
	sync_to_client(titan.pools["citus"], "titan", manu, "public.eligibility_ledger", whr, f2c, strm)
	return nil
}
func (s *titanServer) GetAuths(req *SyncReq, strm grpc.ServerStreamingServer[Auth]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	manu := metaManu(strm.Context())
	whr := fmt.Sprintf("manu = '%s'", manu)
	sync_to_client(titan.pools["titan"], "titan", manu, "titan.auth", whr, nil, strm)
	return nil
}

func (s *titanServer) Rebates(strm grpc.ClientStreamingServer[TitanRebate, Res]) error {
	if err := validate_client(strm.Context(), titan.pools["titan"], "titan"); err != nil {
		return err
	}
	sync_fm_client(titan.pools["titan"], "titan", manu, "titan.rebates", strm)
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

