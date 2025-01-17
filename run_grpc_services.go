package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)


func run_grpc_server[T any](done *sync.WaitGroup, stop chan any, name string, port int, cert *tls.Certificate, regis func(grpc.ServiceRegistrar, T), srv T, ui grpc.UnaryServerInterceptor, si grpc.StreamServerInterceptor) {
    cfg := &tls.Config{
        Certificates: []tls.Certificate{*cert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    X509pool,
    }
    done.Add(1)
    go func() {
        defer done.Done()
        var gsr *grpc.Server
        for {
            select {
            case <-time.After(time.Duration(5) * time.Second):
                if gsr == nil {
                    cred := credentials.NewTLS(cfg)
                    gsr  = grpc.NewServer(grpc.Creds(cred), grpc.UnaryInterceptor(ui), grpc.StreamInterceptor(si))
                    go func() {
                        if lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port)); err == nil {
                            regis(gsr, srv)
                            Log(name, "run_grpc_server", "", "server starting", 0, nil, nil)
                            if err := gsr.Serve(lis); err != nil {
                                Log(name, "run_grpc_server", "", "server cannot Serve()", 0, nil, err)
                                gsr = nil
                            }
                        } else {
                            gsr = nil
                        }
                    }()
                }
            case <-stop:
                Log(name, "run_grpc_server", "", "received stop signal, stopping", 0, nil, nil)
                if gsr != nil {
                    Log(name, "run_grpc_server", "", "stopping endpoint", 0, nil, nil)
                    gsr.GracefulStop()
                    Log(name, "run_grpc_server", "", "endpoint stopped", 0, nil, nil)
                }
                Log(name, "run_grpc_server", "", "server returning", 0, nil, nil)
                return
            }
        }
    }()
}
