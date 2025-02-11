package main

import (
	"flag"
	"fmt"
	"strings"
)

type Opts struct {
    runShell  bool
    runAtlas  bool
    runTitan  bool
    comd      string
    auth      string
    kind      string
    hdrm      string
    csep      string
    invoice   string
    scrub     string    // Run foreground scrub on this scrub_id or invoice file.
    queue     string    // Run background scrub on this scrub_id or invoice file.
    scrubs    string    // List of scrub ids - fetch their summaries and return them.
    invoices  string    // List of invoice ids - fetch their summaries and return them.
    report    string    // Scrub_id of scrub to download and save to one or more files.
    policy    string
    test      string
}

func options() *Opts {
    var runVers, runPing, runConf bool
    opts := &Opts{}
    flag.BoolVar(&runVers,      "version", false,       "Print application details and exit")
    flag.BoolVar(&runPing,      "ping",    false,       "Ping the server and exit")
    flag.StringVar(&opts.auth,  "auth",    "",          "Authorization token")
    
    if envr != "prod" {
        flag.BoolVar(&runConf,  "config",  false,       "Print application configuration and exit")
    }

    if appl == "shell" {
        opts.runShell = true
        opts.runAtlas = false
        opts.runTitan = false
    } else if appl == "atlas" {
        opts.runShell = false
        opts.runAtlas = true
        opts.runTitan = false
    } else if appl == "titan" {
        opts.runShell = false
        opts.runAtlas = false
        opts.runTitan = true
    } else if appl == "all" {
        flag.BoolVar(&opts.runShell,   "shell",   false,         "Run shell")
        flag.BoolVar(&opts.runAtlas,   "atlas",   false,         "Run atlas")
        flag.BoolVar(&opts.runTitan,   "titan",   false,         "Run titan")
    }

    if appl == "shell" || appl == "all" {
        flag.StringVar(&opts.invoice,  "invoice", "",            "Upload this invoice file")
        flag.StringVar(&opts.scrub,    "scrub",   "",            "Run scrub on this scrub_id")
        flag.StringVar(&opts.queue,    "queue",   "",            "Run background scrub on this scrub_id")
        flag.StringVar(&opts.scrubs,   "scrubs",  "",            "Fetch summaries for this list of scrubs (eg. 1-8,10,12)")
        flag.StringVar(&opts.invoices, "invoices","",            "Fetch summaries for this list of invoices (eg. 1-8,10,12)")
        flag.StringVar(&opts.report,   "report",  "",            "Download/save report files for a given scrub")

        flag.StringVar(&manu,          "manu",    manu,          "Manufacturer name")
        flag.StringVar(&opts.kind,     "kind",    "pharmacy",    "Scrub type (pharmacy, medical, etc.)")
        flag.StringVar(&opts.policy,   "policy",  "default",     "Policy name")
        flag.StringVar(&opts.hdrm,     "hdrs",    "",            "Header map (cust1:std1,cust2:std2,...)")
        flag.StringVar(&opts.csep,     "csep",    ",",           "Rebate file column separator")

        if strings.EqualFold(name, "brg") {
            flag.StringVar(&name,      "proc",    name,          "Run as processor name")
            flag.StringVar(&opts.test, "test",    "",            "Test directory")
        }
    }

    flag.Parse()
    
    if opts.runShell && (opts.runAtlas || opts.runTitan) {
        exit(nil, 51, "cannot specify both shell and atlas/titan (either run shell client or one/both servers)")
    }

    if runConf || runPing || runVers {
        if runPing {
            if !opts.runShell && !opts.runAtlas {
                exit(nil, 52, "-ping only valid with shell and atlas")
            }
            opts.comd = "ping"
        } else if runConf {
            opts.comd = "conf"
        } else if runVers {
            opts.comd = "vers"
        }
    } else {    // Running full appl - either shell, or ... atlas and/or titan
        if opts.runShell || opts.runAtlas {
            if manu == "" {     // could be from command line or embedded
                exit(nil, 53, "missing manu")
            }
            if opts.auth == "" {
                exit(nil, 54, "missing auth")
            }
        }
        if opts.runShell {
            if opts.invoice != "" {
                opts.comd = "invoice"
            } else if opts.scrub != "" {
                opts.comd = "scrub"
            } else if opts.queue != "" {
                opts.comd = "queue"
            } else if opts.scrubs != "" {
                opts.comd = "scrubs"
            } else if opts.invoices != "" {
                opts.comd = "invoices"
            } else if opts.report != "" {
                opts.comd = "report"
            } else {
                exit(nil, 55, "please provide one of: -ping, -conf, -version, -upload, -scrub, -queue, -scrubs, -report")
            }
        } else {    // atlas and/or titan
            if opts.invoice != "" || opts.scrub != "" || opts.queue != "" || opts.scrubs != "" || opts.invoices != "" || opts.report != "" {
                fmt.Printf("invoice=%s\n", opts.invoice)
                fmt.Printf("scrub=%s\n", opts.scrub)
                fmt.Printf("queue=%s\n", opts.queue)
                fmt.Printf("scrubs=%s\n", opts.scrubs)
                fmt.Printf("invoices=%s\n", opts.invoices)
                fmt.Printf("report=%s\n", opts.report)
                exit(nil, 56, "options only valid for shell: -invoice, -scrub, -queue, -scrubs, -invoices, -report")
            }
        }
    }
    return opts
}

