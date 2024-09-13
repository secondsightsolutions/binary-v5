package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
    CryptInit(cert, cacr, "", pkey, salt, phrs)
    getEnv()
    parseCommandLine()
    
    if doPing {
        ping()
        return
    } else if doVers {
        version()
        return
    }

    done := make(chan any)
    memoryWatch(done)

    if Http {
        screen(time.Now(), nil, -1, 0, ScreenLevel.Text, ScreenLevel.Text, true, "http binary starting")
        http.HandleFunc("/", httpRun)
        http.ListenAndServe(":80", nil)
    } else {
        screen(time.Now(), nil, -1, 0, ScreenLevel.Text, ScreenLevel.Text, true, "local binary starting")
        scid = createScrub()
        sc := new_scrub(scid, manu)
        sc.sr.files["rebates"].path = fin
        sc.sr.outf = fout
        sc.load_caches()
        sc.spis.load(sc.cs["spis"])
        sc.run()
    }
    done <-nil
}

func parseCommandLine() {
    name   = strings.ToLower(X509cname())
    
    flag.StringVar(&auth,    "auth",     auth,    "Authorization token")
    flag.BoolVar(&doPing,    "ping",     false,   "Ping the server and exit")
    flag.BoolVar(&doVers,    "version",  false,   "Print application details and exit")
    flag.BoolVar(&Http,      "http",     false,   "Run as HTTP server")
    flag.StringVar(&fin,     "in",       fin,     "Rebate input file")
    flag.StringVar(&fout,    "out",      fout,    "Rebate output file")
    flag.StringVar(&test,    "test",     "",      "Test directory")

    if Type != "manu" || strings.EqualFold(name, "brg") {
        flag.StringVar(&manu, "manu", manu, "Manufacturer name")
    }
    if strings.EqualFold(name, "brg") {
        flag.StringVar(&name,    "proc",       name,    "Run as processor name")
    }
    flag.Parse()
    if manu == "" {
        exit(nil, 1, "missing manu")
    }
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
        update_scrub(sc, mesg)
    }
	os.Exit(code)
}

func getEnv() {
    setIf(&manu,    "BIN_MANU")
    setIf(&hash,    "BIN_HASH")
    setIf(&pkey,    "BIN_PKEY")
    setIf(&cacr,    "BIN_CACR")
    setIf(&cert,    "BIN_MYCR")
    setIf(&phrs,    "BIN_PHRS")
    setIf(&salt,    "BIN_SALT")
    setIf(&host,    "BIN_HOST")
    setIf(&port,    "BIN_PORT")
    setIf(&desc,    "BIN_DESC")
    setIf(&envr,    "BIN_ENVR")
    setIf(&auth,	"BIN_AUTH")
}

func setIf(envVar *string, envName string) {
    if envVal := os.Getenv(envName); envVal != "" {
        *envVar = envVal
    }
}
