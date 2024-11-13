package main

import (
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

    manu = "teva"

    atlas.getEnv()
    atlas.connect()
    atlas.load(stop)

    atlas.pools["atlas"] = db_pool(atlas.db_host, atlas.db_port, atlas.db_name, atlas.db_user, atlas.db_pass, true)

    if atlas.done {
        return
    }
    readyWG := &sync.WaitGroup{}
    doneWG  := &sync.WaitGroup{}
    readyWG.Add(3)
    doneWG.Add(4)
    go run_datab_ping(readyWG, doneWG, stop, 60, "atlas", nil)
    go run_titan_ping(readyWG, doneWG, stop, 60, atlas)
    go run_atlas_sync(readyWG, doneWG, stop, 60, atlas)
    readyWG.Wait()

    go run_grpc_server(doneWG, stop, "atlas", srvp, RegisterAtlasServer, atlas.atlas)
    doneWG.Wait()
}

func run_atlas_sync(readyWG, doneWG *sync.WaitGroup, stop chan any, intv int, atlas *Atlas) {
    defer doneWG.Done()
    atlas.sync()
    readyWG.Done()
    for {
        select {
        case <-time.After(time.Duration(intv)*time.Second):
            atlas.sync()
        case <-stop:
            log("atlas", "main", "titan sync returning", 0, nil)
            return
        }
    }
}
func run_titan_ping(readyWG, doneWG *sync.WaitGroup, stop chan any, intv int, atlas *Atlas) {
    defer doneWG.Done()
    pingService := func() {
        started := time.Now()
        if _, err := atlas.titan.Ping(metaGRPC(), &Req{}); err == nil {
            log("atlas", "main", "ping to titan service succeeded", time.Since(started), nil)
        } else {
            log("atlas", "main", "ping to titan service failed", time.Since(started), err)
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
            return
        }
    }
}

func (atlas *Atlas) getEnv() {
    setIf(&atlas.db_host, "ATLAS_DB_HOST")
    setIf(&atlas.db_port, "ATLAS_DB_PORT")
    setIf(&atlas.db_name, "ATLAS_DB_NAME")
    setIf(&atlas.db_user, "ATLAS_DB_USER")
    setIf(&atlas.db_pass, "ATLAS_DB_PASS")
    //setIf(&atlas.environment, "BIN_ENVR")
}

func (atlas *Atlas) load(stop chan any) {
    //atlas.ca.clms = new_cache(atlas.getClaims(stop))
    atlas.ca.esp1 = new_cache(atlas.getESP1(    stop))
    atlas.ca.ents = new_cache(atlas.getEntities(stop))
    atlas.ca.ledg = new_cache(atlas.getLedger(  stop))
    atlas.ca.ndcs = new_cache(atlas.getNDCs(    stop))
    atlas.ca.phms = new_cache(atlas.getPharms(  stop))
    atlas.ca.spis = new_cache(atlas.getSPIs(    stop))
    atlas.spis.load(atlas.ca.spis)
    atlas.ca.done = true
}
