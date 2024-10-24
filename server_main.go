package main

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
    svc  BinaryV5SvcClient
	srv  BinaryV5SrvServer
    pools map[string]*pgxpool.Pool
	db_host	string
    db_port string
    db_name string
    db_user string
    db_pass string
	ca   CA
	spis *SPIs
	done bool
}
var server *Server

func server_main(wg *sync.WaitGroup, stop chan any) {
	defer wg.Done()

	server = &Server{srv: &binaryV5SrvServer{}, pools: map[string]*pgxpool.Pool{}, spis: newSPIs()}
    	
    server.getEnv()
	server.connect()
    server.load(stop)

	server.pools["binary"] = db_pool(server.db_host, server.db_port, server.db_name, server.db_user, server.db_pass, true)

	if server.done {
		return
	}
	srvWGrp := &sync.WaitGroup{}
	srvWGrp.Add(3)
	go run_database_ping(srvWGrp, stop, 60, nil)
	go run_services_ping(srvWGrp, stop, 60, server)
    go run_grpc_services(srvWGrp, stop, "server", srvp, RegisterBinaryV5SrvServer, server.srv)
	srvWGrp.Wait()
}

func run_services_ping(wg *sync.WaitGroup, stop chan any, intv int, server *Server) {
	defer wg.Done()
	pingService := func() {
		started := time.Now()
		if _, err := server.svc.Ping(context.Background(), &Req{Auth: auth, Ver: vers}); err == nil {
			log("server", "main", "ping to rebate service succeeded", time.Since(started), nil)
		} else {
			log("server", "main", "ping to rebate service failed: %s", time.Since(started), err)
		}
	}
	pingService()
	durn := time.Duration(intv) * time.Second
	for {
		select {
		case <-time.After(durn):
			pingService()
		case <-stop:
			return
		}
	}
}

func (srv* Server) getEnv() {
	setIf(&srv.db_host, "SRV_DB_HOST")
	setIf(&srv.db_port, "SRV_DB_PORT")
	setIf(&srv.db_name, "SRV_DB_NAME")
	setIf(&srv.db_user, "SRV_DB_USER")
	setIf(&srv.db_pass, "SRV_DB_PASS")
	//setIf(&srv.environment, "BIN_ENVR")
}

func (srv *Server) load(stop chan any) {
	srv.ca.clms = new_cache(srv.getClaims(	stop, 100000, 5))
	srv.ca.esp1 = new_cache(srv.getESP1(	stop, 100000, 5))
	srv.ca.ents = new_cache(srv.getEntities(stop, 100000, 5))
	srv.ca.ledg = new_cache(srv.getLedger(	stop, 100000, 5))
	srv.ca.ndcs = new_cache(srv.getNDCs(	stop, 100000, 5))
	srv.ca.phms = new_cache(srv.getPharms(	stop, 100000, 5))
	srv.ca.spis = new_cache(srv.getSPIs(	stop, 100000, 5))
	srv.spis.load(srv.ca.spis)
	srv.ca.done = true
}

