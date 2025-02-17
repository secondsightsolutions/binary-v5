package main

import (
	"bytes"
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
	db_insert_run(&wg, pool, "atlas", "atlas.invoice_cols", nil, ichn, 100, "", false, false, nil, nil, nil, nil, stop)
	toks := strings.Split(hdrs, ",")
	for i, tok := range toks {
		ichn <-&invoice_col{Manu: manu, Ivid: ivid, Indx: i, Name: tok}
	}
	close(ichn)
	wg.Wait()
	
	// Invoice rebate rows (and row data rows)
	rchn := make(chan *Rebate, 5000)
	var rbt *Rebate
	wg.Add(1)
	db_insert_run(&wg, pool, "atlas", "atlas.rebates", nil, rchn, 5000, "", false, true, nil, nil, nil, nil, stop)
	rbid := int64(0)
	setStrHashed := func(fm, to *string) {
		if *fm != "" && *to == "" {
			if Is64bitHash(*fm) {
				*to = *fm
			} else {
				*to,_ = Hash(*fm)
			}
		}
	}
	setUnixHashed := func(fm int64, to *string) {
		if fm > 0 && *to == "" {
			*to,_ = Hash(ParseI64ToTime(fm).Format("2006-01-02"))
		}
	}
	for {
		if rbt, err = strm.Recv(); err == nil {
			rbt.Manu = manu
			rbt.Ivid = ivid
			rbt.Rbid = rbid
			setStrHashed(&rbt.Rxn, &rbt.Hrxn)	// Let's possibly fix up a couple things on the rebate.
			setUnixHashed(rbt.Dos, &rbt.Hdos)
			rchn <-rbt
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
		Prof: metaGet(strm.Context(), "prof"),
		Test: metaGet(strm.Context(), "test"),
	}
	pool := atlas.pools["atlas"]

	if scid, err := db_insert_one[Scrub](strm.Context(), pool, "atlas.scrubs", nil, scr, "scid"); err == nil {
		scr.Scid = scid
	} else {
		return err
	}
	grpc.SendHeader(strm.Context(), metadata.Pairs("scid", fmt.Sprintf("%d", scr.Scid)))

	scrb := new_scrub(scr)
	
	atlas.scrubs[scrb.scid] = scrb
	go scrb.run()

	// Send the scrub updates to the shell (client).
	for range time.After(time.Duration(1)*time.Second) {
		scrb.lckM.Lock()
		strm.Send(scrb.metr)
		scrb.lckM.Unlock()

		if scrb.done {
			return nil
		}
	}
	return nil
}

func (s *atlasServer) RunQueue(ctx context.Context, req *InvoiceIdent) (*ScrubRes, error) {
	scr := &Scrub{
		Ivid: req.Ivid,
		Cmid: fmt.Sprintf("%d", s.cmid),
		Manu: metaGet(ctx, "manu"),
		Plcy: metaGet(ctx, "plcy"),
		Test: metaGet(ctx, "test"),
	}
	res  := &ScrubRes{Scid: -1}
	pool := atlas.pools["atlas"]

	if scid, err := db_insert_one[Scrub](ctx, pool, "atlas.scrubs", nil, scr, "scid"); err == nil {
		scr.Scid = scid
		res.Scid = scid
	} else {
		return res, err
	}

	scrb := new_scrub(scr)
	
	atlas.scrubs[scrb.scid] = scrb
	go scrb.run()

	return res, nil
}

func (s *atlasServer) GetScrub(ctx context.Context, req *ScrubIdent) (*Scrub, error) {
	if scrb, ok := atlas.scrubs[req.Scid];ok {
		return scrb.scrb, nil
	}
	return atlas_db_read[Scrub](ctx, "GetScrub", "atlas.scrubs", -1, req.Scid, "")
}

func (s *atlasServer) GetScrubMetrics(ctx context.Context, req *ScrubIdent) (*Metrics, error) {
	if scrb, ok := atlas.scrubs[req.Scid];ok {
		return scrb.metr, nil
	}
	return atlas_db_read[Metrics](ctx, "GetScrubMetrics", "atlas.metrics", -1, req.Scid, "")
}

func (s *atlasServer) GetScrubRebates(req *ScrubIdent, strm grpc.ServerStreamingServer[ScrubRebate]) error {
	if scrb, ok := atlas.scrubs[req.Scid];ok {
		for _, rbt := range scrb.rbts {
			strm.Send(rbt.new_scrub_rebate(scrb))
		}
		return nil
	}
	return atlas_db_read_strm("GetScrubRebates", "atlas.scrub_rebates", strm, -1, req.Scid, "Rbid")
}

func (s *atlasServer) GetScrubFile(req *ScrubIdent, strm grpc.ServerStreamingServer[ScrubRow]) error {
	hdrs := metaGet(strm.Context(), "hdrs")		// Comma-separated list of headers indicating columns to return.
	scid := metaGetI64(strm.Context(), "scid")
	manu := metaGet(strm.Context(), "manu")
	ivid := int64(-1)
	stop := make(chan any)
	pool := atlas.pools["atlas"]
	cols := []*invoice_col{}
	mem  := true
	defer close(stop)

	var scrb *scrub
	var Scrb *Scrub
	var err  error

	// The hdrs metadata contains the list/order of columns to be returned. This list may be from the set of headers on the
	// invoice along with additional data like 340b_id, 340b_status, etc. (could be fields from claims, LU tables, etc - written by policy).
	// Here is where the data come from:
	// invoice_cols: indx, name (0/spid 1/ndc 2/rxn)			headers from invoice
	// scrubs: cust (340bID,340stat,foo,bar)					custom policy-generated headers (attribute names)
	// scrub_rebates: (aaa,bb,cccc,d)							custom policy-generated columns (attribute values)

	// These maps map the header name to its index position in the CS list of values.
	imap := map[string]int{}			// header map from invoice  (header row on invoice).
	smap := map[string]int{}			// header map from standard (standard values we add, like stat, errc, errm, etc).
	cmap := map[string]int{}			// header map from custom   (custom values added by policy, like 340bID, 340b_stat, etc).
	hmap := map[string]map[string]int{}	// maps header name to which map that header's data comes from: "stat"=>smap, "street"=>imap, ...
	hsrc := map[string]string{}			// maps header name to which map (name) that header's data comes from: "stat"=>"smap", "stree"=>"imap", ...

	// First, build the basic smap (not yet looking at requested hdrs).
	smap["stat"] = 0
	smap["excl"] = 1
	smap["errc"] = 2
	smap["errm"] = 3

	// Get the Scrub so that we can get the invoice id and custom headers. Need that to get the column definitions and all original rebate data.
	if scrb = atlas.scrubs[scid]; scrb == nil {
		mem = false
		if Scrb, err = atlas_db_read[Scrub](context.Background(), "GetScrubFile", "atlas.scrubs", -1, scid, ""); err != nil {
			return err
		}
	} else {
		Scrb = scrb.scrb
	}
	ivid = Scrb.Ivid

	// Put the custom headers into cmap (foo=>0, bar=>1, ...)
	if len(Scrb.Cust) > 0 {
		cust := strings.Split(Scrb.Cust, ",")
		for i, hdr := range cust {
			cmap[strings.ToLower(hdr)] = i
		}
	}

	// Get invoice columns
	if chn, err := db_select[invoice_col](pool, "atlas", "atlas.invoice_cols", nil, fmt.Sprintf("ivid = %d", ivid), "indx", stop); err == nil {
		for ic := range chn {
			cols = append(cols, ic)
		}
		// Put the invoice headers into imap
		for i, col := range cols {
			imap[strings.ToLower(col.Name)] = i
		}
	}

	// Finally, go through the requested header list. For each header, map it to its map source.
	insMissing := func(hdrs, val string) string {
		if !strings.Contains(hdrs, val) {
			return val + "," + hdrs
		}
		return hdrs
	}
	addMissing := func(hdrs, val string) string {
		if !strings.Contains(hdrs, val) {
			return hdrs + "," + val
		}
		return hdrs
	}
	hdrs = strings.ToLower(hdrs)			// header list from metadata (requested columns in output)
	if len(hdrs) == 0 {
		hdrs = "rbid,stat,excl,errc,errm"	// If nothing requested, this is the basic minimal list.
	}
	hdrs = insMissing(hdrs, "stat")
	hdrs = addMissing(hdrs, "excl")
	hdrs = addMissing(hdrs, "errc")
	hdrs = addMissing(hdrs, "errm")
	miss := []string{}
	Hdrs := strings.Split(hdrs, ",")
	for _, hdr := range Hdrs {
		if _,ok := smap[hdr]; ok {			// "stat", "excl", "errc", "errm"		- standard headers
			hmap[hdr] = smap
			hsrc[hdr] = "smap"
		} else if _,ok := cmap[hdr]; ok {	// atlas.scrubs.cust ("foo,bar")		- custom n/v pairs added by policy execution
			hmap[hdr] = cmap
			hsrc[hdr] = "cmap"
		} else if _,ok := imap[hdr]; ok {	// atlas.invoice_cols (indx/name, ...)	- headers in the invoice file
			hmap[hdr] = imap
			hsrc[hdr] = "imap"
		} else {
			miss = append(miss, hdr)
		}
	}
	if len(miss) > 0 {
		Log("atlas", "GetScrubFile", "", "missing hdr(s): %v", 0, nil, nil, miss)
	}

	// Now we have a map that maps a header name to a secondary map (foo=>imap, stat=>smap, street=>imap, ...)
	// Each secondary map maps a column name to the index position within that secondary source's data (CSV).
	// smap: stat=>0, errc=>3, ...
	// imap: street=>0, city=>1, ...
	// So to get the data we use the header name on the hmap to get the secondary map, then use it again on the secondary map to get the index position.
	
	srCh := make(chan *ScrubRebate, 5000)
	if mem {
		go func() {									// Reads rebates in memory
			for _, rbt := range scrb.rbts {
				srCh <- rbt.new_scrub_rebate(scrb)
			}
		}()
	} else {
		go func() {									// Reads rebates from scrub_rebates table
			qry := fmt.Sprintf(`
			select 
				sr.rbid, sr.stat, sr.spmt, sr.errc, sr.errm, sr.cust,
				rb.data
				
			from 
				atlas.scrub_rebates sr
				join atlas.rebates sr on (sr.ivid = rb.ivid and sr.rbid = rb.rbid)
			where sr.manu = '%s' and sr.ivid = %d and sr.scid = %d
			order by sr.rbid;
			`, manu, ivid, scid)
			if chn, err := db_select[ScrubRebate](pool, "atlas", "atlas.scrub_rebates", nil, qry, "rbid", stop); err == nil {
				for sr := range chn {
					srCh <- sr
				}
			}
		}()
	}

	// Regardless of whether the scrub/rebates are still in memory or on disk, we're getting a stream of scrub rebates.
	for sr := range srCh {
		rd := strings.Split(sr.Data, ",")	// Rebate data (columns) from the rebate (atlas.rebates.data - rows from the invoice)
		cd := strings.Split(sr.Cust, ",")	// Custom data (columns) from the rebate (atlas.rebates.cust - custom data added by the policy)
		sb := bytes.Buffer{}
		for i, hdr := range Hdrs {			// Whatever was requested in metadata: stat,rxn,spid,city,state,zip,errc,errm,340bid
			val := ""
			if _map := hmap[hdr];_map != nil {	// Which of the three secondary maps holds this header? (smap, imap, or cmap)
				if i, ok := _map[hdr]; ok {		// Whichever... this is the secondary map. (i) is the lookup index into the list of data items
					switch hsrc[hdr] {			// Is this header in "smap", "imap", or "cmap"?
						case "smap":			// smap not a data string - a standard attribute (stat, excl, etc)
							switch hdr {		// A little redundant, but simple. Go look at the header again, and get the data from the field.
							case "stat":
								val = sr.Stat
							case "errc":
								val = sr.Errc
							case "errm":
								val = sr.Errm
							default:
								// header not found
							}
						case "cmap":				// in the scrub_rebates.cust string ("a,b,c")	- added by policy
							val = cd[i]
						case "imap":				// in the rebates.cust string ("a,b,c")			- in the uploaded invoice
							val = rd[i]
						default:
						}
				}
			}
			sb.WriteString(val)
			if i < len(Hdrs)-1 {
				sb.WriteString(",")
			}
		}
		if err := strm.Send(&ScrubRow{Row: sb.String()}); err != nil {
			return err
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

