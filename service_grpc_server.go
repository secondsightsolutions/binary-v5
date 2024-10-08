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

func run_grpc_service() {
    cfg := &tls.Config{
        Certificates: []tls.Certificate{TLSCert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    X509pool,
    }
    cred := credentials.NewTLS(cfg)
    service.gsr = grpc.NewServer(grpc.Creds(cred))

    if lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port)); err == nil {
        service.srv = &svcServer{}
        RegisterBinaryV5SvcServer(service.gsr, service.srv)
        fmt.Println("service: grpc server starting")
        if err := service.gsr.Serve(lis); err != nil {
            fmt.Printf("cannot start GRPC server endpoint: %s\n", err.Error())
        }
        fmt.Println("service: grpc server exiting")
        return
    }
}

func (s *svcServer) Ping(context.Context, *Req) (*Res, error) {
    return &Res{}, nil
}
func (s *svcServer) GetSPIs(req *Req, strm grpc.ServerStreamingServer[SPI]) error {
    create := func(obj map[string]any) any {
        spi := &SPI{}
        spi.Cde = toStr(obj["status_code_340b"])
        spi.Chn = toStr(obj["chain_name"])
        spi.Dea = toStr(obj["dea_registration_id"])
        spi.Lbn = toStr(obj["legal_business_name"])
        spi.Ncp = toStr(obj["ncpdp_provider_id"])
        spi.Npi = toStr(obj["national_provider_id"])
        spi.Sto = toStr(obj["store_number"])
        return spi
    }
    qry := "SELECT ncpdp_provider_id, national_provider_id, dea_registration_id, store_number, legal_business_name, status_code_340b, chain_name FROM ncpdp_providers"
    return db_read(strm, svcCentralPool, qry, create)
}
func (s *svcServer) GetNDCs(req *Req, strm grpc.ServerStreamingServer[NDC]) error {
    create := func(obj map[string]any) any {
        ndc := &NDC{}
        ndc.Name = toStr(obj["product_name"])
        ndc.Ndc  = toStr(obj["item"])
        ndc.Netw = toStr(obj["network"])
        return ndc
    }
    qry := fmt.Sprintf("SELECT item, product_name, network FROM ndcs WHERE manufacturer_name = '%s'", req.Manu)
    return db_read(strm, svcCentralPool, qry, create)
}
func (s *svcServer) GetEntities(req *Req, strm grpc.ServerStreamingServer[Entity]) error {
    create := func(obj map[string]any) any {
        ent := &Entity{}
        ent.I340  = toStr(obj["id_340b"])
        ent.State = toStr(obj["state"])
        ent.Strt  = toU64(obj["participating_start_date"])
        ent.Term  = toU64(obj["term_date"])
        return ent
    }
    qry := "SELECT id_340b, state, participating_start_date, term_date FROM covered_entities"
    return db_read(strm, svcCentralPool, qry, create)
}
func (s *svcServer) GetPharmacies(req *Req, strm grpc.ServerStreamingServer[Pharmacy]) error {
    create := func(obj map[string]any) any {
        phm := &Pharmacy{}
        phm.Chnm = toStr(obj["chain_name"])
        phm.Dea  = toStr(obj["dea_id"])
        phm.Deas = toStrList(obj["dea"])
        phm.I340 = toStr(obj["id_340b"])
        phm.Ncp  = toStr(obj["ncpdp_provider_id"])
        phm.Ncps = toStrList(obj["ncpdp"])
        phm.Npi  = toStr(obj["national_provider_id"])
        phm.Npis = toStrList(obj["npi"])
        phm.Phid = toStr(obj["pharmacy_id"])
        phm.State= toStr(obj["pharmacy_state"])
        return phm
    }
    qry := "SELECT chain_name, id_340b, pharmacy_id, dea_id, national_provider_id, ncpdp_provider_id, dea, npi, ncpdp, pharmacy_state from contracted_pharmacies"
    return db_read(strm, svcCentralPool, qry, create)
}
func (s *svcServer) GetESP1Pharms(req *Req, strm grpc.ServerStreamingServer[ESP1PharmNDC]) error {
    create := func(obj map[string]any) any {
        phm := &ESP1PharmNDC{}
        phm.Ndc  = toStr(obj["ndc"])
        phm.Spid = toStr(obj["service_provider_id"])
        phm.Start= toU64(obj["start"])
        phm.Term = toU64(obj["term"])
        return phm
    }
    qry := fmt.Sprintf("SELECT service_provider_id, ndc, start, term FROM esp1_providers WHERE manufacturer = '%s'", req.Manu)
    return db_read(strm, svcCitusPool, qry, create)
}
func (s *svcServer) GetClaims(req *Req, strm grpc.ServerStreamingServer[Claim]) error {
    create := func(obj map[string]any) any {
        clm := &Claim{}
        clm.Chnm  = toStr(obj["chain_name"])
        clm.Clid  = toStr(obj["id"])
        clm.Cnfm  = toBool(obj["claim_conforms_flag"])
        clm.Doc   = toU64(obj["created_at"])
        clm.Dop   = toU64(obj["formatted_dop"])
        clm.Dos   = toU64(obj["formatted_dos"])
        clm.Hdop  = toStr(obj["date_prescribed"])
        clm.Hdos  = toStr(obj["date_of_service"])
        clm.Hfrx  = toStr(obj["formatted_rx_number"])
        clm.Hrxn  = toStr(obj["rx_number"])
        clm.I340  = toStr(obj["contracted_entity_id"])   // TODO: is this right?
        clm.Isgr  = IsGrantee(clm.I340)
        clm.Isih  = len(toStrList(obj["in_house_pharmacy_ids"])) > 0
        clm.Lauth = toStr(obj["rbt_hdos_auth"])
        clm.Lhdos = toStr(obj["rbt_hdos"])
        clm.Lownr = toStr(obj["rbt_hdos_owner"])
        clm.Lscid = toI64(obj["rbt_rrid"])
        clm.Manu  = toStr(obj["manufacturer"])
        clm.Ndc   = toStr(obj["ndc"])
        clm.Netw  = toStr(obj["network"])
        clm.Prnm  = toStr(obj["product_name"])
        clm.Qty   = toF64(obj["quantity"])
        clm.Shid  = toStr(obj["short_id"])
        clm.Spid  = toStr(obj["service_provider_id"])
        clm.Prid  = toStr(obj["prescriber_id"])
        clm.Elig  = toBool(obj["eligible_at_submission"])
        clm.Susp  = toBool(obj["suspended_at_submission"])
        clm.Valid = true
        return clm
    }
    qry := fmt.Sprintf("SELECT * FROM submission_rows WHERE manufacturer = '%s'", req.Manu)
    return db_read(strm, svcCitusPool, qry, create)
}
func (s *svcServer) GetEligibilityLedger(req *Req, strm grpc.ServerStreamingServer[Eligibility]) error {
    create := func(obj map[string]any) any {
        elg := &Eligibility{}
        elg.Elid = toI64(obj["id"])
        elg.I340 = toStr(obj["id_340b"])
        elg.Manu = toStr(obj["manufacturer"])
        elg.Netw = toStr(obj["network"])
        elg.Phid = toStr(obj["pharmacy_id"])
        elg.Strt = toU64(obj["start_at"])
        elg.Term = toU64(obj["end_at"])
        return elg
    }
    qry := fmt.Sprintf("SELECT id, id_340b, pharmacy_id, manufacturer, network, start_at, end_at FROM eligibility_ledger WHERE manufacturer = '%s'", req.Manu)
    return db_read(strm, svcCitusPool, qry, create)
}
func (s *svcServer) Start(ctx context.Context, req *StartReq) (*StartRes, error) {
    create := func(obj any) map[string]any {
        m := map[string]any{}
        m["auth"] = req.Auth
        m["cmdl"] = req.Cmdl
        m["desc"] = req.Desc
        m["auth"] = req.Hash
        m["auth"] = req.Hdrs
        m["auth"] = req.Host
        m["auth"] = req.Manu
        m["auth"] = req.Name
        m["auth"] = req.Plcy
        m["auth"] = req.Type
        m["auth"] = req.Vers
        return m
    }
    id, err := db_insert1(ctx, "scrubs", req, create)
    return &StartRes{ScrubId: id}, err
}
func (s *svcServer) AddRebates(strm grpc.ClientStreamingServer[RebateRec, Res]) error {
    create := func(obj any) map[string]any {
        m := map[string]any{}
        rr := obj.(*RebateRec)
        m["scrub_id"] = rr.ScrubId
        m["fprt"]     = rr.Fprt
        m["rnum"]     = rr.Rnum
        m["status"]   = rr.Status
        return m
    }
    return db_insert(strm, "rebates", create)
}
func (s *svcServer) AddClaims(strm grpc.ClientStreamingServer[ClaimRec, Res]) error {
    create := func(obj any) map[string]any {
        m := map[string]any{}
        cr := obj.(*ClaimRec)
        m["scrub_id"] = cr.ScrubId
        m["clm_guid"] = cr.Clid
        m["excl"]     = cr.Excl
        return m
    }
    return db_insert(strm, "claims", create)
}
func (s *svcServer) UpdateClaims(strm grpc.ClientStreamingServer[ClaimUpdate, Res]) error {
    return nil
}
func (s *svcServer) Done(ctx context.Context, m *Metrics) (*Res, error) {
    return nil, nil
}