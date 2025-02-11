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

	shell = &Shell{opts: opts}
	
	shell.X509cert, shell.TLSCert = crypt_init("shell", "run_shell", 31, shell_cert, cacr, "", shell_pkey)
	shell.atlas = grpc_connect[AtlasClient](atlas_grpc, atlas_grpc_port, shell.TLSCert, NewAtlasClient)

	switch opts.comd {
	case "vers":
		version()
		exit(nil, 0, "")
		
	case "conf":
		config()
		exit(nil, 0, "")

	case "ping":
		if err := shell.ping(); err == nil {
			exit(nil, 0, "ping succeeded")
		} else {
			exit(nil, 11, "ping failed: %s", err.Error())
		}

	case "invoice":
		if ivid, err := shell.upload_invoice(opts.invoice); err == nil {
			exit(nil, 0, "upload succeeded (%d)", ivid)
		} else {
			exit(nil, 12, "upload %s failed: %s", opts.invoice, err.Error())
		}

	case "scrub":
		if ivid, err := strconv.ParseInt(opts.scrub, 10, 64); err == nil {
			if scid, err := shell.run_scrub(ivid); err == nil {
				exit(nil, 0, "%d", scid)
			} else {
				exit(nil, 13, "scrub %s failed: %s", opts.scrub, err.Error())
			}
		} else {
			exit(nil, 14, "please provide a valid integer as the scrub_id")
		}

	case "queue":
		if ivid, err := strconv.ParseInt(opts.scrub, 10, 64); err == nil {
			if scid, err := shell.run_queue(ivid); err == nil {
				exit(nil, 0, "%d", scid)
			} else {
				exit(nil, 15, "scrub %s failed: %s", opts.scrub, err.Error())
			}
		} else {
			exit(nil, 16, "please provide a valid integer as the scrub_id")
		}

	// case "scrubs":
	// 	if err := shell.get_scrubs(opts.scrubs); err == nil {
	// 		exit(nil, 0, "")
	// 	} else {
	// 		exit(nil, 17, "")
	// 	}

	// case "invoices":
	// 	if err := shell.get_invoices(opts.invoices); err == nil {
	// 		exit(nil, 0, "")
	// 	} else {
	// 		exit(nil, 18, "")
	// 	}

	// case "report":
	// 	if err := shell.save_report(opts.report); err == nil {
	// 		exit(nil, 0, "")
	// 	} else {
	// 		exit(nil, 19, "")
	// 	}

	default:
		exit(nil, 20, "unrecognized shell command: %s", opts.comd)
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
