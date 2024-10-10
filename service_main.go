package main

import (
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	grpc "google.golang.org/grpc"
)


type Service struct {
	srv   BinaryV5SvcServer
	gsr   *grpc.Server
	pools map[string]*pgxpool.Pool
	cit_db_host	string
    cit_db_port string
    cit_db_name string
    cit_db_user string
    cit_db_pass string

    esp_db_host	string
    esp_db_port string
    esp_db_name string
    esp_db_user string
    esp_db_pass string

    rbt_db_host	string
    rbt_db_port string
    rbt_db_name string
    rbt_db_user string
    rbt_db_pass string

	environment string
}
var service *Service

func service_main(wg *sync.WaitGroup, stop chan any) {
	defer wg.Done()
	
	service = &Service{pools: map[string]*pgxpool.Pool{}}
	
	service.pools["citus"]  = db_pool(service.cit_db_host, service.cit_db_port, service.cit_db_name, service.cit_db_user, service.cit_db_pass, "", true)
	service.pools["binary"] = db_pool(service.rbt_db_host, service.rbt_db_port, service.rbt_db_name, service.rbt_db_user, service.rbt_db_pass, "", true)
	service.pools["esp"]    = db_pool(service.esp_db_host, service.esp_db_port, service.esp_db_name, service.esp_db_user, service.esp_db_pass, "", true)
	
	svcWGrp := &sync.WaitGroup{}
	svcWGrp.Add(4)
	go run_database_ping(svcWGrp, 60, service.pools, stop)
	go run_grpc_services(svcWGrp, stop)
	go run_save_to_azure(svcWGrp, 5, azac, azky, stop)
	go run_save_to_datab(svcWGrp, 5, azac, azky, service.pools, stop)
	svcWGrp.Wait()
}

func run_database_ping(wg *sync.WaitGroup, intv int, pools map[string]*pgxpool.Pool, stop chan any) {
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

func (s *Service) getEnv() {
	setIf(&s.cit_db_host, "CIT_DB_HOST")
	setIf(&s.cit_db_name, "CIT_DB_PORT")
	setIf(&s.cit_db_pass, "CIT_DB_NAME")
	setIf(&s.cit_db_port, "CIT_DB_USER")
	setIf(&s.cit_db_user, "CIT_DB_PASS")
	setIf(&s.esp_db_host, "ESP_DB_HOST")
	setIf(&s.esp_db_name, "ESP_DB_PORT")
	setIf(&s.esp_db_pass, "ESP_DB_NAME")
	setIf(&s.esp_db_port, "ESP_DB_USER")
	setIf(&s.esp_db_user, "ESP_DB_PASS")
	setIf(&s.rbt_db_host, "RBT_DB_HOST")
	setIf(&s.rbt_db_name, "RBT_DB_PORT")
	setIf(&s.rbt_db_pass, "RBT_DB_NAME")
	setIf(&s.rbt_db_port, "RBT_DB_USER")
	setIf(&s.rbt_db_user, "RBT_DB_PASS")
	setIf(&s.environment, "BIN_ENVR")
}