package main

import (
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func run_database_ping(wg *sync.WaitGroup, stop chan any, intv int, pools map[string]*pgxpool.Pool) {
	defer wg.Done()
	pingDBs := func(pools map[string]*pgxpool.Pool) {
		for name, pool := range pools {
			pingDB("service", name, pool)
		}
	}
	pingDBs(pools)
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
