package main

import (
	context "context"
	"crypto/tls"
	"fmt"
	"io"
	"strings"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func (clt *Shell) connect() (AtlasClient, error) {
	tgt := fmt.Sprintf("%s:%d", atlas_grpc, atlas_grpc_port)
	cfg := &tls.Config{
		Certificates: []tls.Certificate{*clt.TLSCert},
		RootCAs:      X509pool,
	}
	crd := credentials.NewTLS(cfg)
	if conn, err := grpc.NewClient(tgt, 
		grpc.WithTransportCredentials(crd),
	); err == nil {
		clt.atlas = NewAtlasClient(conn)
		return NewAtlasClient(conn), nil
	} else {
		return nil, err
	}
}

func (sh *Shell) ping() error {
	if sh.atlas == nil {
		if clt, err := sh.connect(); err == nil {
			sh.atlas = clt
		}
	}
	ctx := addMeta(context.Background(), nil)
	c,f := context.WithCancel(ctx)
	defer f()
	if _, err := sh.atlas.Ping(c, &Req{}); err == nil {
		fmt.Println("Pong!")
		return nil
	} else {
		fmt.Println(err.Error())
		return err
	}
}

func (sh *Shell) upload(file string) (int64, error) {
	if sh.atlas == nil {
		if clt, err := sh.connect(); err == nil {
			sh.atlas = clt
		}
	}
	ivid := int64(-1)

	if hdrs, chn, err := import_file[Rebate](file, ","); err == nil {
		ctx := addMeta(context.Background(), map[string]string{
			"file": file,
			"hdrs": strings.Join(hdrs, ","),
		})
		c,f := context.WithCancel(ctx)
		defer f()
		if strm, err := sh.atlas.Invoice(c); err == nil {
			if hdr, err := strm.Header(); err == nil {
				ivid = metaValueInt64(hdr, "ivid")
			} else {
				fmt.Println("failed reading header")
				return -1, err
			}
			for rbt := range chn {
				fmt.Printf("sending object: %v\n", rbt)
				if err := strm.Send(rbt); err != nil {
					fmt.Printf("sending object: %v : %s\n", rbt, err.Error())
					return ivid, err
				}
			}
			strm.CloseSend()
			strm.CloseAndRecv()
		} else {
			fmt.Printf("shell.upload(): atlas.Invoice() failed: %s\n", err.Error())
		}
	} else {
		fmt.Printf("shell.upload(): import_file failed: %s\n", err.Error())
	}
	return ivid, nil
}

func (sh *Shell) scrub(ivid int64) (int64, error) {
	ctx := addMeta(context.Background(), map[string]string{
		"plcy": sh.opts.policy,
		"kind": sh.opts.kind,
		"test": "",
	})
	c,f := context.WithCancel(ctx)
	defer f()
	req := &ScrubReq{Manu: manu, Ivid: ivid}
	scid := int64(-1)

	if strm, err := sh.atlas.Scrub(c, req); err == nil {
		if hdr, err := strm.Header(); err == nil {
			scid = metaValueInt64(hdr, "scid")
		} else {
			return scid, err
		}
		for {
			if rr, err := strm.Recv(); err == nil {
				fmt.Printf("%v\n", rr)
			} else if err == io.EOF {
				return scid, nil
			} else {
				return scid, err
			}
		}
	}
	return scid, nil
}
