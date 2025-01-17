package main

import (
	context "context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type atlasServer struct {
	UnimplementedAtlasServer
	cmid int64
}

type atlasServerStream struct {
	grpc.ServerStream
}

func atlasUnaryServerInterceptor(ctx context.Context, req any, si *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	cmd := &Command{
		Comd: si.FullMethod,
		Manu: metaGet(ctx, "manu"),
		Name: metaGet(ctx, "name"),
		Auth: metaGet(ctx, "auth"),
		Vers: metaGet(ctx, "vers"),
		Kind: metaGet(ctx, "kind"),
		Dscr: metaGet(ctx, "dscr"),
		Hash: metaGet(ctx, "hash"),
		Netw: metaGet(ctx, "netw"),
		Host: metaGet(ctx, "host"),
		Cwd:  metaGet(ctx, "cwd"),
		User: metaGet(ctx, "user"),
		Cmdl: metaGet(ctx, "cmdl"),
		Addr: getPublicAddr(ctx),
		Crat: 0,
	}
	strt := time.Now()
	pool := atlas.pools["atlas"]
	vald := validate_client(ctx, pool, "atlas")
	if vald != nil {
		cmd.Rslt = vald.Error()
	}
	dbm := new_dbmap[Command]()
	dbm.table(pool, "atlas.commands")

	if cmid, err := db_insert_one[Command](ctx, pool, "atlas.commands", dbm, cmd, "cmid"); err == nil {
		srvr := si.Server.(*atlasServer)
		srvr.cmid = cmid
		if vald == nil {
			if res, err := handler(ctx, req); err != nil {
				Log("atlas", "unary_int", si.FullMethod, "command failed", time.Since(strt), map[string]any{"cmid": cmid, "name": cmd.Name, "auth": cmd.Auth, "user": cmd.User, "netw": cmd.Netw, "host": cmd.Host}, err)
				cmd.Rslt = err.Error()
				if err := db_update(context.Background(), cmd, nil, pool, "atlas.commands", dbm, map[string]string{"cmid": fmt.Sprintf("%d", cmid)}); err != nil {
					Log("atlas", "unary_int", si.FullMethod, "failed to update command row", time.Since(strt), map[string]any{"cmid": cmid, "name": cmd.Name, "auth": cmd.Auth, "user": cmd.User, "netw": cmd.Netw, "host": cmd.Host}, err)
				}
				return res, err
			} else {
				Log("atlas", "unary_int", si.FullMethod, "command succeeded", time.Since(strt), map[string]any{"cmid": cmid, "name": cmd.Name, "auth": cmd.Auth, "user": cmd.User, "netw": cmd.Netw, "host": cmd.Host}, err)
				return res, err
			}
		} else {
			Log("atlas", "unary_int", si.FullMethod, "validation failed", time.Since(strt), map[string]any{"cmid": cmid, "name": cmd.Name, "auth": cmd.Auth, "user": cmd.User, "netw": cmd.Netw, "host": cmd.Host}, vald)
			return nil, vald
		}
	} else {
		Log("atlas", "unary_int", si.FullMethod, "failed to insert command row", time.Since(strt), map[string]any{"cmid": cmid, "name": cmd.Name, "auth": cmd.Auth, "user": cmd.User, "netw": cmd.Netw, "host": cmd.Host}, err)
		return nil, err
	}
}

