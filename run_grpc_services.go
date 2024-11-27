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


func run_grpc_server[T any](wg *sync.WaitGroup, stop chan any, name string, port int, cert *tls.Certificate, regis func(grpc.ServiceRegistrar, T), srv T) {
    defer wg.Done()

    cfg := &tls.Config{
        Certificates: []tls.Certificate{*cert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    X509pool,
    }
    var gsr *grpc.Server
    for {
        select {
        case <-time.After(time.Duration(5) * time.Second):
            if gsr == nil {
                cred := credentials.NewTLS(cfg)
                gsr  = grpc.NewServer(grpc.Creds(cred))

                go func() {
                    if lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port)); err == nil {
                        regis(gsr, srv)
                        log(name, "run_grpc_server", "server starting", 0, nil)
                        if err := gsr.Serve(lis); err != nil {
                            log(name, "run_grpc_server", "server cannot Serve()", 0, err)
                            gsr = nil
                        }
                    } else {
                        gsr = nil
                    }
                }()
            }
        case <-stop:
            if gsr != nil {
                log(name, "run_grpc_server", "received stop signal, stopping", 0, nil)
                gsr.GracefulStop()
                log(name, "run_grpc_server", "server stopped", 0, nil)
            }
            log(name, "run_grpc_server", "server returning", 0, nil)
            return
        }
    }
}
