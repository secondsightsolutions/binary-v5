package main

import (
	context "context"
	"fmt"

	grpc "google.golang.org/grpc"
)

type atlasServer struct {
	UnimplementedAtlasServer
}

func (s *atlasServer) Ping(ctx context.Context, in *Req) (*Res, error) {
	if err := validate_client(ctx, atlas.pools["atlas"], "atlas"); err != nil {
		return &Res{}, err
	}
	return &Res{}, nil
}

func (s *atlasServer) Rebates(strm grpc.BidiStreamingServer[Rebate, Rebate]) error {
	if err := validate_client(strm.Context(), atlas.pools["atlas"], "atlas"); err != nil {
		return err
	}
	scr := &Scrub{
		Auth: metaGet(strm.Context(), "auth"),
		Manu: metaGet(strm.Context(), "manu"),
		Plcy: metaGet(strm.Context(), "plcy"),
		Kind: metaGet(strm.Context(), "kind"),
		Name: metaGet(strm.Context(), "name"),
		Vers: metaGet(strm.Context(), "vers"),
		Desc: metaGet(strm.Context(), "desc"),
		Hash: metaGet(strm.Context(), "hash"),
		Host: metaGet(strm.Context(), "host"),
		Appl: metaGet(strm.Context(), "appl"),
		Hdrs: metaGet(strm.Context(), "hdrs"),
		Cmdl: metaGet(strm.Context(), "cmdl"),
		Test: metaGet(strm.Context(), "test"),
	}
	// Create the scrub row from the metadata on the stream.
	if scid, err := db_insert_one(strm.Context(), atlas.pools["atlas"], "atlas.scrubs", scr, "scid"); err == nil {
		scr.Scid = scid
	} else {
		fmt.Printf("yes, NewScrub() failed with %s\n", err.Error())
		return err
	}
	stop := make(chan any)
	chnT, chnR := strm_fmto_clnt("atlas", "", strm, stop)

	// Pull in the rebates from the shell (client) and write them into the rebates table.
	if cnt, seq, err := db_insert(atlas.pools["atlas"], "atlas", "atlas.rebates", nil, chnT, 5000, false); err == nil {
		log_sync(appl, "Rebates", "atlas.Rebates", manu, "", 0, cnt, seq, err, 0)
	} else {
		log_sync(appl, "Rebates", "atlas.Rebates", manu, "db_insert failed", 0, cnt, seq, err, 0)
	}

	// Create the in-memory "scrub" and scrub the rebates
	scrb := new_scrub(scr)
	scrb.ca = atlas.ca.clone()
	if scrb.ca.spis != atlas.ca.spis {
		scrb.spis = newSPIs()
		scrb.spis.load(scrb.ca.spis)
	}
	atlas.add_scrub(scrb)
	go scrb.run()
	
	// Send the rebates to the shell (client) on the other channel of the stream.
	whr := fmt.Sprintf("scid = %d", scrb.scid)
	if chn, err := db_select[Rebate](atlas.pools["atlas"], "atlas.rebates", nil, whr, "sort?", stop); err == nil {
		for rbt := range chn {
			chnR <-rbt
		}
	}
	close(stop)
	return nil
}

