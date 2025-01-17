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
    auth      string
    kind      string
    hdrm      string
    csep      string
    upload    string
    invoice   string
    //fileOut   string
    policy    string
    test      string
}

func options() *Opts {
    opts := &Opts{}
    flag.BoolVar(&opts.runVers, "version", false,       "Print application details and exit")
    flag.StringVar(&opts.kind,  "kind",    "pharmacy",  "Scrub type (pharmacy, medical, etc.)")

    if strings.EqualFold(name, "brg") {
        flag.BoolVar(&opts.runClient, "client",  false,         "Run client")
        flag.BoolVar(&opts.runAtlas,  "atlas",   false,         "Run atlas")
        flag.BoolVar(&opts.runTitan,  "titan",   false,         "Run titan")
        flag.BoolVar(&opts.runPing,   "ping",    false,         "Ping the server and exit")
        flag.BoolVar(&opts.runConf,   "config",  false,         "Print application configuration and exit")
        flag.StringVar(&opts.auth,    "auth",    "",            "Authorization token")
        flag.StringVar(&opts.upload,  "upload",  "",            "Upload invoice file")
        flag.StringVar(&opts.invoice, "scrub",   "",            "Rebate file or invoice number")
        //flag.StringVar(&opts.fileOut, "out",     "",            "Scrub output file")
        flag.StringVar(&opts.policy,  "policy",  "default",     "Policy name")
        flag.StringVar(&opts.hdrm,    "hdrs",    "",            "Header map (cust1:std1,cust2:std2,...)")
        flag.StringVar(&opts.csep,    "csep",    ",",           "Rebate file column separator")
        flag.StringVar(&manu,         "manu",    manu,          "Manufacturer name")
        flag.StringVar(&name,         "proc",    name,          "Run as processor name")
        flag.StringVar(&opts.test,    "test",    "",            "Test directory")
       
    } else if Type == "manu" {
        flag.BoolVar(&opts.runClient, "client",  false,         "Run client")
        flag.BoolVar(&opts.runPing,   "ping",    false,         "Ping the server and exit")
        flag.BoolVar(&opts.runConf,   "config",  false,         "Print application configuration and exit")
        flag.StringVar(&opts.auth,    "auth",    "",            "Authorization token")
        flag.StringVar(&opts.upload,  "upload",  "",            "Upload invoice file")
        flag.StringVar(&opts.invoice, "scrub",   "",            "Rebate file or invoice number")
        // flag.StringVar(&opts.fileOut, "out",     "",            "Rebate output file")
        flag.StringVar(&opts.policy,  "policy",  "default",     "Policy name")
        flag.StringVar(&opts.hdrm,    "hdrs",    "",            "Header map (cust1:std1,cust2:std2,...)")
        flag.StringVar(&opts.csep,    "csep",    ",",           "Rebate file column separator")
    } else {
        flag.BoolVar(&opts.runPing,   "ping",    false,         "Ping the server and exit")
        flag.BoolVar(&opts.runConf,   "config",  false,         "Print application configuration and exit")
        flag.StringVar(&opts.auth,    "auth",    "",            "Authorization token")
        flag.StringVar(&opts.upload,  "upload",  "",            "Upload invoice file")
        flag.StringVar(&opts.invoice, "scrub",   "",            "Rebate file or invoice number")
        // flag.StringVar(&opts.fileOut, "out",     "",            "Rebate output file")
        flag.StringVar(&opts.policy,  "policy",  "default",     "Policy name")
        flag.StringVar(&opts.hdrm,    "hdrs",    "",            "Header map (cust1:std1,cust2:std2,...)")
        flag.StringVar(&opts.csep,    "csep",    ",",           "Rebate file column separator")
    }

    flag.Parse()

    if !opts.runVers && !opts.runConf && !opts.runPing {
        if name != "brg" {
            if Type == "manu" {
                if !opts.runClient {
                    opts.runAtlas = true
                }
            } else {
                opts.runClient = true
            }
        }
    }
    if (opts.runClient || opts.runAtlas) && manu == "" {
        exit(nil, 1, "missing manu")
    }
    if opts.runClient && opts.invoice == "" && opts.upload == "" && !opts.runPing {
        exit(nil, 2, "missing invoice file (or number) - please provide -upload or -scrub (or both)")
    }
    if opts.runClient && opts.auth == "" {
        exit(nil, 3, "missing auth token")
    }
    return opts
}

