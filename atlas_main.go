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
	dates    *Dates
	TLSCert  *tls.Certificate
	X509cert *x509.Certificate
	ca       cache_set
	spis     *SPIs
	metrics  metrics
}

var atlas *Atlas

func run_atlas(done *sync.WaitGroup, opts *Opts) {
	atlas = &Atlas{
		opts:   opts,
		atlas:  &atlasServer{},
		claims: new_gcache(),
		scrubs: map[int64]*scrub{}, 
		pools:  map[string]*pgxpool.Pool{}, 
		spis:   new_spis(), 
	}

	atlas.X509cert, atlas.TLSCert = crypt_init("atlas", "run_atlas", 32, atlas_cert, cacr, "", atlas_pkey)
	atlas.dates = new_dates(10)
	atlas.pools["atlas"] = db_pool("atlas", atlas_host, atlas_port, atlas_name, atlas_user, atlas_pass, true)
	atlas.titan = grpc_connect[TitanClient](titan_grpc, titan_grpc_port, atlas.TLSCert, NewTitanClient)
	atlas.load()

	run_datab_ping( done, "atlas", 60, atlas.pools)
	run_titan_ping( done, "atlas", 60, atlas)
	run_atlas_sync( done, "atlas", 60, atlas)
    run_memr_watch( done)
	run_grpc_server(done, "atlas", atlas_grpc_port, atlas.TLSCert, RegisterAtlasServer, atlas.atlas, atlasUnaryServerInterceptor, atlasStreamServerInterceptor)
	run_data_writer(done, "atlas", 5, atlas)
}

func run_data_writer(done *sync.WaitGroup, appl string, intv int, atlas *Atlas) {
	done.Add(1)
	go func() {
		defer done.Done()
		stopping := false
		for {
			select {
			case <-time.After(time.Duration(intv) * time.Second):
				for scid, scrb := range atlas.scrubs {
					if scrb.done {
						scrb.save_rebates(5000)
						delete(atlas.scrubs, scid)
					}
				}
				if stopping {
					Log(appl, "run_data_writer", "", "received stop signal, returning", 0, nil, nil)
					return
				}

			case <-stop:
				Log(appl, "run_data_writer", "", "received stop signal, finishing up", 0, nil, nil)
				stopping = true
				stop = nil
			}
		}
	}()
}

func run_atlas_sync(done *sync.WaitGroup, appl string, intv int, atlas *Atlas) {
	durn := time.Duration(0)
	done.Add(1)
	go func() {
		defer done.Done()
		for {
			select {
			case <-time.After(durn):
				atlas.sync()
				durn = time.Duration(intv) * time.Second
			case <-stop:
				Log(appl, "run_atlas_sync", "", "received stop signal, returning", 0, nil, nil)
				return
			}
		}
	}()
}
func run_titan_ping(done *sync.WaitGroup, appl string, intv int, atlas *Atlas) {
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

func (atlas *Atlas) load() {
	strt := time.Now()
	Log("atlas", "load", "all caches", "starting", time.Since(strt), nil, nil)
	atlas.sync()
	atlas.spis.load(atlas.ca.spis, &atlas.metrics.init_spis)
	Log("atlas", "load", "all caches", "completed", time.Since(strt), nil, nil)
}

func (atlas *Atlas) sync() {
	strt := time.Now()
	Log("atlas", "sync", "sync to/fm titan", "starting", time.Since(strt), nil, nil)
	pool := atlas.pools["atlas"]
	sync_fm_server(pool, "atlas", "atlas.claims",        	false,	true,		atlas.X509cert, atlas.titan.GetClaims)
	sync_fm_server(pool, "atlas", "atlas.auth",          	true,	false,		atlas.X509cert, atlas.titan.GetAuths)
	sync_to_server(pool, "atlas", "atlas.commands",         "commands",			atlas.X509cert, atlas.titan.SyncCommands)
	sync_to_server(pool, "atlas", "atlas.scrubs",        	"scrubs",			atlas.X509cert, atlas.titan.SyncScrubs)
	sync_to_server(pool, "atlas", "atlas.scrub_rebates",	"scrub_rebates",	atlas.X509cert, atlas.titan.SyncScrubRebates)
	sync_to_server(pool, "atlas", "atlas.scrub_claims",    	"scrub_claims",		atlas.X509cert, atlas.titan.SyncScrubClaims)
	sync_to_server(pool, "atlas", "atlas.scrub_matches",	"scrub_matches",	atlas.X509cert, atlas.titan.SyncScrubMatches)
	sync_to_server(pool, "atlas", "atlas.scrub_attempts",	"scrub_attempts",	atlas.X509cert, atlas.titan.SyncScrubAttempts)
	sync_to_server(pool, "atlas", "atlas.metrics", 			"metrics",			atlas.X509cert, atlas.titan.SyncMetrics)
	Log("atlas", "sync", "sync to/fm titan", "completed", time.Since(strt), nil, nil)

	strt = time.Now()
	Log("atlas", "sync", "load/update caches", "starting", time.Since(strt), nil, nil)
	wgrp := &sync.WaitGroup{}
	wgrp.Add(9)
	load_gclms(wgrp, atlas.claims, 			 &atlas.metrics.load_claims)
	load_cache(wgrp, &atlas.ca.esp1, "esp1", &atlas.metrics.load_esp1, 		atlas.getESP1)
	load_cache(wgrp, &atlas.ca.ents, "ents", &atlas.metrics.load_entities, 	atlas.getEntities)
	load_cache(wgrp, &atlas.ca.ledg, "elig", &atlas.metrics.load_ledger, 	atlas.getLedger)
	load_cache(wgrp, &atlas.ca.ndcs, "ndcs", &atlas.metrics.load_ndcs, 		atlas.getNDCs)
	load_cache(wgrp, &atlas.ca.phms, "phms", &atlas.metrics.load_pharms, 	atlas.getPharms)
	load_cache(wgrp, &atlas.ca.spis, "spis", &atlas.metrics.load_spis, 		atlas.getSPIs)
	load_cache(wgrp, &atlas.ca.desg, "desg", &atlas.metrics.load_desg, 		atlas.getDesignations)
	load_cache(wgrp, &atlas.ca.ldns, "ldns", &atlas.metrics.load_ldns, 		atlas.getLDNs)
	wgrp.Wait()
	atlas.metrics.load_data = time.Since(strt)
	Log("atlas", "sync", "load/update caches", "completed", time.Since(strt), nil, nil)
}
