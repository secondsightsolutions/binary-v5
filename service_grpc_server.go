package main

import (
	context "context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
    svcCitusPool   *pgxpool.Pool
    svcCentralPool *pgxpool.Pool
)
type svcServer struct {
    UnimplementedBinaryV5SvcServer
}

func run_grpc_service() chan error {
    cfg := &tls.Config{
        Certificates: []tls.Certificate{TLSCert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    X509pool,
    }
    chn  := make(chan error)
    cred := credentials.NewTLS(cfg)
    service.gsr = grpc.NewServer(grpc.Creds(cred))

    go func() {
        if lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port)); err == nil {
            service.srv = &svcServer{}
            RegisterBinaryV5SvcServer(service.gsr, service.srv)
            fmt.Println("service: grpc server starting")
            if err := service.gsr.Serve(lis); err != nil {
                chn <-err
            }
            close(chn)
        } else {
            chn <-err
            close(chn)
        }
    }()
    return chn
}

func (s *svcServer) Ping(context.Context, *Req) (*Res, error) {
    return &Res{}, nil
}
func (s *svcServer) GetSPIs(req *Req, strm grpc.ServerStreamingServer[SPI]) error {
    cols := map[string]string{
        "ncpdp_provider_id":        "ncp",
        "national_provider_id":     "npi",
        "dea_registration_id":      "dea",
        "store_number":             "sto",
        "legal_business_name":      "lbn",
        "status_code_340b":         "cde",
        "chain_name":               "chn",
    }
    return db_read(strm, svcCentralPool, "ncpdp_providers", cols, "")
}
func (s *svcServer) GetNDCs(req *Req, strm grpc.ServerStreamingServer[NDC]) error {
    cols := map[string]string{
        "item":         "ndc",
        "product_name": "name",
        "network":      "netw",
    }
    return db_read(strm, svcCentralPool, "ndcs", cols, fmt.Sprintf("manufacturer_name = '%s'", req.Manu))
}
func (s *svcServer) GetEntities(req *Req, strm grpc.ServerStreamingServer[Entity]) error {
    cols := map[string]string{
        "id_340b":                  "i340",
        "state":                    "state",
        "participating_start_date": "strt",
        "term_date":                "term",
    }
    return db_read(strm, svcCentralPool, "covered_entities", cols, "")
}
func (s *svcServer) GetPharmacies(req *Req, strm grpc.ServerStreamingServer[Pharmacy]) error {
    cols := map[string]string{
        "chain_name":           "chnm",
        "id_340b":              "i340",
        "pharmacy_id":          "phid",
        "dea_id":               "dea",
        "national_provider_id": "npi",
        "ncpdp_provider_id":    "ncp",
        "dea":                  "deas",
        "npi":                  "npis",
        "ncpdp":                "ncps",
        "pharmacy_state":       "state",
    }
    return db_read(strm, svcCentralPool, "contracted_pharmacies", cols, "")
}
func (s *svcServer) GetESP1Pharms(req *Req, strm grpc.ServerStreamingServer[ESP1PharmNDC]) error {
    cols := map[string]string{
        "service_provider_id":  "spid",
        "ndc":                  "ndc",
        "start":                "strt",
        "term":                 "term",
    }
    return db_read(strm, svcCitusPool, "esp1_providers", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
}
func (s *svcServer) GetClaims(req *Req, strm grpc.ServerStreamingServer[Claim]) error {
    cols := map[string]string{
        "id":                       "clid",
        "chain_name":               "chnm",
        "claim_conforms_flag":      "cnfm",
        "created_at":               "doc",
        "formatted_dop":            "dop",
        "formatted_dos":            "dos",
        "date_prescribed":          "hdop",
        "date_of_service":          "dos",
        "formatted_rx_number":      "hfrx",
        "rx_number":                "hrxn",
        "contracted_entity_id":     "i340",
        "rbt_hdos_auth":            "lauth",
        "rbt_hdos_owner":           "lownr",
        "rbt_rrid":                 "lscid",
        "manufacturer":             "manu",
        "ndc":                      "ndc",
        "network":                  "netw",
        "product_name":             "prnm",
        "quantity":                 "qty",
        "short_id":                 "shid",
        "service_provider_id":      "spid",
        "prescriber_id":            "prid",
        "eligible_at_submission":   "elig",
        "suspended_at_submission":  "susp",
        "in_house_pharmacy_ids":    "ihph",
    }
    return db_read(strm, svcCitusPool, "submission_rows", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
}
func (s *svcServer) GetEligibilityLedger(req *Req, strm grpc.ServerStreamingServer[Eligibility]) error {
    cols := map[string]string{
        "elid":         "id",
        "id_340b":      "i340",
        "pharmacy_id":  "phid",
        "manufacturer": "manu",
        "network":      "netw",
        "start_at":     "strt",
        "end_at":       "term",
    }
    return db_read(strm, svcCitusPool, "eligibility_ledger", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
}
func (s *svcServer) Start(ctx context.Context, req *StartReq) (*StartRes, error) {
    return &StartRes{ScrubId: 0}, nil
}
func (s *svcServer) AddRebates(strm grpc.ClientStreamingServer[RebateRec, Res]) error {
    toSlice := func(obj any) []any {
        rbt := obj.(*RebateRec)
        vals := make([]any, 4)
        vals[0] = rbt.ScrubId
        vals[1] = rbt.Fprt
        vals[2] = rbt.Rnum
        vals[3] = rbt.Status
        return vals
    }
    return db_insert(strm, svcCitusPool, "rbtbin.rebates", []string{"scrub_id", "fprt", "rnum", "status"}, toSlice, 10000)
}
func (s *svcServer) AddClaims(strm grpc.ClientStreamingServer[ClaimRec, Res]) error {
    return nil
}
func (s *svcServer) UpdateClaims(strm grpc.ClientStreamingServer[ClaimUpdate, Res]) error {
    return nil
}
func (s *svcServer) Done(ctx context.Context, m *Metrics) (*Res, error) {
    return nil, nil
}