package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
    var wg sync.WaitGroup

    stop := make(chan any)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    
    //dflts()
    getEnv()
    CryptInit(cert, cacr, "", pkey, salt, phrs)
    options()

    if doPing {
        ping()
        exit(nil, 0, "")
    } else if doVers {
        version()
        exit(nil, 0, "")
    }
    
    if runClient {
        wg.Add(1)
        go client_main(&wg, stop)
    }
    if runServer {
        wg.Add(1)
        go server_main(&wg, stop)
    }
    if runService {
        wg.Add(1)
        go service_main(&wg, stop)
    }
    <-sigs
    close(stop)
    wg.Wait()
}

func getEnv() {
    setIf(&manu,    "BIN_MANU")
    setIf(&hash,    "BIN_HASH")
    setIf(&pkey,    "BIN_PKEY")
    setIf(&cacr,    "BIN_CACR")
    setIf(&cert,    "BIN_MYCR")
    setIf(&phrs,    "BIN_PHRS")
    setIf(&salt,    "BIN_SALT")
    setIf(&srvh,    "BIN_SRVH")
    setIf(&svch,    "BIN_SVCH")
    setIf(&desc,    "BIN_DESC")
    setIf(&envr,    "BIN_ENVR")
    setIf(&auth,	"BIN_AUTH")
}

func setIf(envVar *string, envName string) {
    if envVal := os.Getenv(envName); envVal != "" {
        *envVar = envVal
    }
}

func version() {
	fmt.Printf("%s: %s\n", "appl", appl)
    fmt.Printf("%s: %s\n", "name", X509ou())
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

