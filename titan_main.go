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

func run_titan(done *sync.WaitGroup, opts *Opts, stop chan any) {
	titan = &Titan{titan: &titanServer{}, pools: map[string]*pgxpool.Pool{}, opts: opts}

	var err error
    if titan.TLSCert, titan.X509cert, err = CryptInit(titan_cert, cacr, "", titan_pkey, salt, phrs); err != nil {
		Log("titan", "run_titan", "", "cannot initialize crypto", 0, nil, err)
		exit(nil, 1, fmt.Sprintf("titan cannot initialize crypto: %s", err.Error()))
	}

	titan.pools["citus"] = db_pool("titan", citus_host, citus_port, citus_name, citus_user, citus_pass, true)
	titan.pools["titan"] = db_pool("titan", titan_host, titan_port, titan_name, titan_user, titan_pass, true)
	titan.pools["esp"]   = db_pool("titan", espdb_host, espdb_port, espdb_name, espdb_user, espdb_pass, true)

	run_datab_ping( done, stop, "titan", 60, titan.pools)
	run_grpc_server(done, stop, "titan", titan_grpc_port, titan.TLSCert, RegisterTitanServer, titan.titan)
}
