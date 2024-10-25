package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func (clt *Client) connect() {
	tgt := fmt.Sprintf("%s:%d", svch, svcp)
	cfg := &tls.Config{
		Certificates: []tls.Certificate{TLSCert},
		RootCAs:      X509pool,
	}
	crd := credentials.NewTLS(cfg)
	if conn, err := grpc.NewClient(tgt, grpc.WithTransportCredentials(crd)); err == nil {
		clt.srv = NewBinaryV5SrvClient(conn)
	}
}

func (clt *Client) start() int64 {
    md  := metadata.New(map[string]string{"auth": clt.opts.auth, "vers": vers, "manu": manu})
    ctx := metadata.NewOutgoingContext(context.Background(), md)
    req := &StartReq{
    	Auth: clt.opts.auth,
    	Manu: manu,
    	Plcy: clt.opts.policy,
    	Name: clt.opts.name,
    	Vers: vers,
    	Desc: desc,
    	Hash: hash,
    	Host: srvh,
    	Type: appl,
    	Hdrs: strings.Join(clt.hdrs,   ","),
    	Cmdl: strings.Join(os.Args, " "),
    }
    if res, err := clt.srv.Start(ctx, req); err == nil {
        return res.Scid
    } else {
        return -1
    }
}

func (clt *Client) scrub(rbts chan *Rebate) {
    var strm grpc.ClientStreamingClient[Rebate, Metrics]
    md  := metadata.New(map[string]string{"auth": clt.opts.auth, "vers": vers, "manu": manu, "scid": fmt.Sprintf("%d", clt.scid)})
    ctx := metadata.NewOutgoingContext(context.Background(), md)
	for rbt := range rbts {
        for strm == nil {
            var err error
            if strm, err = clt.srv.Scrub(ctx); err != nil {
                time.Sleep(time.Duration(1)*time.Second)
            }
        }
        if err := strm.Send(rbt); err != nil {
            strm = nil
        }
	}
}

func (clt *Client) done() {

}