package main

import (
	"sync"

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
	go run_grpc_service()
	<-stop
	service.gsr.GracefulStop()
}