func atlasStreamServerInterceptor(srv any, ss grpc.ServerStream, si *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	cmd := &Command{
		Comd: si.FullMethod,
		Manu: metaGet(ss.Context(), "manu"),
		Name: metaGet(ss.Context(), "name"),
		Auth: metaGet(ss.Context(), "auth"),
		Vers: metaGet(ss.Context(), "vers"),
		Kind: metaGet(ss.Context(), "kind"),
		Dscr: metaGet(ss.Context(), "dscr"),
		Hash: metaGet(ss.Context(), "hash"),
		Netw: metaGet(ss.Context(), "netw"),
		Host: metaGet(ss.Context(), "host"),
		Cwd:  metaGet(ss.Context(), "cwd"),
		User: metaGet(ss.Context(), "user"),
		Cmdl: metaGet(ss.Context(), "cmdl"),
		Addr: getPublicAddr(ss.Context()),
		Crat: 0,
	}
	strt := time.Now()
	pool := atlas.pools["atlas"]
	vald := validate_client(ss.Context(), pool, "atlas")
	if vald != nil {
		cmd.Rslt = vald.Error()
	}
	dbm := new_dbmap[Command]()
	dbm.table(pool, "atlas.commands")
	
	if cmid, err := db_insert_one[Command](context.Background(), pool, "atlas.commands", dbm, cmd, "cmid"); err == nil {
		srvr := srv.(*atlasServer)
		srvr.cmid = cmid
		if vald == nil {
			if err := handler(srv, &atlasServerStream{ss}); err != nil {
				Log("atlas", "stream_int", si.FullMethod, "command failed", time.Since(strt), map[string]any{"cmid": cmid, "name": cmd.Name, "auth": cmd.Auth, "user": cmd.User, "netw": cmd.Netw, "host": cmd.Host}, err)
				cmd.Rslt = err.Error()
				if err := db_update(context.Background(), cmd, nil, pool, "atlas.commands", dbm, map[string]string{"cmid": fmt.Sprintf("%d", cmid)}); err != nil {
					Log("atlas", "stream_int", si.FullMethod, "failed to update command row", time.Since(strt), map[string]any{"cmid": cmid, "name": cmd.Name, "auth": cmd.Auth, "user": cmd.User, "netw": cmd.Netw, "host": cmd.Host}, err)
				}
				return err
			} else {
				Log("atlas", "stream_int", si.FullMethod, "command succeeded", time.Since(strt), map[string]any{"cmid": cmid, "name": cmd.Name, "auth": cmd.Auth, "user": cmd.User, "netw": cmd.Netw, "host": cmd.Host}, err)
				return nil
			}
		} else {
			Log("atlas", "stream_int", si.FullMethod, "validation failed", time.Since(strt), map[string]any{"cmid": cmid, "name": cmd.Name, "auth": cmd.Auth, "user": cmd.User, "netw": cmd.Netw, "host": cmd.Host}, vald)
			return vald
		}
	} else {
		Log("atlas", "stream_int", si.FullMethod, "failed to insert command row", time.Since(strt), map[string]any{"cmid": cmid, "name": cmd.Name, "auth": cmd.Auth, "user": cmd.User, "netw": cmd.Netw, "host": cmd.Host}, err)
		return err
	}
}

func (as *atlasServerStream) RecvMsg(m any) error {
	return as.ServerStream.RecvMsg(m)
}

func (as *atlasServerStream) SendMsg(m any) error {
	return as.ServerStream.SendMsg(m)
}

func (s *atlasServer) Ping(ctx context.Context, in *Req) (*Res, error) {
	ou := ""
	cn := ""
	netw := ""
	if p, ok := peer.FromContext(ctx); ok && p != nil {
		netw = p.Addr.String()
		if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			cn, ou = getCreds(tlsInfo)
		}
	}
	Log("titan", "ping", cn, "", 0, map[string]any{
		"cn": cn,
		"ou": ou,
		"manu": manu,
		"netw": netw,
	}, nil)
	return &Res{}, nil
}

type invoice struct {
	Manu string
	File string
	Cmid int64
}
type invoice_col struct {
	Ivid int64
	Indx int
	Name string
}
type invoice_row struct {
	Ivid int64
	Rbid int64
	Data string
}

