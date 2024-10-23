package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
    _ "embed"
)

func client_main(wg *sync.WaitGroup, stop chan any) {
    defer wg.Done()
    memoryWatch(stop)
    <-stop
}

func exit(sc *Scrub, code int, msg string, args ...any) {
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
    // if sc != nil {
    //     update_scrub(sc, mesg)
    // }
	os.Exit(code)
}
