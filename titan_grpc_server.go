package main

import (
	context "context"
	"fmt"

	grpc "google.golang.org/grpc"
)

type binaryV5SvcServer struct {
	UnimplementedBinaryV5SvcServer
}

func (s *binaryV5SvcServer) Ping(context.Context, *Req) (*Res, error) {
	return &Res{}, nil
}
func (s *binaryV5SvcServer) GetSPIs(req *Req, strm grpc.ServerStreamingServer[SPI]) error {
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
func (s *binaryV5SvcServer) GetNDCs(req *Req, strm grpc.ServerStreamingServer[NDC]) error {
	cols := map[string]string{
		"ndc":  "COALESCE(item, '')",
		"name": "COALESCE(product_name, '')",
		"netw": "COALESCE(network, '')",
	}
	return db_strm_select(strm, titan.pools["esp"], "ndcs", cols, fmt.Sprintf("manufacturer_name = '%s'", req.Manu))
}
func (s *binaryV5SvcServer) GetEntities(req *Req, strm grpc.ServerStreamingServer[Entity]) error {
	cols := map[string]string{
		"i340":  "COALESCE(id_340b, '')",
		"state": "COALESCE(state, '')",
		"strt":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM participating_start_date::timestamp) *1000000, 0), 0)",
		"term":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM term_date::timestamp)                *1000000, 0), 0)",
	}
	return db_strm_select(strm, titan.pools["esp"], "covered_entities", cols, "")
}
func (s *binaryV5SvcServer) GetPharmacies(req *Req, strm grpc.ServerStreamingServer[Pharmacy]) error {
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
func (s *binaryV5SvcServer) GetESP1Pharms(req *Req, strm grpc.ServerStreamingServer[ESP1PharmNDC]) error {
	cols := map[string]string{
		"spid": "service_provider_id",
		"ndc":  "ndc",
		"strt": "COALESCE(TRUNC(EXTRACT(EPOCH FROM start::timestamp)*1000000, 0), 0)",
		"term": "COALESCE(TRUNC(EXTRACT(EPOCH FROM term::timestamp) *1000000, 0), 0)",
	}
	return db_strm_select(strm, titan.pools["citus"], "esp1_providers", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
}
func (s *binaryV5SvcServer) GetClaims(req *Req, strm grpc.ServerStreamingServer[Claim]) error {
	cols := map[string]string{
		"clid":  "COALESCE(id, '')",
		"chnm":  "COALESCE(chain_name, '')",
		"cnfm":  "COALESCE(claim_conforms_flag, true)",
		"doc":   "COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)   *1000000, 0), 0)",
		"dop":   "COALESCE(TRUNC(EXTRACT(EPOCH FROM formatted_dop)*1000000, 0), 0)",
		"dos":   "COALESCE(TRUNC(EXTRACT(EPOCH FROM formatted_dos)*1000000, 0), 0)",
		"hdop":  "COALESCE(date_prescribed, '')",
		"hdos":  "COALESCE(date_of_service, '')",
		"hfrx":  "COALESCE(formatted_rx_number, '')",
		"hrxn":  "COALESCE(rx_number, '')",
		"i340":  "COALESCE(contracted_entity_id, '')",
		"lauth": "COALESCE(rbt_hdos_auth, '')",
		"lownr": "COALESCE(rbt_hdos_owner, '')",
		"lscid": "COALESCE(rbt_rrid, -1)",
		"manu":  "COALESCE(manufacturer, '')",
		"ndc":   "COALESCE(ndc, '')",
		"netw":  "COALESCE(network, '')",
		"prnm":  "COALESCE(product_name, '')",
		"qty":   "COALESCE(quantity, 0)",
		"shid":  "COALESCE(short_id, '')",
		"spid":  "COALESCE(service_provider_id, '')",
		"prid":  "COALESCE(prescriber_id, '')",
		"elig":  "COALESCE(eligible_at_submission, true)",
		"susp":  "COALESCE(suspended_submission, false)",
		"ihph":  "COALESCE(in_house_pharmacy_ids, '{}')",
	}
	return db_strm_select(strm, titan.pools["citus"], "submission_rows", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
}
func (s *binaryV5SvcServer) GetEligibilityLedger(req *Req, strm grpc.ServerStreamingServer[Eligibility]) error {
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
func (s *binaryV5SvcServer) Start(ctx context.Context, req *StartReq) (*StartRes, error) {
	return &StartRes{Scid: 0}, nil
}
func (s *binaryV5SvcServer) AddRebates(strm grpc.ClientStreamingServer[RebateRec, Res]) error {
	cols := map[string]string{
		"manu": "getManu",
		"scid": "getScid",
		"shrt": "getShrt",
		"excl": "getExcl",
	}
	return db_strm_insert(strm, titan.pools["titan"], "titan.rebates", cols, 10000)
}
func (s *binaryV5SvcServer) AddClaims(strm grpc.ClientStreamingServer[ClaimRec, Res]) error {
	cols := map[string]string{
		"manu": "getManu",
		"scid": "getScid",
		"rbid": "getRbid",
		"stat": "getStat",
		"errc": "getErrc",
		"errm": "getErrm",
	}
	return db_strm_insert(strm, titan.pools["titan"], "titan.claims", cols, 10000)
}
// set locks

func (s *binaryV5SvcServer) Done(ctx context.Context, m *Metrics) (*Res, error) {
	
	return &Res{}, nil
}

