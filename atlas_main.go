package main

import (
	context "context"
	"crypto/tls"
	"crypto/x509"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Atlas struct {
	opts     *Opts
	titan    TitanClient
	atlas    AtlasServer
	claims   *gcache
	scrubs   map[int64]*scrub
	pools    map[string]*pgxpool.Pool
	TLSCert  *tls.Certificate
	X509cert *x509.Certificate
	ca       cache_set
	spis     *SPIs
}

var atlas *Atlas

func run_atlas(done *sync.WaitGroup, opts *Opts, stop chan any) {
	atlas = &Atlas{
		opts:   opts,
		atlas:  &atlasServer{},
		claims: new_gcache(),
		scrubs: map[int64]*scrub{}, 
		pools:  map[string]*pgxpool.Pool{}, 
		spis:   new_spis(), 
	}

	atlas.X509cert, atlas.TLSCert = crypt_init("atlas", "run_atlas", 32, atlas_cert, cacr, "", atlas_pkey)
	atlas.pools["atlas"] = db_pool("atlas", atlas_host, atlas_port, atlas_name, atlas_user, atlas_pass, true)
	atlas.titan = grpc_connect[TitanClient](titan_grpc, titan_grpc_port, atlas.TLSCert, NewTitanClient)
	atlas.load(stop)

	run_datab_ping( done, stop, "atlas", 60, atlas.pools)
	run_titan_ping( done, stop, "atlas", 60, atlas)
	run_atlas_sync( done, stop, "atlas", 60, atlas)
    run_memr_watch( done, stop)
	run_grpc_server(done, stop, "atlas", atlas_grpc_port, atlas.TLSCert, RegisterAtlasServer, atlas.atlas, atlasUnaryServerInterceptor, atlasStreamServerInterceptor)
}

func run_atlas_sync(done *sync.WaitGroup, stop chan any, appl string, intv int, atlas *Atlas) {
	durn := time.Duration(0)
	done.Add(1)
	go func() {
		defer done.Done()
		for {
			select {
			case <-time.After(durn):
				atlas.sync(stop)
				durn = time.Duration(intv) * time.Second
			case <-stop:
				Log(appl, "run_atlas_sync", "", "received stop signal, returning", 0, nil, nil)
				return
			}
		}
	}()
}
func run_titan_ping(done *sync.WaitGroup, stop chan any, appl string, intv int, atlas *Atlas) {
	pingService := func() {
		strt  := time.Now()
		_,err := atlas.titan.Ping(addMeta(context.Background(), atlas.X509cert, nil), &Req{})
		Log("atlas", "run_titan_ping", "titan", "ping completed", time.Since(strt), nil, err)
	}
	durn := time.Duration(intv) * time.Second
	done.Add(1)
	go func() {
		defer done.Done()
		for {
			select {
			case <-time.After(durn):
				pingService()
			case <-stop:
				Log(appl, "run_titan_ping", "", "received stop signal, returning", 0, nil, nil)
				return
			}
		}
	}()
}

func (atlas *Atlas) load(stop chan any) {
	strt := time.Now()
	done := &sync.WaitGroup{}
	Log("atlas", "load", "all caches", "starting", time.Since(strt), nil, nil)
	done.Add(7)
	load_gclms(stop, done)
	load_cache(stop, done, &atlas.ca.esp1, "esp1", atlas.getESP1)
	load_cache(stop, done, &atlas.ca.ents, "ents", atlas.getEntities)
	load_cache(stop, done, &atlas.ca.ledg, "elig", atlas.getLedger)
	load_cache(stop, done, &atlas.ca.ndcs, "ndcs", atlas.getNDCs)
	load_cache(stop, done, &atlas.ca.phms, "phms", atlas.getPharms)
	load_cache(stop, done, &atlas.ca.spis, "spis", atlas.getSPIs)
	done.Wait()
	atlas.spis.load(atlas.ca.spis)
	atlas.ca.done = true
	Log("atlas", "load", "all caches", "completed", time.Since(strt), nil, nil)
}

func (atlas *Atlas) sync(stop chan any) {
	pool := atlas.pools["atlas"]
	sync_fm_server(pool, "atlas", "atlas.claims",        		false,				atlas.X509cert, atlas.titan.GetClaims,    			stop)
	sync_fm_server(pool, "atlas", "atlas.auth",          		true,				atlas.X509cert, atlas.titan.GetAuths,     			stop)
	sync_to_server(pool, "atlas", "atlas.commands",             "commands",			atlas.X509cert, atlas.titan.SyncCommands,           stop)
	sync_to_server(pool, "atlas", "atlas.scrubs",        		"scrubs",			atlas.X509cert, atlas.titan.SyncScrubs,       		stop)
	sync_to_server(pool, "atlas", "atlas.scrub_rebates",       	"scrub_rebates",	atlas.X509cert, atlas.titan.SyncScrubRebates,   	stop)
	sync_to_server(pool, "atlas", "atlas.scrub_claims",    		"scrub_claims",		atlas.X509cert, atlas.titan.SyncScrubClaims,   		stop)
	sync_to_server(pool, "atlas", "atlas.scrub_rebates_claims",	"scrub_reb_clms",	atlas.X509cert, atlas.titan.SyncScrubRebatesClaims, stop)
	sync_to_server(pool, "atlas", "atlas.metrics", 				"metrics",			atlas.X509cert, atlas.titan.SyncMetrics,			stop)
}
