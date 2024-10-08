package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
)

type Server struct {
	ca   CA
	spis *SPIs
	svc  BinaryV5SvcClient
	srv  BinaryV5SrvServer
	gsr  *grpc.Server
}
var server *Server

func server_main(wg *sync.WaitGroup, stop chan any) {
	defer wg.Done()
	server   = &Server{spis: newSPIs()}
	interval := time.Duration(0)
	stopping := false
	for {
		select {
		case <-time.After(interval):
			interval = time.Duration(60) * time.Second
			server.connect()
			if _, err := server.svc.Ping(context.Background(), &Req{Auth: auth, Ver: vers}); err == nil {
				log("server", "main", "ping to rebate service succeeded")
				if !server.ca.done {
					server.load()
				}
			} else {
				log("server", "main", "ping to rebate service failed: %s", err.Error())
			}

		case err := <-run_grpc_server():
			if !stopping {
				log("server", "main", "grpc failure: %s", err.Error())
			} else {
				log("server", "main", "grpc completed, returning")
				return
			}
			
		case <-stop:
			log("server", "main", "stop requested, shutting down grpc")
			stopping = true
			server.gsr.GracefulStop()
			log("server", "main", "grpc completed, returning")
			return
		}
	}
}

func (srv *Server) load() {
	srv.ca.clms = new_cache(srv.getClaims())
	srv.ca.esp1 = new_cache(srv.getESP1PharmNDCs())
	srv.ca.ents = new_cache(srv.getEntities())
	srv.ca.ledg = new_cache(srv.getEligibilityLedger())
	srv.ca.ndcs = new_cache(srv.getNDCs())
	srv.ca.phms = new_cache(srv.getPharmacies())
	srv.ca.spis = new_cache(srv.getSPIs())
	srv.spis.load(srv.ca.spis)
	srv.ca.done = true
}

