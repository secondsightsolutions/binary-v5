package main

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Shell struct {
	opts     *Opts
	file     *rebate_file
	scid     int64 // scrub id
	atlas    AtlasClient
	TLSCert  *tls.Certificate
    X509cert *x509.Certificate
}

var shell *Shell

func run_shell(wg *sync.WaitGroup, opts *Opts, stop chan any) {
	defer wg.Done()

	//memoryWatch(stop)

	var err error
	shell = &Shell{opts: opts}
	if shell.TLSCert, shell.X509cert, err = CryptInit(shell_cert, cacr, "", shell_pkey, salt, phrs); err != nil {
		log("shell", "run_shell", "cannot initialize crypto", 0, err)
		exit(nil, 1, fmt.Sprintf("shell cannot initialize crypto: %s", err.Error()))
	}
	intv := time.Duration(0)
	
	shell.file = new_rebate_file(opts)
	if rbtc, err := shell.file.read(); err != nil {
		log("shell", "run_shell", "cannot read file %s", 0, err, shell.file.path)
	} else {
		intv = 0
		for {
			select {
			case <-time.After(intv):
				intv = time.Duration(10)
				shell.connect()
				if err := shell.rebates(stop, rbtc); err != nil {
					if strings.Contains(err.Error(), "Unavailable") {
						log("shell", "run_shell", "error sending rebates (network error, retrying)", 0, err)
					} else {
						log("shell", "run_shell", "error sending rebates (giving up)", 0, err)
						return
					}
				} else {
					goto rebates_sent
				}
			case <-stop:
				log("shell", "run_shell", "stopping", 0, err)
				return
			}
		}
	}
	rebates_sent:
}

func exit(sc *scrub, code int, msg string, args ...any) {
	nargs := []any{}
	for _, arg := range args {
		if sarg, ok := arg.(string); ok {
			sarg = strings.TrimPrefix(sarg, "rpc error: code = Unknown desc = ")
			nargs = append(nargs, sarg)
		} else {
			nargs = append(nargs, arg)
		}
	}
	mesg := ""
	if msg != "" {
		mesg = fmt.Sprintf(msg, nargs...)
		fmt.Println(mesg)
	}
	if sc != nil {
		fmt.Println("scrub exit")
	}
	os.Exit(code)
}
