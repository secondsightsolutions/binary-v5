package main

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Shell struct {
	opts     *Opts
	atlas    AtlasClient
	TLSCert  *tls.Certificate
    X509cert *x509.Certificate
}

var shell *Shell

func run_shell(opts *Opts) {
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
		Log("shell", "run_shell", "", "cannot initialize crypto", 0, nil, err)
		exit(nil, 1, fmt.Sprintf("shell cannot initialize crypto: %s", err.Error()))
	}

	if opts.runPing {
		shell.ping()
		exit(nil, 0, "")
	}
	var ivid int64 = -1
	var scid int64 = -1
	if opts.upload != "" {
		if _, err = shell.upload(opts.upload); err != nil {
			exit(nil, 2, fmt.Sprintf("upload %s failed: %s", opts.upload, err.Error()))
		}
		exit(nil, 0, "upload succeeded")
	}
	if opts.invoice != "" {
		if ivid < 0 {
			if ivid, err = strconv.ParseInt(opts.invoice, 10, 64); err != nil {
				if ivid, err = shell.upload(opts.invoice); err != nil {
					exit(nil, 3, fmt.Sprintf("upload %s failed: %s", opts.invoice, err.Error()))
				}
			}
		}
	}
	if ivid >= 0 {
		if scid, err = shell.scrub(ivid); err != nil {
			exit(nil, 3, fmt.Sprintf("scrub %s failed: %s", opts.invoice, err.Error()))
		} else {
			fmt.Printf("scid: %d\n", scid)
		}
	}

	// if _, objs, err := import_file[Rebate](opts.fileIn, opts.csep, hdrm); err == nil {
	// 	fmt.Println("run_shell: import_file returned, calling shell.rebates")
		// if out, err := shell.rebates(stop, objs); err == nil {
		// 	fmt.Println("run_shell: shell.rebates returned successfully, starting to send rebates to atlas")
		// 	for rbt := range out { // TODO: getting stuck here
		// 		fmt.Printf("rebate=%v\n", rbt)
		// 		out <-rbt
		// 	}
		// } else {
		// 	fmt.Println(err.Error())
		// }
	// } else {
	// 	fmt.Printf("import_file failed, err=%s\n", err.Error())
	// }
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
