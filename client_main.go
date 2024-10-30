package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Shell struct {
	opts  *Opts
	file  *rebate_file
	scid  int64 // scrub id
	atlas AtlasClient
}

var shell *Shell

func run_shell(wg *sync.WaitGroup, opts *Opts, stop chan any) {
	defer wg.Done()

	//memoryWatch(stop)

	shell = &Shell{opts: opts}
	shell.connect()
	
	for {
		if err := shell.newScrub(); err != nil {
			log("shell", "run_shell", "cannot create scrub", 0, err)
			time.Sleep(time.Duration(10)*time.Second)
			shell.connect()
		} else {
			break
		}
	}
	shell.file = new_rebate_file(opts)
	
	for {
		if rbtc, err := shell.file.read(); err != nil {
			log("shell", "run_shell", "cannot read file %s", 0, err, shell.file.path)
			break
		} else {
			if err := shell.rebates(stop, rbtc); err != nil {
				time.Sleep(time.Duration(10)*time.Second)
				shell.connect()
			} else {
				break
			}
		}
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
