package main

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Atlas struct {
    opts    *Opts
    titan   TitanClient
    atlas   AtlasServer
    pools   map[string]*pgxpool.Pool
    db_host string
    db_port string
    db_name string
    db_user string
    db_pass string
    ca      CA
    spis    *SPIs
    done    bool
}

var atlas *Atlas

func run_atlas(wg *sync.WaitGroup, opts *Opts, stop chan any) {
    defer wg.Done()

    atlas = &Atlas{atlas: &atlasServer{}, pools: map[string]*pgxpool.Pool{}, spis: newSPIs(), opts: opts}

    atlas.getEnv()
    atlas.connect()
    atlas.load(stop)

    atlas.pools["binary"] = db_pool(atlas.db_host, atlas.db_port, atlas.db_name, atlas.db_user, atlas.db_pass, true)

    if atlas.done {
        return
    }
    atlasWG := &sync.WaitGroup{}
    atlasWG.Add(4)
    go run_datab_ping(atlasWG, stop, 60, "atlas", nil)
    go run_titan_ping(atlasWG, stop, 60, atlas)
    go run_grpc_server(atlasWG, stop, "atlas", srvp, RegisterAtlasServer, atlas.atlas)
    go run_atlas_sync(atlasWG, stop)
    atlasWG.Wait()
}

func run_atlas_sync(wg *sync.WaitGroup, stop chan any) {
    defer wg.Done()
    for {
        select {
        case <-time.After(60*time.Second):
        case <-stop:
            return
        }
    }
}
func run_titan_ping(wg *sync.WaitGroup, stop chan any, intv int, atlas *Atlas) {
    defer wg.Done()
    pingService := func() {
        started := time.Now()
        if _, err := atlas.titan.Ping(context.Background(), &Req{Auth: atlas.opts.auth, Vers: vers}); err == nil {
            log("atlas", "main", "ping to titan service succeeded", time.Since(started), nil)
        } else {
            log("atlas", "main", "ping to titan service failed", time.Since(started), err)
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

func (srv *Atlas) getEnv() {
    setIf(&srv.db_host, "SRV_DB_HOST")
    setIf(&srv.db_port, "SRV_DB_PORT")
    setIf(&srv.db_name, "SRV_DB_NAME")
    setIf(&srv.db_user, "SRV_DB_USER")
    setIf(&srv.db_pass, "SRV_DB_PASS")
    //setIf(&srv.environment, "BIN_ENVR")
}

func (srv *Atlas) load(stop chan any) {
    //srv.ca.clms = new_cache(srv.getClaims(stop, 100000))
    srv.ca.esp1 = new_cache(srv.getESP1(    stop, 100000))
    srv.ca.ents = new_cache(srv.getEntities(stop, 100000))
    srv.ca.ledg = new_cache(srv.getLedger(  stop, 100000))
    srv.ca.ndcs = new_cache(srv.getNDCs(    stop, 100000))
    srv.ca.phms = new_cache(srv.getPharms(  stop, 100000))
    srv.ca.spis = new_cache(srv.getSPIs(    stop, 100000))
    srv.spis.load(srv.ca.spis)
    srv.ca.done = true
}
