package main

import (
	context "context"
	
	grpc "google.golang.org/grpc"
)

type binaryV5SrvServer struct {
    UnimplementedBinaryV5SrvServer
}


func (s *binaryV5SrvServer) Ping(ctx context.Context, in *Req) (*Res, error) {
    return &Res{}, nil
}

func (s *binaryV5SrvServer) Start(ctx context.Context, req *StartReq) (*StartRes, error) {
    
    return nil, nil
}

func (s *binaryV5SrvServer) Scrub(strm grpc.ClientStreamingServer[Rebate, Metrics]) error {
    return nil
}

func (s *binaryV5SrvServer) Done(ctx context.Context, req *DoneReq) (*ScrubRes, error) {
    return nil, nil
}