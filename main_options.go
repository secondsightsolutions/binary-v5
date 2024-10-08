package main

import (
    "flag"
    "strings"
)

func options() {
    name = strings.ToLower(X509cname())
    
    if app == "client" || strings.EqualFold(name, "brg") {
        flag.StringVar(&auth, "auth",     auth,  "Authorization token")
        flag.BoolVar(&doPing, "ping",     false, "Ping the server and exit")
        flag.BoolVar(&doVers, "version",  false, "Print application details and exit")
        flag.StringVar(&fin,  "in",       fin,   "Rebate input file")
        flag.StringVar(&fout, "out",      fout,  "Rebate output file")

        if Type == "proc" || strings.EqualFold(name, "brg") {
            flag.StringVar(&manu, "manu", manu, "Manufacturer name")
        }
        if strings.EqualFold(name, "brg") {
            flag.StringVar(&name, "proc",  name, "Run as processor name")
            flag.StringVar(&test, "test",  "",   "Test directory")
        }
    }
    // if app == "server" || strings.EqualFold(name, "brg") {

    // }
    // if app == "service" || strings.EqualFold(name, "brg") {

    // }
    if strings.EqualFold(name, "brg") {
        flag.BoolVar(&runClient,  "client",  runClient,  "Run binary client")
        flag.BoolVar(&runServer,  "server",  runServer,  "Run binary server")
        flag.BoolVar(&runService, "service", runService, "Run binary service")
    }
    
    flag.Parse()
    if app == "client" {
        if manu == "" {
            exit(nil, 1, "missing manu")
        }
    }
}

