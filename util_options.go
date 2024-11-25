package main

import (
	"flag"
	"strings"
)

type Opts struct {
    runVers   bool
    runPing   bool
    runConf   bool
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
    flag.BoolVar(&opts.runVers, "version",  false, "Print application details and exit")

    if strings.EqualFold(name, "brg") {
        flag.BoolVar(&opts.runClient, "client",  opts.runClient, "Run client")
        flag.BoolVar(&opts.runAtlas,  "atlas",   opts.runAtlas,  "Run atlas")
        flag.BoolVar(&opts.runTitan,  "titan",   opts.runTitan,  "Run titan")
        flag.BoolVar(&opts.runPing,   "ping",    false,          "Ping the server and exit")
        flag.BoolVar(&opts.runConf,   "config",  false,          "Print application configuration and exit")
        flag.StringVar(&opts.auth,    "auth",    "",             "Authorization token")
        flag.StringVar(&opts.fileIn,  "in",      "",             "Rebate input file")
        flag.StringVar(&opts.fileOut, "out",     "",             "Rebate output file")
        flag.StringVar(&opts.policy,  "policy",  "default",      "Rebate output file")
        flag.StringVar(&opts.hdrm,    "hdrs",    "",             "Header map (cust1:std1,cust2:std2,...)")
        flag.StringVar(&opts.csep,    "csep",    ",",            "Rebate file column separator")
        flag.StringVar(&manu,         "manu",    manu,           "Manufacturer name")
        flag.StringVar(&opts.name,    "proc",    name,           "Run as processor name")
        flag.StringVar(&opts.test,    "test",    "",             "Test directory")
        
        if strings.EqualFold(appl, "shell") {
            opts.runClient = true
        } else if strings.EqualFold(appl, "atlas") {
            opts.runAtlas = true
        } else if strings.EqualFold(appl, "titan") {
            opts.runTitan = true
        }

    } else {
        if strings.EqualFold(appl, "shell") {
            opts.runClient = true
            flag.BoolVar(&opts.runPing,   "ping",    false,     "Ping the server and exit")
            flag.StringVar(&opts.auth,    "auth",    "",        "Authorization token")
            flag.StringVar(&opts.fileIn,  "in",      "",        "Rebate input file")
            flag.StringVar(&opts.fileOut, "out",     "",        "Rebate output file")
            flag.StringVar(&opts.policy,  "policy",  "default", "Rebate output file")
            flag.StringVar(&opts.hdrm,    "hdrs",    "",        "Header map (cust1:std1,cust2:std2,...)")
            flag.StringVar(&opts.csep,    "csep",    ",",       "Rebate file column separator")
            if Type == "proc" {
                flag.StringVar(&manu, "manu", manu, "Manufacturer name")
            }
        } else if strings.EqualFold(appl, "atlas") {
            opts.runAtlas = true
            flag.BoolVar(&opts.runPing, "ping",     false, "Ping the server and exit")

        } else if strings.EqualFold(appl, "titan") {
            exit(nil, 99, "cannot run a titan build with name %s", name)
        }
    }
    flag.Parse()
    if appl == "shell" && !opts.runPing && !opts.runVers && !opts.runConf {
        if manu == "" {
            exit(nil, 1, "missing manu")
        }
        if opts.fileIn == "" {
            exit(nil, 2, "missing rebate file")
        }
    }
    return opts
}

