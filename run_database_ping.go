package main

import (
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func run_datab_ping(readyWG, doneWG *sync.WaitGroup, stop chan any, intv int, appl string, pools map[string]*pgxpool.Pool) {
	defer doneWG.Done()
	pingDBs := func(pools map[string]*pgxpool.Pool) {
		for name, pool := range pools {
			pingDB(appl, name, pool)
		}
	}
	pingDBs(pools)
	readyWG.Done()
	durn := time.Duration(intv) * time.Second
	for {
		select {
		case <-time.After(durn):
			pingDBs(pools)
		case <-stop:
			return
		}
	}
}
