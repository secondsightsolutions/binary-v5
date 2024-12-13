package main

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"sync"
)

type Shell struct {
	opts     *Opts
	scid     int64 // scrub id
	atlas    AtlasClient
	TLSCert  *tls.Certificate
    X509cert *x509.Certificate
}

var shell *Shell

func run_shell(wg *sync.WaitGroup, opts *Opts, stop chan any) {
	defer wg.Done()

	//memoryWatch(stop)

	hdrm := map[string]string{}
	toks1 := strings.Split(opts.hdrm, ",")
	for _, tok1 := range toks1 {
		if tok1 != "" {
			toks2 := strings.Split(tok1, "=")
			if len(toks2) == 2 {
				cust := toks2[0]
				stnd := toks2[1]
				hdrm[cust] = stnd
			}
		}
	}

	var err error
	shell = &Shell{opts: opts}
	if shell.TLSCert, shell.X509cert, err = CryptInit(shell_cert, cacr, "", shell_pkey, salt, phrs); err != nil {
		log("shell", "run_shell", "cannot initialize crypto", 0, err)
		exit(nil, 1, fmt.Sprintf("shell cannot initialize crypto: %s", err.Error()))
	}
	
	if _, objs, err := import_file[Rebate](opts.fileIn, opts.csep, hdrm); err == nil {
		if out, err := shell.rebates(stop, objs); err == nil {
			for rbt := range out {
				fmt.Printf("rebate=%v\n", rbt)
			}
		} else {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Printf("import_file failed, err=%s\n", err.Error())
	}
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
