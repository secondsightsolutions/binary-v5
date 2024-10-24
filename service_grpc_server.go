package main

import (
	context "context"
	"fmt"
	"io"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type binaryV5SvcServer struct {
    UnimplementedBinaryV5SvcServer
}

func (s *binaryV5SvcServer) Ping(context.Context, *Req) (*Res, error) {
    return &Res{}, nil
}
func (s *binaryV5SvcServer) GetSPIs(req *Req, strm grpc.ServerStreamingServer[SPI]) error {
    cols := map[string]string{
        "ncp":  "COALESCE(ncpdp_provider_id, '')",
        "npi":  "COALESCE(national_provider_id, '')",
        "dea":  "COALESCE(dea_registration_id, '')",
        "sto":  "COALESCE(store_number, '')",
        "lbn":  "COALESCE(legal_business_name, '')",
        "cde":  "COALESCE(status_code_340b, '')",
        "chn":  "COALESCE(chain_name, '')",
    }
    return db_strm_select(strm, service.pools["esp"], "ncpdp_providers", cols, "")
}
func (s *binaryV5SvcServer) GetNDCs(req *Req, strm grpc.ServerStreamingServer[NDC]) error {
    cols := map[string]string{
        "ndc":  "COALESCE(item, '')",
        "name": "COALESCE(product_name, '')",
        "netw": "COALESCE(network, '')",
    }
    return db_strm_select(strm, service.pools["esp"], "ndcs", cols, fmt.Sprintf("manufacturer_name = '%s'", req.Manu))
}
func (s *binaryV5SvcServer) GetEntities(req *Req, strm grpc.ServerStreamingServer[Entity]) error {
    cols := map[string]string{
        "i340": "COALESCE(id_340b, '')",
        "state":"COALESCE(state, '')",
        "strt": "COALESCE(TRUNC(EXTRACT(EPOCH FROM participating_start_date::timestamp) *1000000, 0), 0)",
        "term": "COALESCE(TRUNC(EXTRACT(EPOCH FROM term_date::timestamp)                *1000000, 0), 0)",
    }
    return db_strm_select(strm, service.pools["esp"], "covered_entities", cols, "")
}
func (s *binaryV5SvcServer) GetPharmacies(req *Req, strm grpc.ServerStreamingServer[Pharmacy]) error {
    cols := map[string]string{
        "chnm": "COALESCE(chain_name, '')",
        "i340": "COALESCE(id_340b, '')",
        "phid": "COALESCE(pharmacy_id, '')",
        "dea":  "COALESCE(dea_id, '')",
        "npi":  "COALESCE(national_provider_id, '')",
        "ncp":  "COALESCE(ncpdp_provider_id, '')",
        "deas": "COALESCE(dea, '{}')",
        "npis": "COALESCE(npi, '{}')",
        "ncps": "COALESCE(ncpdp, '{}')",
        "state":"COALESCE(pharmacy_state, '')",
    }
    return db_strm_select(strm, service.pools["esp"], "contracted_pharmacies", cols, "")
}
func (s *binaryV5SvcServer) GetESP1Pharms(req *Req, strm grpc.ServerStreamingServer[ESP1PharmNDC]) error {
    cols := map[string]string{
        "spid": "service_provider_id",
        "ndc":  "ndc",
        "strt": "COALESCE(TRUNC(EXTRACT(EPOCH FROM start::timestamp)*1000000, 0), 0)",
        "term": "COALESCE(TRUNC(EXTRACT(EPOCH FROM term::timestamp) *1000000, 0), 0)",
    }
    return db_strm_select(strm, service.pools["citus"], "esp1_providers", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
}
func (s *binaryV5SvcServer) GetClaims(req *Req, strm grpc.ServerStreamingServer[Claim]) error {
    cols := map[string]string{
        "clid": "COALESCE(id, '')",
        "chnm": "COALESCE(chain_name, '')",
        "cnfm": "COALESCE(claim_conforms_flag, true)",
        "doc":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)   *1000000, 0), 0)",
        "dop":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM formatted_dop)*1000000, 0), 0)",
        "dos":  "COALESCE(TRUNC(EXTRACT(EPOCH FROM formatted_dos)*1000000, 0), 0)",
        "hdop": "COALESCE(date_prescribed, '')",
        "hdos": "COALESCE(date_of_service, '')",
        "hfrx": "COALESCE(formatted_rx_number, '')",
        "hrxn": "COALESCE(rx_number, '')",
        "i340": "COALESCE(contracted_entity_id, '')",
        "lauth":"COALESCE(rbt_hdos_auth, '')",
        "lownr":"COALESCE(rbt_hdos_owner, '')",
        "lscid":"COALESCE(rbt_rrid, -1)",
        "manu": "COALESCE(manufacturer, '')",
        "ndc":  "COALESCE(ndc, '')",
        "netw": "COALESCE(network, '')",
        "prnm": "COALESCE(product_name, '')",
        "qty":  "COALESCE(quantity, 0)",
        "shid": "COALESCE(short_id, '')",
        "spid": "COALESCE(service_provider_id, '')",
        "prid": "COALESCE(prescriber_id, '')",
        "elig": "COALESCE(eligible_at_submission, true)",
        "susp": "COALESCE(suspended_submission, false)",
        "ihph": "COALESCE(in_house_pharmacy_ids, '{}')",
    }
    return db_strm_select(strm, service.pools["citus"], "submission_rows", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
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
    return db_strm_select(strm, service.pools["citus"], "eligibility_ledger", cols, fmt.Sprintf("manufacturer = '%s'", req.Manu))
}
func (s *binaryV5SvcServer) Start(ctx context.Context, req *StartReq) (*StartRes, error) {
    return &StartRes{ScrubId: 0}, nil
}
func (s *binaryV5SvcServer) AddRebates(strm grpc.ClientStreamingServer[RebateRec, Res]) error {
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
func (s *binaryV5SvcServer) AddClaims(strm grpc.ClientStreamingServer[ClaimRec, Res]) error {
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
func (s *binaryV5SvcServer) UpdateClaims(strm grpc.ClientStreamingServer[ClaimUpdate, Res]) error {
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
func (s *binaryV5SvcServer) Done(ctx context.Context, m *Metrics) (*Res, error) {
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