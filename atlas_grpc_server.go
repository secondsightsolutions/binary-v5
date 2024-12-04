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

func (s *atlasServer) NewScrub(ctx context.Context, req *Scrub) (*ScrubRes, error) {
	if err := validate_client(ctx, atlas.pools["atlas"], "atlas"); err != nil {
		return &ScrubRes{}, err
	}
	if scid, err := db_insert_one(ctx, atlas.pools["atlas"], "atlas.scrubs", nil, req, "scid"); err == nil {
		req.Scid = scid
		scrb := new_scrub(req)
		scrb.ca = atlas.ca.clone(req.Data)
		if scrb.ca.spis != atlas.ca.spis {
			scrb.spis = newSPIs()
			scrb.spis.load(scrb.ca.spis)
		}
		atlas.add_scrub(scrb)
		go scrb.run()
		return &ScrubRes{Scid: scid}, nil
	} else {
		return nil, err
	}
}

func (s *atlasServer) Rebates(strm grpc.ClientStreamingServer[Rebate, Res]) error {
	if err := validate_client(strm.Context(), atlas.pools["atlas"], "atlas"); err != nil {
		return err
	}
	for {
		if rbt, err := strm.Recv(); err == nil {
			fmt.Printf("atlas: %v\n", rbt)
		} else {
			break
		}
	}
	return nil
}

