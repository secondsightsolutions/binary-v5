package main

import (
	context "context"
	"crypto/tls"
	"fmt"
	"io"
	"os"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)


func (srv *Server) connect() {
	target := fmt.Sprintf("%s:%s", host, port)
	if srv.svc != nil {
		return
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{TLSCert},
		RootCAs:      X509pool,
	}
	crd := credentials.NewTLS(cfg)
	if conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(crd)); err == nil {
		srv.svc = NewBinaryV5SvcClient(conn)
	} else {
		fmt.Printf("failed to connect to grpc server at %s:%s: %s\n", host, port, err.Error())
		os.Exit(27)
	}
}

func ping() {
}
func server_load[T any](name string, f func(context.Context, *Req, ...grpc.CallOption)(grpc.ServerStreamingClient[T], error)) []any {
	log := func(src string, cnt int) {
		fmt.Printf("%-15s: loaded %d rows\n", src, cnt)
	}
	list := make([]any, 0)
	req := &Req{Auth: auth, Ver: vers, Manu: manu}
	if strm, err := f(context.Background(), req); err == nil {
		for {
			if obj, err := strm.Recv(); err == nil {
				list = append(list, obj)
				if len(list)%25000 == 0 {
					log(name, len(list))
				}
			} else if err == io.EOF {
				strm.CloseSend()
				break
			} else {
				strm.CloseSend()
				break
			}
		}
		if len(list)%25000 != 0 {
			log(name, len(list))
		}
	}
	return list
}

func (srv *Server) getClaims() []any {
	return server_load[Claim]("claims", srv.svc.GetClaims)
}
func (srv *Server) getESP1PharmNDCs() []any {
	return server_load[ESP1PharmNDC]("esp1", srv.svc.GetESP1Pharms)
}
func (srv *Server) getEntities() []any {
	return server_load[Entity]("entities", srv.svc.GetEntities)
}
func (srv *Server) getEligibilityLedger() []any {
	return server_load[Eligibility]("ledger", srv.svc.GetEligibilityLedger)
}
func (srv *Server) getNDCs() []any {
	return server_load[NDC]("ndcs", srv.svc.GetNDCs)
}
func (srv *Server) getPharmacies() []any {
	return server_load[Pharmacy]("pharms", srv.svc.GetPharmacies)
}
func (srv *Server) getSPIs() []any {
	return server_load[SPI]("esp1", srv.svc.GetSPIs)
}

