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

func version() {
    fmt.Printf("%s: %s\n", "name", X509ou())
    fmt.Printf("%s: %s\n", "desc", desc)
    fmt.Printf("%s: %s\n", "type", Type)
    fmt.Printf("%s: %s\n", "envr", envr)
    fmt.Printf("%s: %s\n", "vers", vers)
    fmt.Printf("%s: %s\n", "hash", hash)
    fmt.Printf("%s: %s\n", "manu", manu)
    fmt.Printf("%s: %s\n", "host", host)
    fmt.Printf("%s: %s\n", "port", port)
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
