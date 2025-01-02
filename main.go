package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	atlas_grpc_port int = 23460
	titan_grpc_port int = 23461
	opts *Opts
)

func main() {
	var done sync.WaitGroup

	stop := make(chan any)
	catchSignals(stop)

	getEnv()
	opts = options()

	if opts.runPing {
		ping()
		exit(nil, 0, "")
	} else if opts.runVers {
		version()
		exit(nil, 0, "")
	} else if opts.runConf {
		config()
		exit(nil, 0, "")
	}

	if opts.runTitan {
		run_titan(&done, opts, stop)
	}
	if opts.runAtlas {
		run_atlas(&done, opts, stop)
	}
	if opts.runClient {
		run_shell(opts, stop)
	}
	done.Wait()
}

func catchSignals(stop chan any) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-sigs
		Log("main", "catchSignals", "", "signal received, starting shutdown", 0, nil, nil)
		close(stop)
	}()
}

func getEnv() {
	if name == "brg" {
		setIf(&titan_host, "V5_TITAN_HOST")
		setIf(&titan_port, "V5_TITAN_PORT")
		setIf(&titan_name, "V5_TITAN_NAME")
		setIf(&titan_user, "V5_TITAN_USER")
		setIf(&titan_pass, "V5_TITAN_PASS")

		setIf(&atlas_host, "V5_ATLAS_HOST")
		setIf(&atlas_port, "V5_ATLAS_PORT")
		setIf(&atlas_name, "V5_ATLAS_NAME")
		setIf(&atlas_user, "V5_ATLAS_USER")
		setIf(&atlas_pass, "V5_ATLAS_PASS")

		setIf(&citus_host, "V5_CITUS_HOST")
		setIf(&citus_port, "V5_CITUS_PORT")
		setIf(&citus_name, "V5_CITUS_NAME")
		setIf(&citus_user, "V5_CITUS_USER")
		setIf(&citus_pass, "V5_CITUS_PASS")
		setIf(&espdb_host, "V5_ESPDB_HOST")
		setIf(&espdb_port, "V5_ESPDB_PORT")
		setIf(&espdb_name, "V5_ESPDB_NAME")
		setIf(&espdb_user, "V5_ESPDB_USER")
		setIf(&espdb_pass, "V5_ESPDB_PASS")

		setIf(&titan_grpc, "V5_TITAN_GRPC")
		setIf(&atlas_grpc, "V5_ATLAS_GRPC")
		setIf(&atlas_ogtm, "V5_ATLAS_OGTM")
		setIf(&titan_ogtm, "V5_TITAN_OGTM")
		setIf(&atlas_ogky, "V5_ATLAS_OGKY")
		setIf(&titan_ogky, "V5_TITAN_OGKY")

		setIf(&shell_cert, "V5_SHELL_CERT")
		setIf(&atlas_cert, "V5_ATLAS_CERT")
		setIf(&titan_cert, "V5_TITAN_CERT")
		setIf(&shell_pkey, "V5_SHELL_PKEY")
		setIf(&atlas_pkey, "V5_ATLAS_PKEY")
		setIf(&titan_pkey, "V5_TITAN_PKEY")

	} else {
		if Type == "manu" {
			setIf(&titan_grpc, "V5_TITAN_GRPC")

			setIf(&atlas_host, "V5_ATLAS_HOST")
			setIf(&atlas_port, "V5_ATLAS_PORT")
			setIf(&atlas_name, "V5_ATLAS_NAME")
			setIf(&atlas_user, "V5_ATLAS_USER")
			setIf(&atlas_pass, "V5_ATLAS_PASS")

			setIf(&atlas_grpc, "V5_ATLAS_GRPC")
			setIf(&atlas_ogtm, "V5_ATLAS_OGTM")
			setIf(&atlas_ogky, "V5_ATLAS_OGKY")

			setIf(&atlas_cert, "V5_ATLAS_CERT")
			setIf(&atlas_pkey, "V5_ATLAS_PKEY")

			setIf(&shell_cert, "V5_SHELL_CERT")
			setIf(&shell_pkey, "V5_SHELL_PKEY")

		} else {
			setIf(&atlas_grpc, "V5_ATLAS_GRPC")

			setIf(&shell_cert, "V5_SHELL_CERT")
			setIf(&shell_pkey, "V5_SHELL_PKEY")
		}
	}
}

func setIf(envVar *string, envName string) {
	if envVal := os.Getenv(envName); envVal != "" {
		if *envVar != envVal {
			fmt.Printf("%s changing from %s to %s\n", envName, *envVar, envVal)
			*envVar = envVal
		}
	}
}

