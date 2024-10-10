package main

import (
	context "context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type svcServer struct {
    UnimplementedBinaryV5SvcServer
}

func run_grpc_services(wg *sync.WaitGroup, stop chan any) {
    defer wg.Done()

    cfg := &tls.Config{
        Certificates: []tls.Certificate{TLSCert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    X509pool,
    }
    for {
        select {
        case <-time.After(time.Duration(5) * time.Second):
            if service.gsr == nil {
                cred := credentials.NewTLS(cfg)
                service.gsr = grpc.NewServer(grpc.Creds(cred))

                go func() {
                    if lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port)); err == nil {
                        service.srv = &svcServer{}
                        RegisterBinaryV5SvcServer(service.gsr, service.srv)
                        log("service", "grpc", "server starting", 0, nil)
                        if err := service.gsr.Serve(lis); err != nil {
                            service.gsr = nil
                        }
                    } else {
                        service.gsr = nil
                    }
                }()
            }
        case <-stop:
            if service.gsr != nil {
                service.gsr.GracefulStop()
                return
            }
        }
    }
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
    return db_strm_select(strm, service.pools["esp"], "ncpdp_providers", cols, "")
}
func (s *svcServer) GetNDCs(req *Req, strm grpc.ServerStreamingServer[NDC]) error {
    cols := map[string]string{
        "item":         "ndc",
        "product_name": "name",
        "network":      "netw",
    }
    return db_strm_select(strm, service.pools["esp"], "ndcs", cols, fmt.Sprintf("manufacturer_name = '%s'", req.Manu))
}
func (s *svcServer) GetEntities(req *Req, strm grpc.ServerStreamingServer[Entity]) error {
    cols := map[string]string{
        "id_340b":                  "i340",
        "state":                    "state",
        "participating_start_date": "strt",
        "term_date":                "term",
    }
    return db_strm_select(strm, service.pools["esp"], "covered_entities", cols, "")
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
    return db_strm_select(strm, service.pools["esp"], "contracted_pharmacies", cols, "")
}
func (s *svcServer) GetESP1Pharms(req *Req, strm grpc.ServerStreamingServer[ESP1PharmNDC]) error {
    cols := map[string]string{
        "service_provider_id":  "spid",
        "ndc":                  "ndc",
        "start":                "strt",
        "term":                 "term",
    }
    return db_strm_select(strm, service.pools["citus"], "esp1_providers", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
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
    return db_strm_select(strm, service.pools["citus"], "submission_rows", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
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
    return db_strm_select(strm, service.pools["citus"], "eligibility_ledger", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
}
func (s *svcServer) Start(ctx context.Context, req *StartReq) (*StartRes, error) {
    return &StartRes{ScrubId: 0}, nil
}
func (s *svcServer) AddRebates(strm grpc.ClientStreamingServer[RebateRec, Res]) error {
    toSlice := func(obj any) []string {
        rbt  := obj.(*RebateRec)
        vals := make([]string, 4)
        vals[0] = rbt.Scid
        vals[1] = rbt.Rnum
        vals[2] = rbt.Status
        vals[3] = rbt.Fprt
        return vals
    }
    return scrubStreamExec[RebateRec](strm, "binary", "rbtbin.rebates", "insert", []string{"scrub_id", "rnum", "status", "fprt"}, toSlice, nil, nil)
}
func (s *svcServer) AddClaims(strm grpc.ClientStreamingServer[ClaimRec, Res]) error {
    toSlice := func(obj any) []string {
        clm  := obj.(*ClaimRec)
        vals := make([]string, 4)
        vals[0] = clm.Scid
        vals[1] = clm.Clid
        vals[2] = clm.Excl
        return vals
    }
    return scrubStreamExec[ClaimRec](strm, "binary", "rbtbin.claims", "insert", []string{"scrub_id", "clm_guid", "exclude"}, toSlice, nil, nil)
}
func (s *svcServer) UpdateClaims(strm grpc.ClientStreamingServer[ClaimUpdate, Res]) error {
    values := func(obj any) map[string]string {
        cu   := obj.(*ClaimUpdate)
        vals := map[string]string{}
        vals["rbt_rrid"]       = cu.Lscid
        vals["rbt_hdos"]       = cu.Lhdos
        vals["rbt_hdos_owner"] = cu.Lownr
        vals["rbt_hdos_auth"]  = cu.Lauth
        return vals
    }
    where := func(obj any) map[string]string {
        cu  := obj.(*ClaimUpdate)
        whr := map[string]string{}
        whr["id"] = cu.Clid
        return whr
    }
    return scrubStreamExec[ClaimUpdate](strm, "citus", "public.submission_rows", "update", nil, nil, values, where)
}
func (s *svcServer) Done(ctx context.Context, m *Metrics) (*Res, error) {
    values := func(obj any) map[string]string {
        met  := obj.(*Metrics)
        vals := map[string]string{}
        vals[""] = fmt.Sprintf("%d", met.ClmInvalid)
        vals[""] = fmt.Sprintf("%d", met.ClmMatched)
        vals[""] = fmt.Sprintf("%d", met.ClmNomatch)
        vals[""] = fmt.Sprintf("%d", met.ClmTotal)
        vals[""] = fmt.Sprintf("%d", met.ClmValid)
        vals[""] = fmt.Sprintf("%d", met.DosAftDoc)
        vals[""] = fmt.Sprintf("%d", met.DosAftDof)
        vals[""] = fmt.Sprintf("%d", met.DosBefDoc)
        vals[""] = fmt.Sprintf("%d", met.DosBefDof)
        vals[""] = fmt.Sprintf("%d", met.DosEquDoc)
        vals[""] = fmt.Sprintf("%d", met.DosEquDof)
        vals[""] = fmt.Sprintf("%d", met.RbtFailed)
        vals[""] = fmt.Sprintf("%d", met.RbtInvalid)
        vals[""] = fmt.Sprintf("%d", met.RbtMatched)
        vals[""] = fmt.Sprintf("%d", met.RbtNomatch)
        vals[""] = fmt.Sprintf("%d", met.RbtPassed)
        vals[""] = fmt.Sprintf("%d", met.RbtTotal)
        vals[""] = fmt.Sprintf("%d", met.RbtValid)
        vals[""] = fmt.Sprintf("%d", met.SpiChain)
        vals[""] = fmt.Sprintf("%d", met.SpiCross)
        vals[""] = fmt.Sprintf("%d", met.SpiExact)
        vals[""] = fmt.Sprintf("%d", met.SpiStack)
        return vals
    }
    scid := ""
    root := "fm_scrub"
    manu := ""
    proc := ""
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        scid = meta(md, "scid")
        manu = meta(md, "manu")
        proc = meta(md, "proc")
    }
    where := func(obj any) map[string]string {
        whr := map[string]string{}
        whr["scrub_id"] = scid
        return whr
    }
    scrubExec(ctx, m, "binary", "scrubs", "update", nil, nil, values, where)

    sd := scrubDir(scid, root, manu, proc)
    sd.close("to_azure")
    return &Res{}, nil
}
func scrubStreamExec[Q any](strm grpc.ClientStreamingServer[Q, Res], pool, tbln, oper string, cols []string, vals func(any)[]string, upds, whr func(any)map[string]string) error {
    scid := ""
    root := "fm_scrub"
    manu := ""
    proc := ""
    if md, ok := metadata.FromIncomingContext(strm.Context()); ok {
        scid = meta(md, "scid")
        manu = meta(md, "manu")
        proc = meta(md, "proc")
    }
    sd := scrubDir(scid, root, manu, proc)
    if sf, err := sd.scrubFile(pool, oper, tbln, cols); err == nil {
        for {
            if msg, err := strm.Recv(); err == nil {
                if oper == "insert" {
                    sf.insert(vals(msg))
                } else {
                    sf.update(upds(msg), whr(msg))
                }
            } else if err == io.EOF {
                return nil
            } else {
                return err
            }
        }
    } else {
        return err
    }
}
func scrubExec[Q any](ctx context.Context, msg Q, pool, tbln, oper string, cols []string, vals func(any)[]string, upds, whr func(any)map[string]string) error {
    scid := ""
    root := "fm_scrub"
    manu := ""
    proc := ""
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        scid = meta(md, "scid")
        manu = meta(md, "manu")
        proc = meta(md, "proc")
    }
    sd := scrubDir(scid, root, manu, proc)
    if sf, err := sd.scrubFile(pool, oper, tbln, cols); err == nil {
        for {
            if oper == "insert" {
                sf.insert(vals(msg))
            } else {
                sf.update(upds(msg), whr(msg))
            }
        }
    } else {
        return err
    }
}
func meta(md metadata.MD, name string) string {
    vals := md.Get(name)
    if len(vals) > 0 {
        return vals[0]
    }
    return ""
}