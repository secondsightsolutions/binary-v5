package main

import (
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Titan struct {
	titan TitanServer
	pools map[string]*pgxpool.Pool
	opts  *Opts

	cit_db_host string
	cit_db_port string
	cit_db_name string
	cit_db_user string
	cit_db_pass string

	esp_db_host string
	esp_db_port string
	esp_db_name string
	esp_db_user string
	esp_db_pass string

	rbt_db_host string
	rbt_db_port string
	rbt_db_name string
	rbt_db_user string
	rbt_db_pass string

	environment string
}

var titan *Titan

func run_titan(wg *sync.WaitGroup, opts *Opts, stop chan any) {
	defer wg.Done()

	titan = &Titan{titan: &titanServer{}, pools: map[string]*pgxpool.Pool{}, opts: opts}

	titan.getEnv()
	titan.pools["citus"] = db_pool(titan.cit_db_host, titan.cit_db_port, titan.cit_db_name, titan.cit_db_user, titan.cit_db_pass, true)
	titan.pools["titan"] = db_pool(titan.rbt_db_host, titan.rbt_db_port, titan.rbt_db_name, titan.rbt_db_user, titan.rbt_db_pass, true)
	titan.pools["esp"]   = db_pool(titan.esp_db_host, titan.esp_db_port, titan.esp_db_name, titan.esp_db_user, titan.esp_db_pass, true)

	readyWG := &sync.WaitGroup{}
	doneWG  := &sync.WaitGroup{}
	readyWG.Add(1)
	doneWG.Add(2)
	go run_datab_ping(readyWG, doneWG, stop, 60, "titan", titan.pools)
	readyWG.Wait()

	go run_grpc_server(doneWG, stop, "titan", svcp, RegisterTitanServer, titan.titan)
	//go run_save_to_azure(svcWGrp, stop, 5, azac, azky)
	doneWG.Wait()
}

func (svc *Titan) getEnv() {
	setIf(&svc.cit_db_host, "CIT_DB_HOST")
	setIf(&svc.cit_db_port, "CIT_DB_PORT")
	setIf(&svc.cit_db_name, "CIT_DB_NAME")
	setIf(&svc.cit_db_user, "CIT_DB_USER")
	setIf(&svc.cit_db_pass, "CIT_DB_PASS")
	setIf(&svc.esp_db_host, "ESP_DB_HOST")
	setIf(&svc.esp_db_port, "ESP_DB_PORT")
	setIf(&svc.esp_db_name, "ESP_DB_NAME")
	setIf(&svc.esp_db_user, "ESP_DB_USER")
	setIf(&svc.esp_db_pass, "ESP_DB_PASS")
	setIf(&svc.rbt_db_host, "RBT_DB_HOST")
	setIf(&svc.rbt_db_port, "RBT_DB_PORT")
	setIf(&svc.rbt_db_name, "RBT_DB_NAME")
	setIf(&svc.rbt_db_user, "RBT_DB_USER")
	setIf(&svc.rbt_db_pass, "RBT_DB_PASS")
	setIf(&svc.environment, "BIN_ENVR")
}
