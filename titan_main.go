package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Titan struct {
	titan TitanServer
	pools map[string]*pgxpool.Pool
	opts  *Opts
	TLSCert  *tls.Certificate
    X509cert *x509.Certificate
}

var titan *Titan

func run_titan(wg *sync.WaitGroup, opts *Opts, stop chan any) {
	defer wg.Done()

	titan = &Titan{titan: &titanServer{}, pools: map[string]*pgxpool.Pool{}, opts: opts}

	var err error
    if titan.TLSCert, titan.X509cert, err = CryptInit(titan_cert, cacr, "", titan_pkey, salt, phrs); err != nil {
		log("titan", "run_titan", "cannot initialize crypto", 0, err)
		exit(nil, 1, fmt.Sprintf("titan cannot initialize crypto: %s", err.Error()))
	}

	titan.pools["citus"] = db_pool(citus_host, citus_port, citus_name, citus_user, citus_pass, true)
	titan.pools["titan"] = db_pool(titan_host, titan_port, titan_name, titan_user, titan_pass, true)
	titan.pools["esp"]   = db_pool(espdb_host, espdb_port, espdb_name, espdb_user, espdb_pass, true)

	readyWG := &sync.WaitGroup{}
	doneWG  := &sync.WaitGroup{}
	readyWG.Add(1)
	doneWG.Add(2)
	go run_datab_ping(readyWG, doneWG, stop, 60, "titan", titan.pools)
	readyWG.Wait()

	go run_grpc_server(doneWG, stop, "titan", titan_grpc_port, titan.TLSCert, RegisterTitanServer, titan.titan)
	//go run_save_to_azure(svcWGrp, stop, 5, azac, azky)
	doneWG.Wait()
}