func version() {
	fmt.Printf("%s: %s\n", "name", name)
	fmt.Printf("%s: %s\n", "desc", desc)
	fmt.Printf("%s: %s\n", "type", Type)
	fmt.Printf("%s: %s\n", "envr", envr)
	fmt.Printf("%s: %s\n", "vers", vers)
	fmt.Printf("%s: %s\n", "hash", hash)
	fmt.Printf("%s: %s\n", "manu", manu)
}

func config() {
	if _, xcrt, err := CryptInit(titan_cert, cacr, "", titan_pkey, salt, phrs); err == nil {
		fmt.Printf("%s: %s\n", "titan_name", X509cname(xcrt))
		fmt.Printf("%s: %s\n", "titan_full", X509ou(xcrt))
	} else {
		fmt.Printf("%s: %s\n", "titan_name", "")
		fmt.Printf("%s: %s\n", "titan_full", "")
	}
	fmt.Printf("%s: %s\n", "titan_grpc", titan_grpc)
	fmt.Printf("%s: %s\n", "titan_host", titan_host)
	fmt.Printf("%s: %s\n", "titan_port", titan_port)
	fmt.Printf("%s: %s\n", "titan_name", titan_name)
	fmt.Printf("%s: %s\n", "titan_user", titan_user)
	fmt.Printf("%s: %s\n", "titan_pass", titan_pass)
	fmt.Printf("%s: %s\n", "titan_ogtm", titan_ogtm)
	fmt.Printf("%s: %s\n", "titan_ogky", titan_ogky)
	fmt.Printf("%s: %s\n", "citus_host", citus_host)
	fmt.Printf("%s: %s\n", "citus_port", citus_port)
	fmt.Printf("%s: %s\n", "citus_name", citus_name)
	fmt.Printf("%s: %s\n", "citus_user", citus_user)
	fmt.Printf("%s: %s\n", "citus_pass", citus_pass)
	fmt.Printf("%s: %s\n", "espdb_host", espdb_host)
	fmt.Printf("%s: %s\n", "espdb_port", espdb_port)
	fmt.Printf("%s: %s\n", "espdb_name", espdb_name)
	fmt.Printf("%s: %s\n", "espdb_user", espdb_user)
	fmt.Printf("%s: %s\n", "espdb_pass", espdb_pass)
	fmt.Printf("%s: %s\n", "titan_cert", titan_cert)
	fmt.Printf("%s: %s\n", "titan_pkey", titan_pkey)

	if _, xcrt, err := CryptInit(atlas_cert, cacr, "", atlas_pkey, salt, phrs); err == nil {
		fmt.Printf("%s: %s\n", "atlas_name", X509cname(xcrt))
		fmt.Printf("%s: %s\n", "atlas_full", X509ou(xcrt))
	} else {
		fmt.Printf("%s: %s\n", "atlas_name", "")
		fmt.Printf("%s: %s\n", "atlas_full", "")
	}
	fmt.Printf("%s: %s\n", "atlas_grpc", atlas_grpc)
	fmt.Printf("%s: %s\n", "atlas_host", atlas_host)
	fmt.Printf("%s: %s\n", "atlas_port", atlas_port)
	fmt.Printf("%s: %s\n", "atlas_name", atlas_name)
	fmt.Printf("%s: %s\n", "atlas_user", atlas_user)
	fmt.Printf("%s: %s\n", "atlas_pass", atlas_pass)
	fmt.Printf("%s: %s\n", "atlas_ogtm", atlas_ogtm)
	fmt.Printf("%s: %s\n", "atlas_ogky", atlas_ogky)
	fmt.Printf("%s: %s\n", "atlas_cert", atlas_cert)
	fmt.Printf("%s: %s\n", "atlas_pkey", atlas_pkey)

	if _, xcrt, err := CryptInit(shell_cert, cacr, "", shell_pkey, salt, phrs); err == nil {
		fmt.Printf("%s: %s\n", "shell_name", X509cname(xcrt))
		fmt.Printf("%s: %s\n", "shell_full", X509ou(xcrt))
	} else {
		fmt.Printf("%s: %s\n", "shell_name", "")
		fmt.Printf("%s: %s\n", "shell_full", "")
	}
	fmt.Printf("%s: %s\n", "shell_cert", shell_cert)
	fmt.Printf("%s: %s\n", "shell_pkey", shell_pkey)
}