func (s *atlasServer) Invoice(strm grpc.ClientStreamingServer[Rebate, Res]) error {
	pool := atlas.pools["atlas"]
	manu := metaGet(strm.Context(), "manu")
	file := metaGet(strm.Context(), "file")
	hdrs := metaGet(strm.Context(), "hdrs")
	inv  := &invoice{Manu: manu, File: file, Cmid: s.cmid}

	strt := time.Now()
	ivid := int64(-1)
	var err error

	// Invoice row
	if ivid, err = db_insert_one[invoice](strm.Context(), pool, "atlas.invoices", nil, inv, "ivid"); err == nil {
		if err := grpc.SendHeader(strm.Context(), metadata.Pairs("ivid", fmt.Sprintf("%d", ivid))); err != nil {
			Log("atlas", "Invoice", "ivid", "return header to client", time.Since(strt), map[string]any{"manu": manu, "file": file, "cmid": s.cmid, "ivid": ivid}, err)
		}
	} else {
		Log("atlas", "Invoice", "atlas.invoices", "invoice creation", time.Since(strt), map[string]any{"manu": manu, "file": file, "cmid": s.cmid, "ivid": ivid}, err)
		return err
	}

	var wg sync.WaitGroup

	// Invoice columns
	ichn := make(chan *invoice_col, 100)
	wg.Add(1)
	db_insert_run(&wg, pool, "atlas", "atlas.invoice_cols", nil, ichn, 100, "", false, nil, nil, nil)
	toks := strings.Split(hdrs, ",")
	for i, tok := range toks {
		ichn <-&invoice_col{Ivid: ivid, Indx: i, Name: tok}
	}
	close(ichn)
	wg.Wait()
	
	// Invoice rebate rows (and row data rows)
	rchn := make(chan *Rebate,      5000)
	dchn := make(chan *invoice_row, 5000)
	var rbt *Rebate
	wg.Add(2)
	db_insert_run(&wg, pool, "atlas", "atlas.invoice_rows", nil, dchn, 5000, "", false, nil, nil, nil)
	db_insert_run(&wg, pool, "atlas", "atlas.rebates",      nil, rchn, 5000, "", false, nil, nil, nil)
	rbid := int64(0)
	for {
		if rbt, err = strm.Recv(); err == nil {
			rbt.Ivid = ivid
			rbt.Rbid = rbid
			rchn <-rbt
			dchn <-&invoice_row{Ivid: ivid, Rbid: rbid, Data: rbt.Data}
			rbid++
		} else if err == io.EOF {
			Log("atlas", "Invoice", "", "invoice created", time.Since(strt), map[string]any{"manu": manu, "file": file, "cmid": fmt.Sprintf("%d", s.cmid), "ivid": ivid, "cnt": rbid+1}, nil)
			break
		} else {
			Log("atlas", "Invoice", "", "error while reading rebates from stream", time.Since(strt), map[string]any{"manu": manu, "file": file, "cmid": fmt.Sprintf("%d", s.cmid), "ivid": ivid, "cnt": rbid+1}, err)
			break
		}
	}
	close(rchn)
	close(dchn)
	wg.Wait()
	if err == io.EOF {
		err = nil
	}
	return err
}

func (s *atlasServer) Scrub(req *ScrubReq, strm grpc.ServerStreamingServer[RebateResult]) error {
	scr := &Scrub{
		Ivid: req.Ivid,
		Cmid: metaGet(strm.Context(), "cmid"),
		Plcy: metaGet(strm.Context(), "plcy"),
		Kind: metaGet(strm.Context(), "kind"),
		Appl: metaGet(strm.Context(), "appl"),
		Hdrs: metaGet(strm.Context(), "hdrs"),
		Test: metaGet(strm.Context(), "test"),
	}
	pool := atlas.pools["atlas"]
	strt := time.Now()

	if scid, err := db_insert_one[Scrub](strm.Context(), pool, "atlas.scrubs", nil, scr, "scid"); err == nil {
		scr.Scid = scid
	} else {
		return err
	}
	grpc.SendHeader(strm.Context(), metadata.Pairs("scid", fmt.Sprintf("%d", scr.Scid)))

	stop := make(chan any)
	scrb := new_scrub(scr, stop)
	defer close(stop)
	
	atlas.scrubs[scrb.scid] = scrb
	scrb.run()

	// Send the rebates to the shell (client).
	whr := fmt.Sprintf("scid = %d", scrb.scid)
	if chn, err := db_select[Rebate](pool, "atlas", "atlas.rebates", nil, whr, "Rbid", stop); err == nil {
		for rbt := range chn {
			rr := &RebateResult{
				Rbid:          rbt.Rbid,
				Indx:          rbt.Indx,
				Rxn:           rbt.Rxn,
				Hrxn:          rbt.Hrxn,
				Ndc:           rbt.Ndc,
				Spid:          rbt.Spid,
				Prid:          rbt.Prid,
				Dos:           rbt.Dos,
				Stat:          rbt.Stat,
				//Excl:          rbt.Excl,
				Spmt:          rbt.Spmt,
				Errc:          rbt.Errc,
				Errm:          rbt.Errm,
			}
			if err := strm.Send(rr); err != nil {
				Log("atlas", "Scrub", "strm.Send", "failed to send, aborting", time.Since(strt), nil, err)
				return err
			}
		}
	} else {
		Log("atlas", "Scrub", "db_select", "failed", time.Since(strt), nil, err)
		return err
	}
	return nil
}

func (s *atlasServer) UploadTest(ctx context.Context, td *TestData) (*Res, error) {
	return nil, nil
}

