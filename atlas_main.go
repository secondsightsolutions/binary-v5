package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Atlas struct {
	opts     *Opts
	titan    TitanClient
	atlas    AtlasServer
	scrubs   map[int64]*scrub
	pools    map[string]*pgxpool.Pool
	TLSCert  *tls.Certificate
	X509cert *x509.Certificate
	ca       cache_set
	spis     *SPIs
	done     bool
}

var atlas *Atlas

func run_atlas(wg *sync.WaitGroup, opts *Opts, stop chan any) {
	defer wg.Done()

	atlas = &Atlas{atlas: &atlasServer{}, scrubs: map[int64]*scrub{}, pools: map[string]*pgxpool.Pool{}, spis: newSPIs(), opts: opts}

	atlas.pools["atlas"] = db_pool(atlas_host, atlas_port, atlas_name, atlas_user, atlas_pass, true)

	if manu == "" {
		manu = "amgen"
	}
	
	var err error
	if atlas.TLSCert, atlas.X509cert, err = CryptInit(atlas_cert, cacr, "", atlas_pkey, salt, phrs); err != nil {
		log("atlas", "run_atlas", "cannot initialize crypto", 0, err)
		exit(nil, 1, fmt.Sprintf("atlas cannot initialize crypto: %s", err.Error()))
	}

	atlas.connect()
	atlas.load(stop)

	if atlas.done {
		return
	}
	readyWG := &sync.WaitGroup{}
	doneWG  := &sync.WaitGroup{}
	readyWG.Add(4)
	doneWG.Add(5)
	go run_datab_ping(readyWG, doneWG, stop, 60, "atlas", atlas.pools)
	go run_titan_ping(readyWG, doneWG, stop, 60, atlas)
	go run_atlas_sync(readyWG, doneWG, stop, 60, atlas)
    go run_memr_watch(readyWG, doneWG, stop)
	readyWG.Wait()

	go run_grpc_server(doneWG, stop, "atlas", atlas_grpc_port, atlas.TLSCert, RegisterAtlasServer, atlas.atlas)
	doneWG.Wait()
}

func run_atlas_sync(readyWG, doneWG *sync.WaitGroup, stop chan any, intv int, atlas *Atlas) {
	defer doneWG.Done()
	atlas.sync(stop)
	readyWG.Done()
	for {
		select {
		case <-time.After(time.Duration(intv) * time.Second):
			atlas.sync(stop)
		case <-stop:
			log("atlas", "run_atlas_sync", "received stop signal, returning", 0, nil)
			return
		}
	}
}
func run_titan_ping(readyWG, doneWG *sync.WaitGroup, stop chan any, intv int, atlas *Atlas) {
	defer doneWG.Done()
	pingService := func() {
		strt := time.Now()
		if _, err := atlas.titan.Ping(metaGRPC(), &Req{}); err == nil {
			log("atlas", "run_titan_ping", "%-21s", time.Since(strt), err, "ping succeeded")
		} else {
			log("atlas", "run_titan_ping", "%-21s", time.Since(strt), err, "ping failed")
		}
	}
	pingService()
	readyWG.Done()
	durn := time.Duration(intv) * time.Second
	for {
		select {
		case <-time.After(durn):
			pingService()
		case <-stop:
			log("atlas", "run_titan_ping", "received stop signal, returning", 0, nil)
			return
		}
	}
}

func (atlas *Atlas) load(stop chan any) {
	atlas.ca.clms = new_cache("clms", atlas.getClaims(stop))
	atlas.ca.esp1 = new_cache("esp1", atlas.getESP1(stop))
	atlas.ca.ents = new_cache("ents", atlas.getEntities(stop))
	atlas.ca.ledg = new_cache("elig", atlas.getLedger(stop))
	atlas.ca.ndcs = new_cache("ndcs", atlas.getNDCs(stop))
	atlas.ca.phms = new_cache("phms", atlas.getPharms(stop))
	atlas.ca.spis = new_cache("spis", atlas.getSPIs(stop))
	atlas.spis.load(atlas.ca.spis)
	atlas.ca.done = true
}

