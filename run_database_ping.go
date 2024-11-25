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
			go pingDB(appl, name, pool)
		}
	}
	durn := time.Duration(0) * time.Second
	for {
		select {
		case <-time.After(durn):
			pingDBs(pools)
			if durn == 0 {
				readyWG.Done()
			}
			durn = time.Duration(intv) * time.Second
		case <-stop:
			log(appl, "run_datab_ping", "received stop signal, returning", 0, nil)
			return
		}
	}
}
