package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	srvp int = 23460
	svcp int = 23461
	opts *Opts
)

func main() {
	var wg sync.WaitGroup

	stop := make(chan any)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	getEnv()
	CryptInit(cert, cacr, "", pkey, salt, phrs)
	opts = options()

	if opts.runPing {
		ping()
		exit(nil, 0, "")
	} else if opts.runVers {
		version()
		exit(nil, 0, "")
	}

	if opts.runTitan {
		wg.Add(1)
		go run_titan(&wg, opts, stop)
	}
	if opts.runAtlas {
		wg.Add(1)
		go run_atlas(&wg, opts, stop)
	}
	if opts.runClient {
		wg.Add(1)
		go run_shell(&wg, opts, stop)
	}
    <-sigs
	close(stop)
	wg.Wait()
}

func getEnv() {
	setIf(&manu, "BIN_MANU")
	setIf(&hash, "BIN_HASH")
	setIf(&pkey, "BIN_PKEY")
	setIf(&cacr, "BIN_CACR")
	setIf(&cert, "BIN_MYCR")
	setIf(&phrs, "BIN_PHRS")
	setIf(&salt, "BIN_SALT")
	setIf(&srvh, "BIN_SRVH")
	setIf(&svch, "BIN_SVCH")
	setIf(&desc, "BIN_DESC")
	setIf(&envr, "BIN_ENVR")
}

func setIf(envVar *string, envName string) {
	if envVal := os.Getenv(envName); envVal != "" {
		*envVar = envVal
	}
}

func version() {
	fmt.Printf("%s: %s\n", "appl", appl)
	fmt.Printf("%s: %s\n", "name", X509cname())
	fmt.Printf("%s: %s\n", "full", X509ou())
	fmt.Printf("%s: %s\n", "desc", desc)
	fmt.Printf("%s: %s\n", "type", Type)
	fmt.Printf("%s: %s\n", "envr", envr)
	fmt.Printf("%s: %s\n", "vers", vers)
	fmt.Printf("%s: %s\n", "hash", hash)
	fmt.Printf("%s: %s\n", "manu", manu)
	fmt.Printf("%s: %s\n", "srvh", srvh)
	fmt.Printf("%s: %d\n", "srvp", srvp)
	fmt.Printf("%s: %s\n", "svch", svch)
	fmt.Printf("%s: %d\n", "svcp", svcp)
}
