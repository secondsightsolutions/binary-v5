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
	return &Res{}, nil
}

func (s *atlasServer) NewScrub(ctx context.Context, req *Scrub) (*ScrubRes, error) {

	return nil, nil
}

func (s *atlasServer) Rebates(strm grpc.ClientStreamingServer[Rebate, Res]) error {
	for {
		if rbt, err := strm.Recv(); err == nil {
			fmt.Printf("atlas: %v\n", rbt)
		} else {
			break
		}
	}
	return nil
}

