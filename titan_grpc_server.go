package main

import (
	context "context"
	"fmt"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type titanServer struct {
	UnimplementedTitanServer
	rqid int64
}

type titanStream struct {
	grpc.ServerStream
}

type request struct {
	Rqid int64
	Seq  int64
	Cmid string
	Comd string // GRPC service API call (endpoint/function name)
	Manu string
	Name string
	Kind string
	Auth string
	Vers string
	Xou  string
	Dscr string 
	Hash string
	Netw string
	Host string
	User string
	Addr string	// public address seen on incoming command
	Cmdl string
	Cwd  string	// current working directory
	Rslt string
	Crat int64
}


func (ts *titanStream) RecvMsg(m any) error {
	return ts.ServerStream.RecvMsg(m)
}

func (ts *titanStream) SendMsg(m any) error {
	return ts.ServerStream.SendMsg(m)
}

func newTitanStream(s grpc.ServerStream) grpc.ServerStream {
	return &titanStream{s}
}

func titanUnaryInterceptor(ctx context.Context, req any, si *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	rqst := &request{
		Comd: si.FullMethod,
		Cmid: metaGet(ctx, "cmid"),
		Manu: metaGet(ctx, "manu"),
		Name: metaGet(ctx, "name"),
		Auth: metaGet(ctx, "auth"),
		Vers: metaGet(ctx, "vers"),
		Kind: metaGet(ctx, "kind"),
		Xou:  metaGet(ctx, "xou"),
		Dscr: metaGet(ctx, "dscr"),
		Hash: metaGet(ctx, "hash"),
		Netw: metaGet(ctx, "netw"),
		Host: metaGet(ctx, "mach"),
		Cwd:  metaGet(ctx, "cwd"),
		User: metaGet(ctx, "user"),
		Cmdl: metaGet(ctx, "cmdl"),
		Addr: getPublicAddr(ctx),
		Crat: 0,
	}
	strt := time.Now()
	pool := titan.pools["titan"]
	vald := validate_client(ctx, pool, "titan")
	if vald != nil {
		rqst.Rslt = vald.Error()
	}
	dbm := new_dbmap[request]()
	dbm.table(pool, "titan.requests")

	if rqid, err := db_insert_one[request](ctx, pool, "titan.requests", dbm, rqst, "rqid"); err == nil {
		srvr := si.Server.(*titanServer)
		srvr.rqid = rqid
		if vald == nil {
			if res, err := handler(ctx, req); err != nil {
				Log("titan", "unary_int", si.FullMethod, "command failed", time.Since(strt), nil, err)
				rqst.Rslt = err.Error()
				if err := db_update(context.Background(), rqst, nil, pool, "titan.requests", dbm, map[string]string{"rqid": fmt.Sprintf("%d", rqid)}); err != nil {
					Log("titan", "unary_int", si.FullMethod, "failed to update request row", time.Since(strt), map[string]any{"rqid": rqid, "name": rqst.Name, "auth": rqst.Auth, "user": rqst.User, "netw": rqst.Netw, "host": rqst.Host}, err)
				}
				return res, err
			} else {
				// Log("titan", "unary_int", si.FullMethod, "request succeeded", time.Since(strt), nil, err)
				return res, err
			}
		} else {
			Log("titan", "unary_int", si.FullMethod, "validation failed", time.Since(strt), nil, vald)
			return nil, vald
		}
	} else {
		Log("titan", "unary_int", si.FullMethod, "failed to insert request row", time.Since(strt), map[string]any{"rqid": rqid, "name": rqst.Name, "auth": rqst.Auth, "user": rqst.User, "netw": rqst.Netw, "host": rqst.Host}, err)
		return nil, err
	}
}

func titanStreamInterceptor(srv any, ss grpc.ServerStream, si *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	rqst := &request{
		Comd: si.FullMethod,
		Cmid: metaGet(ss.Context(), "cmid"),
		Manu: metaGet(ss.Context(), "manu"),
		Name: metaGet(ss.Context(), "name"),
		Auth: metaGet(ss.Context(), "auth"),
		Vers: metaGet(ss.Context(), "vers"),
		Kind: metaGet(ss.Context(), "kind"),
		Xou:  metaGet(ss.Context(), "xou"),
		Dscr: metaGet(ss.Context(), "dscr"),
		Hash: metaGet(ss.Context(), "hash"),
		Netw: metaGet(ss.Context(), "netw"),
		Host: metaGet(ss.Context(), "mach"),
		Cwd:  metaGet(ss.Context(), "cwd"),
		User: metaGet(ss.Context(), "user"),
		Cmdl: metaGet(ss.Context(), "cmdl"),
		Addr: getPublicAddr(ss.Context()),
		Crat: 0,
	}
	strt := time.Now()
	pool := titan.pools["titan"]
	vald := validate_client(ss.Context(), pool, "titan")
	if vald != nil {
		rqst.Rslt = vald.Error()
	}
	dbm := new_dbmap[request]()
	dbm.table(pool, "titan.requests")

	if rqid, err := db_insert_one[request](context.Background(), pool, "titan.requests", dbm, rqst, "rqid"); err == nil {
		srvr := srv.(*titanServer)
		srvr.rqid = rqid
		if vald == nil {
			if err := handler(srvr, newTitanStream(ss)); err != nil {
				Log("titan", "stream_int", si.FullMethod, "request failed", time.Since(strt), nil, err)
				rqst.Rslt = err.Error()
				if err := db_update(context.Background(), rqst, nil, pool, "titan.requests", dbm, map[string]string{"rqid": fmt.Sprintf("%d", rqid)}); err != nil {
					Log("titan", "stream_int", si.FullMethod, "failed to update request row", time.Since(strt), map[string]any{"rqid": rqid, "name": rqst.Name, "auth": rqst.Auth, "user": rqst.User, "netw": rqst.Netw, "host": rqst.Host}, err)
				}
				return err
			} else {
				// Log("titan", "stream_int", si.FullMethod, "request succeeded", time.Since(strt), nil, err)
				return nil
			}
		} else {
			Log("titan", "stream_int", si.FullMethod, "validation failed", time.Since(strt), nil, vald)
			return vald
		}
	} else {
		Log("titan", "stream_int", si.FullMethod, "failed to insert request row", time.Since(strt), map[string]any{"rqid": rqid, "name": rqst.Name, "auth": rqst.Auth, "user": rqst.User, "netw": rqst.Netw, "host": rqst.Host}, err)
		return err
	}
}

func (s *titanServer) Ping(ctx context.Context, req *Req) (*Res, error) {
	xou  := ""
	xcn  := ""
	xorg := ""
	addr := ""
	if p, ok := peer.FromContext(ctx); ok && p != nil {
		addr = p.Addr.String()
		if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			xcn, xou, xorg = getCreds(tlsInfo)
		}
	}
	Log("titan", "ping", xcn, "", 0, map[string]any{
		"xcn":  xcn,
		"xou":  xou,
		"xorg": xorg,
		"addr": addr,
		"manu": metaGet(ctx, "manu"),
		"netw": metaGet(ctx, "netw"),
		"name": metaGet(ctx, "name"),
		"kind": metaGet(ctx, "kind"),
		"host": metaGet(ctx, "mach"),
	}, nil)
	return &Res{}, nil
}

