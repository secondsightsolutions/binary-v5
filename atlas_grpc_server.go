package main

import (
	context "context"
	"fmt"
	"io"
	"slices"
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
		Xou:  metaGet(ctx, "xou"),
		Manu: metaGet(ctx, "manu"),
		Name: metaGet(ctx, "name"),
		Auth: metaGet(ctx, "auth"),
		Vers: metaGet(ctx, "vers"),
		Kind: metaGet(ctx, "kind"),
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
		Xou:  metaGet(ss.Context(), "xou"),
		Manu: metaGet(ss.Context(), "manu"),
		Name: metaGet(ss.Context(), "name"),
		Auth: metaGet(ss.Context(), "auth"),
		Vers: metaGet(ss.Context(), "vers"),
		Kind: metaGet(ss.Context(), "kind"),
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

type invoice struct {
	Manu string
	File string
	Cmid int64
}
type invoice_col struct {
	Manu string
	Ivid int64
	Indx int
	Name string
}
type invoice_row struct {
	Manu string
	Ivid int64
	Rbid int64
	Data string
}
type scrub_row struct {
	Rbid int64
	Indx int64
	Stat string
	Excl string
	Spmt string
	Errc string
	Errm string
	Data string
}

func atlas_db_read[T any](ctx context.Context, fcnn, tbln string, ivid, scid int64, sort string) (*T, error) {
	strt := time.Now()
	stop := make(chan any)
	defer close(stop)

	pool := atlas.pools["atlas"]
	whr  := []string{}
	whrS := ""
	manu := metaManu(ctx)
	dbm  := new_dbmap[T]()
	dbm.table(pool, tbln)

	if scid > -1 && dbm.find("scid", false, true) != nil {
		whr = append(whr, fmt.Sprintf("scid = %d", scid))
	}
	if ivid > -1 && dbm.find("ivid", false, true) != nil {
		whr = append(whr, fmt.Sprintf("ivid = %d", ivid))
	}
	if manu != "" && dbm.find("manu", false, true) != nil {
		whr = append(whr, fmt.Sprintf("manu = '%s'", manu))
	}
	if len(whr) > 0 {
		for _, w := range whr {
			if len(whrS) > 0 {
				whrS += " AND "
			}
			whrS += w
		}
	}
	// Send the rows to the shell (client).
	if chn, err := db_select[T](pool, "atlas", tbln, dbm, whrS, sort, stop); err == nil {
		row := <-chn
		return row, nil
	} else {
		Log("atlas", fcnn, "db_select", "failed", time.Since(strt), nil, err)
		return nil, err
	}
}

func atlas_db_read_strm[T any](fcnn, tbln string, strm grpc.ServerStreamingServer[T], ivid, scid int64, sort string) error {
	strt := time.Now()
	stop := make(chan any)
	defer close(stop)

	pool := atlas.pools["atlas"]
	whr  := []string{}
	whrS := ""
	manu := metaManu(strm.Context())
	dbm  := new_dbmap[T]()
	dbm.table(pool, tbln)

	if scid > -1 && dbm.find("scid", false, true) != nil {
		whr = append(whr, fmt.Sprintf("scid = %d", scid))
	}
	if ivid > -1 && dbm.find("ivid", false, true) != nil {
		whr = append(whr, fmt.Sprintf("ivid = %d", ivid))
	}
	if manu != "" && dbm.find("manu", false, true) != nil {
		whr = append(whr, fmt.Sprintf("manu = '%s'", manu))
	}
	if len(whr) > 0 {
		for _, w := range whr {
			if len(whrS) > 0 {
				whrS += " AND "
			}
			whrS += w
		}
	}
	// Send the rows to the shell (client).
	if chn, err := db_select[T](pool, "atlas", tbln, dbm, whrS, sort, stop); err == nil {
		for row := range chn {
			if err := strm.Send(row); err != nil {
				Log("atlas", fcnn, "strm.Send", "failed", time.Since(strt), nil, err)
				return err
			}
		}
	} else {
		Log("atlas", fcnn, "db_select", "failed", time.Since(strt), nil, err)
		return err
	}
	return nil
}

func (s *atlasServer) UploadInvoice(strm grpc.ClientStreamingServer[Rebate, Res]) error {
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
			Log("atlas", "UploadInvoice", "ivid", "return header to client", time.Since(strt), map[string]any{"manu": manu, "file": file, "cmid": s.cmid, "ivid": ivid}, err)
		}
	} else {
		Log("atlas", "UploadInvoice", "atlas.invoices", "invoice creation", time.Since(strt), map[string]any{"manu": manu, "file": file, "cmid": s.cmid, "ivid": ivid}, err)
		return err
	}

	var wg sync.WaitGroup

	// Invoice columns
	ichn := make(chan *invoice_col, 100)
	wg.Add(1)
	db_insert_run(&wg, pool, "atlas", "atlas.invoice_cols", nil, ichn, 100, "", false, false, nil, nil, nil)
	toks := strings.Split(hdrs, ",")
	for i, tok := range toks {
		ichn <-&invoice_col{Manu: manu, Ivid: ivid, Indx: i, Name: tok}
	}
	close(ichn)
	wg.Wait()
	
	// Invoice rebate rows (and row data rows)
	rchn := make(chan *Rebate,      5000)
	dchn := make(chan *invoice_row, 5000)
	var rbt *Rebate
	wg.Add(2)
	db_insert_run(&wg, pool, "atlas", "atlas.invoice_rows", nil, dchn, 5000, "", false, true, nil, nil, nil)
	db_insert_run(&wg, pool, "atlas", "atlas.rebates",      nil, rchn, 5000, "", false, true, nil, nil, nil)
	rbid := int64(0)
	for {
		if rbt, err = strm.Recv(); err == nil {
			rbt.Manu = manu
			rbt.Ivid = ivid
			rbt.Rbid = rbid
			rchn <-rbt
			dchn <-&invoice_row{Manu: manu, Ivid: ivid, Rbid: rbid, Data: rbt.Data}
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

func (s *atlasServer) RunScrub(req *InvoiceIdent, strm grpc.ServerStreamingServer[Metrics]) error {
	scr := &Scrub{
		Ivid: req.Ivid,
		Cmid: fmt.Sprintf("%d", s.cmid),
		Manu: metaGet(strm.Context(), "manu"),
		Plcy: metaGet(strm.Context(), "plcy"),
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

	// Send the scrub updates to the shell (client).
	for {
		select {
		case <-time.After(time.Duration(1)*time.Second):
			scrb.lckM.Lock()
			strm.Send(scrb.metr)
			scrb.lckM.Unlock()
		case <-scrb.done:
			Log("atlas", "Scrub", "", "completed", time.Since(strt), map[string]any{
				"cmid": scr.Cmid,
				"scid": scr.Scid,
				"ivid": scr.Ivid,
				"plcy": scr.Plcy,
				"manu": metaGet(strm.Context(), "manu"),
				"name": metaGet(strm.Context(), "name"),
			}, nil)
			return nil
		}
	}
}

func (s *atlasServer) RunQueue(ctx context.Context, req *InvoiceIdent) (*ScrubRes, error) {
	scr := &Scrub{
		Ivid: req.Ivid,
		Cmid: fmt.Sprintf("%d", s.cmid),
		Manu: metaGet(ctx, "manu"),
		Plcy: metaGet(ctx, "plcy"),
		Hdrs: metaGet(ctx, "hdrs"),
		Test: metaGet(ctx, "test"),
	}
	res := &ScrubRes{Scid: -1}
	pool := atlas.pools["atlas"]
	strt := time.Now()

	if scid, err := db_insert_one[Scrub](ctx, pool, "atlas.scrubs", nil, scr, "scid"); err == nil {
		scr.Scid = scid
	} else {
		return res, err
	}

	scrb := new_scrub(scr, nil)
	
	atlas.scrubs[scrb.scid] = scrb
	go scrb.run()

	Log("atlas", "ScrubBG", "", "started", time.Since(strt), map[string]any{
		"cmid": scr.Cmid,
		"scid": scr.Scid,
		"ivid": scr.Ivid,
		"plcy": scr.Plcy,
		"manu": metaGet(ctx, "manu"),
		"name": metaGet(ctx, "name"),
	}, nil)
	return res, nil
}

func (s *atlasServer) GetScrub(ctx context.Context, req *ScrubIdent) (*Scrub, error) {
	return atlas_db_read[Scrub](ctx, "GetScrub", "atlas.scrubs", -1, req.Scid, "")
}

func (s *atlasServer) GetScrubMetrics(ctx context.Context, req *ScrubIdent) (*Metrics, error) {
	return atlas_db_read[Metrics](ctx, "GetScrubMetrics", "atlas.metrics", -1, req.Scid, "")
}

func (s *atlasServer) GetScrubRebates(req *ScrubIdent, strm grpc.ServerStreamingServer[ScrubRebate]) error {
	return atlas_db_read_strm("GetScrubRebates", "atlas.scrub_rebates", strm, -1, req.Scid, "Rbid")
}

func (s *atlasServer) GetScrubFile(req *ScrubIdent, strm grpc.ServerStreamingServer[ScrubRow]) error {
	strt := time.Now()
	hdrs := metaGet(strm.Context(), "hdrs")		// Comma-separated list of headers indicating columns to return.
	scid := metaGetI64(strm.Context(), "scid")
	ivid := int64(-1)
	stop := make(chan any)
	pool := atlas.pools["atlas"]
	cols := []string{}
	defer close(stop)

	// Get invoice id
	if scr, err := atlas_db_read[Scrub](context.Background(), "GetScrubFile", "atlas.scrubs", -1, scid, ""); err == nil {
		ivid = scr.Ivid
	}

	// Get invoice columns
	if chn, err := db_select[invoice_col](pool, "atlas", "atlas.invoice_cols", nil, fmt.Sprintf("ivid = %d", ivid), "indx", stop); err == nil {
		for ic := range chn {
			cols = append(cols, ic.Name)
		}
	}

	// Get the scrub rebates (joined to invoice rows)
	qry := `
		select sr.rbid, sr.indx, sr.stat, sr.excl, sr.spmt, ir."data" 
		from atlas.scrub_rebates sr
		join atlas.invoice_rows  ir on (sr.ivid = ir.ivid and sr.rbid = ir.rbid)
		where sr.scid = %d
		order by ir.rbid;
	`
	qry = fmt.Sprintf(qry, scid)

	dbm := new_dbmap[scrub_row]()
	dbm.column("rbid", "rbid", "")
	dbm.column("stat", "stat", "")
	dbm.column("excl", "excl", "")
	dbm.column("spmt", "spmt", "")
	dbm.column("data", "data", "")

	// Make sure we have the four basic/added columns in the list.
	Hdrs := split(hdrs, ",", strings.Join(cols, ","))	// If specific hdr list provided, use it. Else, all hdrs.
	if !slices.Contains(Hdrs, "stat") {
		newh := []string{"stat"}
		newh = append(newh, Hdrs...)
		Hdrs = newh
	}
	if !slices.Contains(Hdrs, "excl") {
		Hdrs = append(Hdrs, "excl")
	}
	if !slices.Contains(Hdrs, "errc") {
		Hdrs = append(Hdrs, "errc")
	}
	if !slices.Contains(Hdrs, "errm") {
		Hdrs = append(Hdrs, "errm")
	}
	
	if chn, err := db_select_cust[scrub_row](pool, "atlas", dbm, qry, stop); err == nil {
		for sr := range chn {
			srow := &ScrubRow{}
			line := []string{}
			cols := split(sr.Data, ",", "")
			coli := 0
			for _, hdr := range Hdrs {
				if hdr == "stat" {
					line = append(line, sr.Stat)
				} else if hdr == "excl" {
					line = append(line, sr.Excl)
				} else if hdr == "errc" {
					line = append(line, sr.Errc)
				} else if hdr == "errm" {
					line = append(line, sr.Errm)
				} else {
					line = append(line, cols[coli])
					coli++
				}
			}
			srow.Row = strings.Join(line, ",")
			if err := strm.Send(srow); err != nil {
				Log("atlas", "GetScrubFile", "", "stream send failed", time.Since(strt), nil, err)
				return err
			}
		}
	}
	return nil
}

func (s *atlasServer) GetInvoice(ctx context.Context, req *InvoiceIdent) (*Invoice, error) {
	return atlas_db_read[Invoice](ctx, "GetInvoice", "atlas.invoices", -1, req.Ivid, "")
}

func (s *atlasServer) GetInvoiceRebates(req *InvoiceIdent, strm grpc.ServerStreamingServer[Rebate]) error {
	return atlas_db_read_strm("GetInvoiceRebates", "atlas.rebates", strm, -1, req.Ivid, "Rbid")
}

func (s *atlasServer) UploadTest(ctx context.Context, td *TestData) (*Res, error) {
	return nil, nil
}

