package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/secondsightsolutions/binary-v4/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var server  api.BinarySvcClient

func connect() api.BinarySvcClient {
	target := fmt.Sprintf("%s:%s", host, port)
	if server != nil {
		return server
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{TLSCert},
		RootCAs:      X509pool,
	}
	crd := credentials.NewTLS(cfg)
	if conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(crd)); err == nil {
		server = api.NewBinarySvcClient(conn)
	} else {
		fmt.Printf("failed to connect to grpc server at %s:%s: %s\n", host, port, err.Error())
		os.Exit(27)
	}
	return server
}

func ping() {
    server  := connect()
    filter  := "filt/prid/="
    request := &api.GetDataReq{
        Attrs: map[string]string{
            "table/" : "rbtbin.proc",
            "pool/"  : "citus",
            "auth/"  : auth,
            "vers/"  : vers,
            filter   : name,
        },
    }
    if strm, err := server.GetData(context.Background(), request); err == nil {
        if _, err := strm.Recv(); err == nil {
            fmt.Println("successfully connected to server")
            exit(0, "")
            return
        } else if err != io.EOF {
            exit(7, "failed to receive ping response from server: %s", err.Error())
        } else {
            exit(8, "successfully connected to server: authorization provisioning required for (%s)", name)
        }
    } else {
        fmt.Printf("%T\n", err)
        exit(9, "failed to connect to server: %s", err.Error())
    }
}

func getData(pool, tbln, Manu string, filts map[string]string) chan data {
    server := connect()
	attrs  := map[string]string{}
	attrs["table/"] = tbln
	attrs["pool/"]  = pool
	attrs["auth/"]  = auth
	attrs["vers/"]  = vers
	attrs["manu/"]  = manu
	for k,v := range filts {
		attrs[k] = v
	}
	if Manu != "" {
		attrs["filt/manufacturer/="] = Manu
	}
    if strm, err := server.GetData(context.Background(), &api.GetDataReq{Attrs: attrs}); err == nil {
		chn := make(chan data, 500)
		go func(chan data) {
			for {
				if row, err := strm.Recv(); err == nil {
					row := row.Fields
					chn <- row
				} else if err == io.EOF {
					break
				} else {
					exit(17, "Cannot read data source %s:%s: %s", host, port, err.Error())
				}
			}
			close(chn)
		}(chn)
		return chn
    } else {
        exit(18, "cannot open grpc data source %s:%s: %s", host, port, err.Error())
    }
	return nil
}

func putData(pool, tbln, oper string) chan data {
    server := connect()
    if strm, err := server.PutData(context.Background()); err == nil {
		chn := make(chan data, 500)
		go func(chan data) {
			for row := range chn {
				row["table/"] = tbln
				row["pool/"]  = pool
				row["oper/"]  = oper
				row["auth/"]  = auth
				row["vers/"]  = vers
				row["manu/"]  = manu
				if err := strm.Send(&api.Data{Fields: row}); err != nil {
					exit(19, "cannot send to server: %s", err.Error())   // Requirement that failure to send to server causes immediate exit.
				}
			}
		}(chn)
		return chn
    } else {
        exit(18, "cannot open grpc data source %s:%s: %s", host, port, err.Error())
    }
	return nil
}

func createScrub() int64 {
	host, _ := os.Hostname()
	req := &api.NewScrubReq{
		Auth: auth,
		Manu: manu,
		Proc: X509cname(),
		Vers: vers,
		Dscr: desc,
		Hash: hash,
		Host: host,
		Line: strings.Join(os.Args, " "),
		Envr: envr,
	}
	server := connect()
    if res, err := server.NewScrub(context.Background(), req); err == nil {
        return res.Scid
    } else {
        exit(99, "cannot create scrub: %s", err.Error())
		return 99
    }
}

func update_scrub(sc *scrub, status string) error {
    plcy := ""
	hdrs := []string{}	// TODO
    status = strings.ReplaceAll(status, "\n", "")
	send := map[string]string{}
	send["table/"]      = "rbtbin.scrubs"
	send["oper/"]       = "update"
	send["pool/"]       = "citus"
    send["auth/"]       = auth
    send["vers/"]       = vers
    send["manu/"]       = manu
	send["filt/scrub_id/=/"] = fmt.Sprintf("%d", scid)
	send["scrub_id"]    = fmt.Sprintf("%d", scid)
    send["stat"]        = status
	send["hdrs"]        = strings.Join(hdrs, " ")
	send["policy"]      = plcy
	send["finished_at"] = time.Now().UTC().Format("2006-01-02 15:04:05.000000")
    
    if status == "" {
        sc.metr.Lock()
        send["rbt_total"]   = fmt.Sprintf("%d", sc.metr.rbt_total)
        send["rbt_valid"]   = fmt.Sprintf("%d", sc.metr.rbt_valid)
        send["rbt_matched"] = fmt.Sprintf("%d", sc.metr.rbt_matched)
        send["rbt_nomatch"] = fmt.Sprintf("%d", sc.metr.rbt_nomatch)
        send["rbt_invalid"] = fmt.Sprintf("%d", sc.metr.rbt_invalid)
        send["rbt_passed"]  = fmt.Sprintf("%d", sc.metr.rbt_passed)
        send["rbt_failed"]  = fmt.Sprintf("%d", sc.metr.rbt_failed)
        send["clm_total"]   = fmt.Sprintf("%d", sc.metr.clm_total)
        send["clm_valid"]   = fmt.Sprintf("%d", sc.metr.clm_valid)
        send["clm_matched"] = fmt.Sprintf("%d", sc.metr.clm_matched)
        send["clm_nomatch"] = fmt.Sprintf("%d", sc.metr.clm_nomatch)
        send["clm_invalid"] = fmt.Sprintf("%d", sc.metr.clm_invalid)
        send["spi_exact"]   = fmt.Sprintf("%d", sc.metr.spi_exact)
        send["spi_cross"]   = fmt.Sprintf("%d", sc.metr.spi_cross)
        send["spi_stack"]   = fmt.Sprintf("%d", sc.metr.spi_stack)
        send["spi_chain"]   = fmt.Sprintf("%d", sc.metr.spi_chain)
        send["dos_bef_doc"] = fmt.Sprintf("%d", sc.metr.dos_bef_doc)
        send["dos_aft_doc"] = fmt.Sprintf("%d", sc.metr.dos_aft_doc)
        send["dos_equ_dof"] = fmt.Sprintf("%d", sc.metr.dos_equ_dof)
        send["dos_bef_dof"] = fmt.Sprintf("%d", sc.metr.dos_bef_dof)
        send["dos_aft_dof"] = fmt.Sprintf("%d", sc.metr.dos_aft_dof)
        send["dos_equ_dof"] = fmt.Sprintf("%d", sc.metr.dos_equ_dof)
        sc.metr.Unlock()
    }
    
    server := connect()
    if grpcStrm, err := server.PutData(context.Background()); err == nil {
        err := grpcStrm.Send(&api.Data{Fields: send})
        grpcStrm.CloseSend()
        time.Sleep(time.Duration(3)*time.Second)    // Crazy bug in grpc. Writes to server are buffered internally - no way to flush!
        return err
    } else {
        return err
    }
}
