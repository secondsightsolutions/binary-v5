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
    svrLocalPool   *pgxpool.Pool
)

type srvServer struct {
    UnimplementedBinaryV5SrvServer
}

func run_grpc_server() chan error {
    cfg := &tls.Config{
        Certificates: []tls.Certificate{TLSCert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    X509pool,
    }
    chn  := make(chan error)
    cred := credentials.NewTLS(cfg)
    server.gsr = grpc.NewServer(grpc.Creds(cred))

    go func() {
        if lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port)); err == nil {
            server.srv = &srvServer{}
            RegisterBinaryV5SrvServer(server.gsr, server.srv)
            fmt.Println("server : grpc server starting")
            if err := server.gsr.Serve(lis); err != nil {
                chn <-err
            }
            close(chn)
        } else {
            chn <-err
            close(chn)
        }
    }()
    return chn
}

func (s *srvServer) Ping(ctx context.Context, in *Req) (*Res, error) {
    return &Res{}, nil
}

func (s *srvServer) Start(ctx context.Context, req *StartReq) (*StartRes, error) {
    getPolicy(req.Manu)
    return nil, nil
}

func (s *srvServer) Scrub(strm grpc.ClientStreamingServer[Rebate, Metrics]) error {
    return nil
}

func (s *srvServer) Done(ctx context.Context, req *DoneReq) (*ScrubRes, error) {
    return nil, nil
}