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
)

func (clt *Shell) connect() (AtlasClient, error) {
	tgt := fmt.Sprintf("%s:%d", atlas_grpc, atlas_grpc_port)
	cfg := &tls.Config{
		Certificates: []tls.Certificate{*clt.TLSCert},
		RootCAs:      X509pool,
	}
	crd := credentials.NewTLS(cfg)
	if conn, err := grpc.NewClient(tgt, grpc.WithTransportCredentials(crd)); err == nil {
		clt.atlas = NewAtlasClient(conn)
		return NewAtlasClient(conn), nil
	} else {
		return nil, err
	}
}

func (sh *Shell) rebates(stop chan any, rbts chan *Rebate) (chan *Rebate, error) {
	if sh.atlas == nil {
		if clt, err := sh.connect(); err == nil {
			fmt.Println("shell connected to atlas")
			sh.atlas = clt
		}
	}
	ctx := metaGRPC(map[string]string{
		"plcy": sh.opts.policy,
		"kind": sh.opts.kind,
		"dscr": desc,
		"hash": hash,
		"host": "",
		"hdrs": "",
		"cmdl": strings.Join(os.Args, " "),
		"test": "",
	})
	c,f := context.WithCancel(ctx)
	dur := time.Duration(0) * time.Second
	out := make(chan *Rebate, 1000)

	defer f()
	for {
		if strm, err := sh.atlas.Rebates(c); err == nil {
			if hdrs, err := strm.Header(); err == nil {
				sh.scid = metaValueInt64(hdrs, "scid")
				fmt.Printf("shell.Rebates() received scid=%d\n", sh.scid)
			} else {
				fmt.Printf("shell.Rebates() cannot access headers, err=%s\n", err.Error())
			}

			// We have the scid (scrub_id). Now start sending up the rebates.
			go func() {
				fmt.Printf("shell.Rebates() starting send loop\n")
				for rbt := range rbts {
					fmt.Printf("shell.Rebates() received input rebate=(%v)\n", rbt)
					select {
					case <-stop:
						goto done
					case <-time.After(dur):
						if err := strm.Send(rbt); err != nil {
							if strings.Contains(err.Error(), "Unavailable") {
								//log("shell", "rebates", "error sending rebates (network error, retrying)", 0, err)
								Log("shell", "rebates", "", "error sending rebates (network error, retrying)", 0, nil, nil)
								dur = time.Duration(5) * time.Second
							} else {
								//log("shell", "rebates", "error sending rebates (giving up)", 0, err)
								Log("shell", "rebates", "", "error sending rebates (giving up)", 0, nil, nil)
								goto done
							}
						} else {
							dur = time.Duration(0) * time.Second
							fmt.Printf("shell.Rebates() successfully sent rebate=(%v)\n", rbt)
						}
					}
				}
				done:
				fmt.Printf("shell.Rebates() finished send loop\n")
				strm.CloseSend()
			}()
			
			go func() {
				fmt.Printf("shell.Rebates() starting recv loop\n")
				for {
					if rbt, err := strm.Recv(); err == nil {
						fmt.Printf("shell.Rebates() successfully recv from server rebate=(%v)\n", rbt)
						out <-rbt
					} else {
						fmt.Printf("shell.Rebates() failed to recv from server err=%s\n", err.Error())
						break
					}
				}
				fmt.Printf("shell.Rebates() finished recv loop\n")
			}()
			return out, nil
		} else {
			//fmt.Printf("shell.Rebates() cannot connect to atlas, err=%s\n", err.Error())
			time.Sleep(time.Duration(5)*time.Second)
		}
	}
}

