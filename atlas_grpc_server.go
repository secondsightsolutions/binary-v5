package main

import (
	context "context"
	"fmt"
	"time"

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
		//Manu: metaGet(strm.Context(), "manu"),
		Plcy: metaGet(strm.Context(), "plcy"),
		Kind: metaGet(strm.Context(), "kind"),
		Name: metaGet(strm.Context(), "name"),
		Vers: metaGet(strm.Context(), "vers"),
		Dscr: metaGet(strm.Context(), "dscr"),
		Hash: metaGet(strm.Context(), "hash"),
		Host: metaGet(strm.Context(), "host"),
		Appl: metaGet(strm.Context(), "appl"),
		Hdrs: metaGet(strm.Context(), "hdrs"),
		Cmdl: metaGet(strm.Context(), "cmdl"),
		Test: metaGet(strm.Context(), "test"),
	}

	// A scrub is done in four steps:
	// 1. create a new scrub in scrubs table from info in metadata.
	// 2. Read in the rebates from the client (if not pre-loaded) and save them to the Rebates table under the new scid.
	// 3. Run the scrub, which will read rebates from table and update them when done.
	// 4. Stream the rebates back down to the client.
	// Note that an option may be to stream back rebates in real time.

	pool := atlas.pools["atlas"]
	strt := time.Now()

	// Create the scrub row from the metadata on the stream.
	if scid, err := db_insert_one[Scrub](strm.Context(), pool, "atlas.scrubs", scr, "scid"); err == nil {
		scr.Scid = scid
	} else {
		return err
	}

	// Create the in-memory "scrub" and scrub the rebates
	stop := make(chan any)
	scrb := new_scrub(scr, stop)
	
	// Now create the two-way connector with the client.
	// The input side are the rebates coming from the client, to be inserted into db.
	// The output side is where/how we send rebates from the rebates table back down to the client.
	
	chnT, chnR := strm_fmto_clnt("atlas", "", strm, stop)

	if scrb.cs.rbts != nil {
		// Insert the rebates from test_rebates into the Rebates table (we won't upload them).
		
	} else {
		// Pull in the rebates from the shell (client) and write them into the rebates table.
		cnt, seq, err := db_insert(pool, "atlas", "atlas.rebates", nil, chnT, 5000, "Rbid", false)
		Log("atlas", "Rebates", "atlas.Rebates", "rebates inserted", time.Since(strt), map[string]any{"cnt": cnt, "seq": seq}, err)
	}
	
	atlas.scrubs[scrb.scid] = scrb
	scrb.run()
	
	// Send the rebates to the shell (client) on the other channel of the stream.
	whr := fmt.Sprintf("scid = %d", scrb.scid)
	if chn, err := db_select[Rebate](pool, "atlas", "atlas.rebates", nil, whr, "Rbid", stop); err == nil {
		for rbt := range chn {
			chnR <-rbt
		}
	}
	close(stop)
	return nil
}

