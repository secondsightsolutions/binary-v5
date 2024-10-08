package main

import (
	"sync"
	"time"

	grpc "google.golang.org/grpc"
)


type Service struct {
	srv  BinaryV5SvcServer
	gsr  *grpc.Server
}
var service *Service

func service_main(wg *sync.WaitGroup, stop chan any) {
	defer wg.Done()
	service = &Service{}
	interval := time.Duration(0)
	stopping := false
	for {
		select {
		case <-time.After(interval):
			interval = time.Duration(60) * time.Second
			pingDB("service", "citus",   svcCitusPool)
			pingDB("service", "central", svcCentralPool)

		case err := <-run_grpc_service():
			if !stopping {
				log("service", "main", "grpc failure: %s", err.Error())
			} else {
				log("service", "main", "grpc completed, returning")
				return
			}
			
		case <-stop:
			log("service", "main", "stop requested, shutting down grpc")
			stopping = true
			server.gsr.GracefulStop()
			log("service", "main", "grpc completed, returning")
			return
		}
	}
}