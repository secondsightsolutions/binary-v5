package main

import (
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	srv   BinaryV5SvcServer
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
	
	service = &Service{srv: &binaryV5SvcServer{}, pools: map[string]*pgxpool.Pool{}}
	
	service.getEnv()
	service.pools["citus"]  = db_pool(service.cit_db_host, service.cit_db_port, service.cit_db_name, service.cit_db_user, service.cit_db_pass, "", true)
	service.pools["binary"] = db_pool(service.rbt_db_host, service.rbt_db_port, service.rbt_db_name, service.rbt_db_user, service.rbt_db_pass, "", true)
	service.pools["esp"]    = db_pool(service.esp_db_host, service.esp_db_port, service.esp_db_name, service.esp_db_user, service.esp_db_pass, "", true)
	
	svcWGrp := &sync.WaitGroup{}
	svcWGrp.Add(2)
	go run_database_ping(svcWGrp, stop, 60, service.pools)
	go run_grpc_services(svcWGrp, stop, "service", svcp, RegisterBinaryV5SvcServer, service.srv)
	//go run_save_to_azure(svcWGrp, stop, 5, azac, azky)
	//go run_save_to_datab(svcWGrp, stop, 5, azac, azky, service.pools)
	svcWGrp.Wait()
}

func (svc *Service) getEnv() {
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