package main

import (
    "flag"
    "strings"
)

func options() {
    name = strings.ToLower(X509cname())
    
    if strings.EqualFold(appl, "client") {
        flag.BoolVar(&doVers, "version",  false, "Print application details and exit")
        flag.BoolVar(&doPing, "ping",     false, "Ping the server and exit")
        flag.StringVar(&auth, "auth",     auth,  "Authorization token")
        flag.StringVar(&fin,  "in",       fin,   "Rebate input file")
        flag.StringVar(&fout, "out",      fout,  "Rebate output file")

        if Type == "proc" || strings.EqualFold(name, "brg") {
            flag.StringVar(&manu, "manu", manu, "Manufacturer name")
        }
        if strings.EqualFold(name, "brg") {
            flag.StringVar(&name, "proc",  name, "Run as processor name")
            flag.StringVar(&test, "test",  "",   "Test directory")
        }
    } else if strings.EqualFold(appl, "server") {
        flag.BoolVar(&doVers, "version",  false, "Print application details and exit")
        flag.BoolVar(&doPing, "ping",     false, "Ping the server and exit")

    } else if strings.EqualFold(appl, "service") {
        flag.BoolVar(&doVers, "version",  false, "Print application details and exit")

    }
   
    if strings.EqualFold(name, "brg") {
        flag.BoolVar(&runClient,  "client",  runClient,  "Run binary client")
        flag.BoolVar(&runServer,  "server",  runServer,  "Run binary server")
        flag.BoolVar(&runService, "service", runService, "Run binary service")
    }
    
    flag.Parse()
    if appl == "client" {
        if manu == "" {
            exit(nil, 1, "missing manu")
        }
    }
}

