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
func (s *titanServer) GetSPIs(req *Req, strm grpc.ServerStreamingServer[SPI]) error {
	cols := map[string]string{
		"ncp": "COALESCE(ncpdp_provider_id, '')",
		"npi": "COALESCE(national_provider_id, '')",
		"dea": "COALESCE(dea_registration_id, '')",
		"sto": "COALESCE(store_number, '')",
		"lbn": "COALESCE(legal_business_name, '')",
		"cde": "COALESCE(status_code_340b, '')",
		"chn": "COALESCE(chain_name, '')",
	}
	return db_strm_select(strm, titan.pools["esp"], "ncpdp_providers", cols, "")
}
func (s *titanServer) GetNDCs(req *Req, strm grpc.ServerStreamingServer[NDC]) error {
	cols := map[string]string{
		"ndc":  "COALESCE(REPLACE(item, '-', ''), '')",
		"name": "COALESCE(product_name, '')",
		"netw": "COALESCE(network, '')",
	}
	return db_strm_select(strm, titan.pools["esp"], "ndcs", cols, fmt.Sprintf("manufacturer_name = '%s'", req.Manu))
}
func (s *titanServer) GetEntities(req *Req, strm grpc.ServerStreamingServer[Entity]) error {
	cols := map[string]string{
		"i340":  "COALESCE(id_340b, '')",
		"state": "COALESCE(state, '')",
		"strt":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM participating_start_date::timestamp) *1000000, 0), 0)",
		"term":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM term_date::timestamp)                *1000000, 0), 0)",
	}
	return db_strm_select(strm, titan.pools["esp"], "covered_entities", cols, "")
}
func (s *titanServer) GetPharmacies(req *Req, strm grpc.ServerStreamingServer[Pharmacy]) error {
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
	return db_strm_select(strm, titan.pools["esp"], "contracted_pharmacies", cols, "")
}
func (s *titanServer) GetESP1Pharms(req *Req, strm grpc.ServerStreamingServer[ESP1PharmNDC]) error {
	cols := map[string]string{
		"spid": "service_provider_id",
		"ndc":  "ndc",
		"strt": "COALESCE(TRUNC(EXTRACT(EPOCH FROM start::timestamp)*1000000, 0), 0)",
		"term": "COALESCE(TRUNC(EXTRACT(EPOCH FROM term::timestamp) *1000000, 0), 0)",
	}
	return db_strm_select(strm, titan.pools["citus"], "esp1_providers", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
}
func (s *titanServer) GetEligibilityLedger(req *Req, strm grpc.ServerStreamingServer[Eligibility]) error {
	cols := map[string]string{
		"id":   "id",
		"i340": "id_340b",
		"phid": "pharmacy_id",
		"manu": "manufacturer",
		"netw": "network",
		"strt": "COALESCE(TRUNC(EXTRACT(EPOCH FROM start_at)*1000000, 0), 0)",
		"term": "COALESCE(TRUNC(EXTRACT(EPOCH FROM end_at)  *1000000, 0), 0)",
	}
	return db_strm_select(strm, titan.pools["citus"], "eligibility_ledger", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
}

// set locks

func (s *titanServer) NewScrub(ctx context.Context, scr *Scrub) (*Res, error) {
	_, err := db_insert_one(ctx, titan.pools["titan"], "titan.scrubs", nil, scr, "")
	return &Res{}, err
}
func (s *titanServer) Rebates(strm grpc.ClientStreamingServer[TitanRebate, Res]) error {
	return db_insert_strm_fm_client(strm, titan.pools["titan"], "titan.rebates", nil, 10000)
}
func (s *titanServer) ClaimsUsed(strm grpc.ClientStreamingServer[ClaimUse, Res]) error {
	return db_insert_strm_fm_client(strm, titan.pools["titan"], "titan.claim_uses", nil, 10000)
}
func (s *titanServer) RebateClaims(strm grpc.ClientStreamingServer[RebateClaim, Res]) error {
	return db_insert_strm_fm_client(strm, titan.pools["titan"], "titan.rebate_claims", nil, 10000)
}
func (s *titanServer) ScrubDone(ctx context.Context, m *Metrics) (*Res, error) {
	cols := map[string]string{
		"rbt_total":	"GetRbtTotal",
		"rbt_valid":  	"GetRbtValid",
		"rbt_matched":	"GetRbtMatched",
		"rbt_nomatch":	"GetRbtNomatch",
		"rbt_passed":	"GetRbtPassed",
		"rbt_failed":	"GetRbtFailed",
		"clm_total":	"GetClmTotal",
		"clm_valid":	"GetClmValid",
		"clm_matched":	"GetClmMatched",
		"clm_nomatch":	"GetClmNomatch",
		"clm_invalid":	"GetClmInvalid",
		"spi_exact":	"GetSpiExact",
		"spi_cross":	"GetSpiCross",
		"spi_stack":	"GetSpiStack",
		"spi_chain":	"GetSpiChain",
		"dos_equ_doc":	"GetDosEquDoc",
		"dos_bef_doc":	"GetDosBefDoc",
		"dos_equ_dof":	"GetDosEquDof",
		"dos_bef_dof":	"GetDosBefDof",
		"dos_aft_dof":	"GetDosAftDof",
	}
	_, _, _, manu, scid := getMetaGRPC(ctx)
	whr := map[string]string{
		"manu": manu,
		"scid": scid,
	}
	return &Res{}, db_update(ctx, m, titan.pools["titan"], "titan.scrubs", cols, whr)
}
func (s *titanServer) SyncClaims(req *SyncReq, strm grpc.ServerStreamingServer[Claim]) error {
	cols := map[string]string{
		// "clid":  "COALESCE(id, '')",
		"chnm":  "COALESCE(chain_name, '')",
		"cnfm":  "COALESCE(claim_conforms_flag, true)",
		"doc":   "COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)   *1000000, 0), 0)",
		"dop":   "COALESCE(TRUNC(EXTRACT(EPOCH FROM formatted_dop)*1000000, 0), 0)",
		"dos":   "COALESCE(TRUNC(EXTRACT(EPOCH FROM formatted_dos)*1000000, 0), 0)",
		"hdop":  "COALESCE(date_prescribed, '')",
		"hdos":  "COALESCE(date_of_service, '')",
		"hfrx":  "COALESCE(formatted_rx_number, '')",
		"hrxn":  "COALESCE(rx_number, '')",
		"i340":  "SPLIT_PART(COALESCE(id_340b, ''), '-', 1)",
		// "lauth": "COALESCE(rbt_hdos_auth, '')",
		// "lownr": "COALESCE(rbt_hdos_owner, '')",
		// "lscid": "COALESCE(rbt_rrid, -1)",
		"manu":  "COALESCE(manufacturer, '')",
		"ndc":   "REPLACE(COALESCE(ndc, ''), '-', '')",
		"netw":  "COALESCE(network, '')",
		"prnm":  "COALESCE(product_name, '')",
		"qty":   "COALESCE(quantity, 0)",
		"shrt":  "COALESCE(short_id, '')",
		"spid":  "COALESCE(service_provider_id, '')",
		"prid":  "COALESCE(prescriber_id, '')",
		"elig":  "COALESCE(eligible_at_submission, true)",
		"susp":  "COALESCE(suspended_submission, false)",
		"ihph":  "COALESCE(in_house_pharmacy_ids, '{}')",
	}
	whr := fmt.Sprintf("manufacturer = '%s' AND COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)*1000000, 0), 0) >= %d", req.Manu, req.Last)
	return db_strm_select(strm, titan.pools["citus"], "submission_rows", cols, whr)
}

func titanValidate(ctx context.Context) error {
	return nil
}