package main

import (
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func run_datab_ping(done *sync.WaitGroup, stop chan any, appl string, intv int, pools map[string]*pgxpool.Pool) {
	pingDBs := func(pools map[string]*pgxpool.Pool) {
		for name, pool := range pools {
			go ping_db(appl, name, pool)
		}
	}
	Log(appl, "run_datab_ping", "", "starting", 0, nil, nil)
	durn := time.Duration(intv) * time.Second
	done.Add(1)
	go func() {
		defer done.Done()
		for {
			select {
			case <-time.After(durn):
				pingDBs(pools)
				
			case <-stop:
				Log(appl, "run_datab_ping", "", "received stop signal, returning", 0, nil, nil)
				return
			}
		}
	}()
}
