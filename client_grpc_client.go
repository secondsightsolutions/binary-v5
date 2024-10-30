package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func (clt *Shell) connect() {
	tgt := fmt.Sprintf("%s:%d", srvh, srvp)
	cfg := &tls.Config{
		Certificates: []tls.Certificate{TLSCert},
		RootCAs:      X509pool,
	}
	crd := credentials.NewTLS(cfg)
	if conn, err := grpc.NewClient(tgt, grpc.WithTransportCredentials(crd)); err == nil {
		clt.atlas = NewAtlasClient(conn)
	} else {
		log("shell", "connect", "cannot connect to atlas", 0, err)
	}
}

func (clt *Shell) newScrub() error {
	md := metadata.New(map[string]string{"auth": clt.opts.auth, "vers": vers, "manu": manu})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	req := &Scrub{
		Auth: clt.opts.auth,
		Manu: manu,
		Plcy: clt.opts.policy,
		Name: clt.opts.name,
		Vers: vers,
		Desc: desc,
		Hash: hash,
		Host: srvh,
		Appl: appl,
		Hdrs: "", // TODO: send these in the update
		Cmdl: strings.Join(os.Args, " "),
	}
	if res, err := clt.atlas.NewScrub(ctx, req); err == nil {
		clt.scid = res.Scid
		return nil
	} else {
		return err
	}
}

func (clt *Shell) rebates(stop chan any, rbts chan *Rebate) error {
	md  := metadata.New(map[string]string{"auth": clt.opts.auth, "vers": vers, "manu": manu, "scid": fmt.Sprintf("%d", clt.scid)})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c,f := context.WithCancel(ctx)
	
	if strm, err := clt.atlas.Rebates(c); err == nil {
		for rbt := range rbts {
			select {
			case <-stop:
				f()
				return nil
			default:
				if err := strm.Send(rbt); err != nil {
					f()
					return err
				}
			}
		}
		strm.CloseSend()
		f()
	} else {
		f()
		return err
	}
	return nil
}

func (clt *Shell) done() {

}