func titan_db_read[T any](tbln string, strm grpc.ServerStreamingServer[T], seq int64) error {
	pool := titan.pools["titan"]
	whr  := ""
	manu := metaManu(strm.Context())
	dbm  := new_dbmap[T]()
	dbm.table(pool, tbln)
	if seq > 0 {
		whr = fmt.Sprintf("seq > %d", seq)
	}
	if dbm.find("manu", false, true) != nil {
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
func (s *titanServer) GetDesignations(req *SyncReq, strm grpc.ServerStreamingServer[Designation]) error {
	return titan_db_read("titan.desigs", strm, req.Last)
}
func (s *titanServer) GetLDNs(req *SyncReq, strm grpc.ServerStreamingServer[LDN]) error {
	return titan_db_read("titan.ldns", strm, req.Last)
}

func (s *titanServer) SyncScrubRebates(strm grpc.ClientStreamingServer[ScrubRebate, Res]) error {
	_,_, err := sync_fm_client(titan.pools["titan"], "titan", "titan.scrub_rebates", true, strm)
	return err
}
func (s *titanServer) SyncScrubs(strm grpc.ClientStreamingServer[Scrub, Res]) error {
	_,_, err := sync_fm_client(titan.pools["titan"], "titan", "titan.scrubs", true, strm)
	return err
}
func (s *titanServer) SyncScrubClaims(strm grpc.ClientStreamingServer[ScrubClaim, Res]) error {
	_,_, err := sync_fm_client(titan.pools["titan"], "titan", "titan.scrub_claims", true, strm)
	return err
}
func (s *titanServer) SyncScrubMatches(strm grpc.ClientStreamingServer[ScrubMatch, Res]) error {
	_,_, err := sync_fm_client(titan.pools["titan"], "titan", "titan.scrub_matches", true, strm)
	return err
}
func (s *titanServer) SyncScrubAttempts(strm grpc.ClientStreamingServer[ScrubAttempt, Res]) error {
	_,_, err := sync_fm_client(titan.pools["titan"], "titan", "titan.scrub_attempts", true, strm)
	return err
}
// SyncMetrics(grpc.ClientStreamingServer[Metrics, Res]) error
func (s *titanServer) SyncMetrics(strm grpc.ClientStreamingServer[Metrics, Res]) error {
	_,_, err := sync_fm_client(titan.pools["titan"], "titan", "titan.metrics", true, strm)
	return err
}
func (s *titanServer) SyncCommands(strm grpc.ClientStreamingServer[Command, Res]) error {
	_,_, err := sync_fm_client(titan.pools["titan"], "titan", "titan.commands", true, strm)
	return err
}
