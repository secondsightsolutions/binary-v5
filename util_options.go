package main

import (
    "flag"
    "strings"
)

type Opts struct {
    runVers   bool
    runPing   bool
    runClient bool
    runAtlas  bool
    runTitan  bool
    name      string
    auth      string
    hdrm      string
    csep      string
    fileIn    string
    fileOut   string
    policy    string
    test      string
}

func options() *Opts {
    opts := &Opts{}
    name := strings.ToLower(X509cname())
    flag.BoolVar(&opts.runVers, "version",  false, "Print application details and exit")

    if strings.EqualFold(appl, "shell") {
        opts.runClient = true
        flag.BoolVar(&opts.runPing,   "ping",    false,     "Ping the server and exit")
        flag.StringVar(&opts.auth,    "auth",    "",        "Authorization token")
        flag.StringVar(&opts.fileIn,  "in",      "",        "Rebate input file")
        flag.StringVar(&opts.fileOut, "out",     "",        "Rebate output file")
        flag.StringVar(&opts.policy,  "policy",  "default", "Rebate output file")
        flag.StringVar(&opts.hdrm,    "hdrs",    "",        "Header map (cust1:std1,cust2:std2,...)")
        flag.StringVar(&opts.csep,    "csep",    ",",       "Rebate file column separator")

        if Type == "proc" || strings.EqualFold(name, "brg") {
            flag.StringVar(&manu, "manu", manu, "Manufacturer name")
        }
        if strings.EqualFold(name, "brg") {
            flag.StringVar(&opts.name, "proc",  name, "Run as processor name")
            flag.StringVar(&opts.test, "test",  "",   "Test directory")
        }
    } else if strings.EqualFold(appl, "atlas") {
        opts.runAtlas = true
        flag.BoolVar(&opts.runPing, "ping",     false, "Ping the server and exit")
        
        if strings.EqualFold(name, "brg") {
            flag.StringVar(&manu, "manu", manu, "Manufacturer name")
            flag.StringVar(&opts.name, "proc", name, "Run as processor name")
        }

    } else if strings.EqualFold(appl, "titan") {
        opts.runTitan = true
    }
   
    if strings.EqualFold(name, "brg") {
        flag.BoolVar(&opts.runClient, "client", opts.runClient, "Run client")
        flag.BoolVar(&opts.runAtlas,  "atlas",  opts.runAtlas,  "Run atlas")
        flag.BoolVar(&opts.runTitan,  "titan",  opts.runTitan,  "Run titan")
    }
    
    flag.Parse()
    if appl == "shell" {
        if manu == "" {
            exit(nil, 1, "missing manu")
        }
        if opts.fileIn == "" {
            exit(nil, 2, "missing rebate file")
        }
    }
    return opts
}

