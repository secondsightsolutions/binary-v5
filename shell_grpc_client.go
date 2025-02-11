package main

import (
	context "context"
	"fmt"
	"io"
	"strings"
)

func (sh *Shell) ping() error {
	ctx := addMeta(context.Background(), sh.X509cert, nil)
	c,f := context.WithCancel(ctx)
	defer f()
	if _, err := sh.atlas.Ping(c, &Req{}); err == nil {
		fmt.Println("Pong!")
		return nil
	} else {
		fmt.Println(err.Error())
		return err
	}
}

func (sh *Shell) upload_invoice(file string) (int64, error) {
	ivid := int64(-1)

	if hdrs, chn, err := import_file[Rebate](file, ","); err == nil {
		ctx := addMeta(context.Background(), sh.X509cert, map[string]string{
			"file": file,
			"hdrs": strings.Join(hdrs, ","),
		})
		c,f := context.WithCancel(ctx)
		defer f()
		if strm, err := sh.atlas.UploadInvoice(c); err == nil {
			if hdr, err := strm.Header(); err == nil {
				ivid = metaValueInt64(hdr, "ivid")
			} else {
				fmt.Println("failed reading header")
				return -1, err
			}
			for rbt := range chn {
				if err := strm.Send(rbt); err != nil {
					return ivid, err
				}
			}
			strm.CloseSend()
			strm.CloseAndRecv()
		} else {
			fmt.Printf("shell.upload_invoice(): atlas.UploadInvoice() failed: %s\n", err.Error())
		}
	} else {
		fmt.Printf("shell.upload_invoice(): import_file failed: %s\n", err.Error())
	}
	return ivid, nil
}

func (sh *Shell) run_scrub(ivid int64) (int64, error) {
	ctx := addMeta(context.Background(), sh.X509cert, map[string]string{
		"plcy": sh.opts.policy,
		"kind": sh.opts.kind,
		"test": "",
	})
	c,f := context.WithCancel(ctx)
	defer f()
	req := &InvoiceIdent{Manu: manu, Ivid: ivid}
	scid := int64(-1)

	if strm, err := sh.atlas.RunScrub(c, req); err == nil {
		if hdr, err := strm.Header(); err == nil {
			scid = metaValueInt64(hdr, "scid")
		} else {
			return scid, err
		}
		for {
			if met, err := strm.Recv(); err == nil {
				fmt.Printf("%v\n", met)
			} else if err == io.EOF {
				return scid, nil
			} else {
				return scid, err
			}
		}
	}
	return scid, nil
}

func (sh *Shell) run_queue(ivid int64) (int64, error) {
	ctx := addMeta(context.Background(), sh.X509cert, map[string]string{
		"plcy": sh.opts.policy,
		"kind": sh.opts.kind,
		"test": "",
	})
	c,f := context.WithCancel(ctx)
	defer f()
	req := &InvoiceIdent{Manu: manu, Ivid: ivid}

	if res, err := sh.atlas.RunQueue(c, req); err == nil {
		return res.Scid, nil
	} else {
		return -1, err
	}
}

func (sh *Shell) get_scrub(scrubs string) (*Scrub, error) {
	return nil, nil
}
func (sh *Shell) get_scrub_metrics(invoices string) (*Metrics, error) {
	return nil, nil
}
func (sh *Shell) get_scrub_rebates(invoices string) ([]*ScrubRebate, error) {
	return nil, nil
}
func (sh *Shell) get_scrub_file(scid int64) (int64, error) {
	ctx := context.Background()
	c,f := context.WithCancel(ctx)
	defer f()
	req := &ScrubIdent{Manu: manu, Scid: scid}

	if strm, err := sh.atlas.GetScrubFile(c, req); err == nil {
		for {
			if rr, err := strm.Recv(); err == nil {
				fmt.Printf("%v\n", rr)
			} else if err == io.EOF {
				return scid, nil
			} else {
				return scid, err
			}
		}
	}
	return scid, nil
}

func (sh *Shell) get_invoice(ivid int64) (int64, error) {
	return 0, nil
}

func (sh *Shell) get_invoice_rebates(invoice string) error {
	return nil
}

func (sh *Shell) upload_test(name, dir string) error {
	return nil
}
