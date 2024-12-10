package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"strings"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func (clt *Shell) connect() {
	tgt := fmt.Sprintf("%s:%d", atlas_grpc, atlas_grpc_port)
	cfg := &tls.Config{
		Certificates: []tls.Certificate{*clt.TLSCert},
		RootCAs:      X509pool,
	}
	crd := credentials.NewTLS(cfg)
	if conn, err := grpc.NewClient(tgt, grpc.WithTransportCredentials(crd)); err == nil {
		clt.atlas = NewAtlasClient(conn)
	} else {
		log("shell", "connect", "cannot connect to atlas", 0, err)
	}
}

func (clt *Shell) rebates(stop chan any, rbts chan *Rebate) error {
	md  := metadata.New(map[string]string{
		"auth": clt.opts.auth, 
		"manu": manu,
		"plcy": clt.opts.policy,
		"kind": kind, 
		"name": name,
		"vers": vers,
		"desc": desc,
		"hash": hash,
		"host": "",
		"appl": appl,
		"hdrs": "",
		"cmdl": strings.Join(os.Args, " "),
		"test": "",
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c,f := context.WithCancel(ctx)
	
	if strm, err := clt.atlas.Rebates(c); err == nil {
		if hdrs, err := strm.Header();err == nil {
			clt.scid = metaValueInt64(hdrs, "scid")
		}
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
		for {
			if res, err := strm.Recv(); err == nil {
				fmt.Printf("res: %v\n", res)
			} else if err == io.EOF {
				break
			} else {
				fmt.Printf("error: %s\n", err.Error())
			}
		}
		f()
	} else {
		f()
		return err
	}
	return nil
}

