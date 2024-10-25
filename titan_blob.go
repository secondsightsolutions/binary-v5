package main

import (
	context "context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/jackc/pgx/v5/pgxpool"
)

func run_save_to_azure(wg *sync.WaitGroup, stop chan any, intv int, account, key string) {
	defer wg.Done()
	for {
		select {
		case <-time.After(time.Duration(intv) * time.Second):
			saveToAzure(account, key)
		case <-stop:
			return
		}
	}
}
func run_save_to_datab(wg *sync.WaitGroup, stop chan any, intv int, account, key string, pools map[string]*pgxpool.Pool) {
	defer wg.Done()
	for {
		select {
		case <-time.After(time.Duration(intv) * time.Second):
			saveToDatabase(account, key, pools)
		case <-stop:
			return
		}
	}
}

func newBlobClient(account, key string) (*azblob.Client, error) {
    url := fmt.Sprintf("https://%s.blob.core.windows.net", account)
    if cred, err := azblob.NewSharedKeyCredential(account, key); err == nil {
        return azblob.NewClientWithSharedKeyCredential(url, cred, nil)
    } else {
        return nil, err
    }
}
func newBlockBlobClient(account, key, container, name string) (*blockblob.Client, error) {
    url := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", account, container, name)
    if cred, err := azblob.NewSharedKeyCredential(account, key); err == nil {
        return blockblob.NewClientWithSharedKeyCredential(url, cred, nil)
    } else {
        return nil, err
    }
}
func newQueueClient(account, key, que string) (*azqueue.QueueClient, error) {
    url := fmt.Sprintf("https://%s.queue.core.windows.net", account)
    if cred, err := azqueue.NewSharedKeyCredential(account, key); err == nil {
        if svc, err := azqueue.NewServiceClientWithSharedKeyCredential(url, cred, nil); err == nil {
            svc.CreateQueue(context.Background(), que, nil)
            return svc.NewQueueClient(que), nil
        } else {
            return nil, err
        } 
    } else {
        return nil, err
    }
}

func saveToAzure(account, key string) {
	readScrubDir := func(root string) []string {
		list := []string{}
		if dirs, err := os.ReadDir(root); err == nil {
			for _, dir := range dirs {									// dir:  "111_12345678_brg_amgen"
				prnt := fmt.Sprintf("%s/%s", root, dir.Name())			// prnt: "root/111_12345678_brg_amgen"
				if files, err := os.ReadDir(prnt); err == nil {
					for _, file := range files {						// file: "citus_insert_rbtbin.attempts"
						full := fmt.Sprintf("%s/%s", prnt, file.Name()) // full: "root/111_12345678_brg_amgen/citus_insert_rbtbin.attempts"
						list = append(list, full)
					}
				}
			}
		}
		return list
	}
	scrubDirParse := func(str string) (scid, secs, proc, manu string) {
		// 111_12345678_brg_novo-nordisk
		toks := strings.Split(str, "_")
		if len(toks) >= 1 {
			scid = toks[0]
			if len(toks) >= 2 {
				secs = toks[1]
				if len(toks) >= 3 {
					proc = strings.ReplaceAll(toks[2], "-", "_")
					if len(toks) >= 4 {
						manu = strings.ReplaceAll(toks[3], "-", "_")
					}
				}
			}
		}
		return
	}
	scrubFileParse := func(str string) (pool, oper, tbln string) {
		toks := strings.Split(str, "_")
		if len(toks) >= 1 {
			pool = toks[0]
			if len(toks) >= 2 {
				oper = toks[1]
				if len(toks) >= 3 {
					tbln = strings.ReplaceAll(toks[2], "-", "_")
				}
			}
		}
		return
	}
	parsePath := func(path string) (root, dir, file string) {
		toks := strings.SplitN(path, "/", 3)
		root = toks[0]
		dir  = toks[1]
		file = toks[2]
		return
	}
	makeMeta := func(manu, proc, scid, pool, oper, tbln, secs string) map[string]string {
		meta := map[string]string{}
		meta["manu"]  = manu
		meta["proc"]  = proc
        meta["scid"]  = scid
        meta["pool"]  = pool
        meta["oper"]  = oper
        meta["tbln"]  = tbln
        meta["crat"]  = secs
        meta["dbat"]  = ""
		return meta
	}
	makeTags := func(manu, proc, scid, secs string) map[string]string {
		tags := map[string]string{}
		sctm := time.Unix(StrDecToInt64(secs), 0).UTC()
		tags["manu"]  = manu
		tags["proc"]  = proc
		tags["scid"]  = scid
		tags["date"]  = sctm.Format("2006-01-02")
		tags["time"]  = sctm.Format("15:04:05")
        tags["day"]   = sctm.Format("02")
		tags["month"] = sctm.Format("01")
		tags["year"]  = sctm.Format("2006")
        tags["indb"]  = "false"
		return tags
	}
    uploadBlob := func(account, key, container, blob, full string, meta, tags map[string]string) error {
        strt := time.Now()
        if fd, err := os.Open(full); err == nil {
            defer fd.Close()
            metp := map[string]*string{}
            for k,v := range meta {
                str := string(v)
                metp[k] = &str
            }
            if client, err := newBlobClient(account, key); err == nil {
                opts := &azblob.UploadFileOptions{
                    BlockSize:      0,
                    Metadata:       metp,
                    Tags:           tags,
                    Concurrency:    2,
                }
                client.CreateContainer(context.Background(), container, nil)
                if _, err := client.UploadFile(context.Background(), container, blob, fd, opts); err != nil {
					log("service", "blob upload", "acct=%s cntr=%s blob=%s file=%s: upload file failed", time.Since(strt), err, account, container, blob, full)
                    return err
                }
            } else {
                log("service", "blob upload", "acct=%s cntr=%s blob=%s file=%s: create az client failed", time.Since(strt), err, account, container, blob, full)
                return err
            }
        } else {
            log("service", "blob upload", "acct=%s cntr=%s blob=%s file=%s: open local file failed", time.Since(strt), err, account, container, blob, full)
            return err
        }
        log("service", "blob upload", "acct=%s cntr=%s blob=%s file=%s: upload file succeeded", time.Since(strt), nil, account, container, blob, full)
        return nil
    }
    sendMessage := func(cntr, blob string) error {
        if qc, err := newQueueClient(account, key, "todb"); err == nil {
            _, err := qc.EnqueueMessage(context.Background(), cntr + "/" + blob, nil)
            return err
        } else {
            return err
        }
    }

	root := "to_azure"
	list := readScrubDir(root)
	dirs := map[string]any{}
    keep := map[string]any{}
    blbP := 0
    blbF := 0
    msgP := 0
    msgF := 0
    strt := time.Now()
    if len(list) == 0 {
        return
    }
	for _, full := range list {							// full: "root/111_12345678_brg_amgen/citus_insert_rbtbin.attempts"
		_, dir, file := parsePath(full)					// dir:  "111_12345678_brg_amgen" file: "citus_insert_rbtbin.attempts"
		scid, secs, proc, manu := scrubDirParse(dir)	// scid: "111" secs: "12345678", proc: "brg" manu: "amgen"
        pool, oper, tbln := scrubFileParse(file)        // pool: "citus" oper: "insert" tbln: "rbtbin.attempts"
		dirs[dir] = nil
		cntr := proc									// cntr: "johnson_n_johnson"
        cntr = strings.ReplaceAll(cntr, "_", "-")       // Azure blob container name limitation ("johnson-n-johnson").
        cntr = strings.ToLower(cntr)                    // Azure blob container name limitation.
        blob := fmt.Sprintf("%s_%s_%s.csv", scid, oper, tbln)
		meta := makeMeta(manu, proc, scid, pool, oper, tbln, secs)
		tags := makeTags(manu, proc, scid, secs)
		if err := uploadBlob(account, key, cntr, blob, full, meta, tags); err == nil {
            blbP++
            if sendMessage(cntr, blob) == nil {
                msgP++
            } else {
                msgF++
            }
			if err := os.Remove(full); err != nil {
				log("service", "SaveToAzure", "failed to remove file (%s)", time.Since(strt), err, full)
			}
		} else {
			log("service", "SaveToAzure", "blob upload failed (%s)", time.Since(strt), err, full)
            blbF++
            keep[dir] = nil    // At least one upload failed. Do not delete the parent directory (below)
        }
	}
    log("service", "SaveToAzure", "blob upload status - good (%d) fail (%d)", time.Since(strt), nil, blbP, blbF)
    log("service", "SaveToAzure", "msgs upload status - good (%d) fail (%d)", time.Since(strt), nil, msgP, msgF)
	for dir := range dirs {
        if _,ok := keep[dir];!ok {  // Only remove dirs that have not been marked for keeping (os.Remove would probably fail anyway...)
            strt := time.Now()
            full := root + "/" + dir
            if err := os.Remove(full); err != nil {
                log("service", "SaveToAzure", "failed to remove directory (%s)", time.Since(strt), err, full)
            }
        }
	}
    return
}

func saveToDatabase(account, key string, pools map[string]*pgxpool.Pool) error {
    setBlobSaved := func(cntr, blob string) error {
        strt := time.Now()
        if bc, err := newBlockBlobClient(account, key, cntr, blob); err == nil {
            if resp, err := bc.GetTags(context.Background(), nil); err == nil {
                tags := map[string]string{}
                for _, tag := range resp.BlobTagSet {
                    tags[*tag.Key] = *tag.Value
                }
                tags["indb"] = "true"
                if _, err := bc.SetTags(context.Background(), tags, nil); err != nil {
                    log("service", "SaveToDatabase", "acct=%s cntr=%s blob=%s: set blob tags failed", time.Since(strt), err, account, cntr, blob)
                    return err
                }
            } else {
                log("service", "SaveToDatabase", "acct=%s cntr=%s blob=%s: get blob tags failed", time.Since(strt), err, account, cntr, blob)
                return err
            }
            if props, err := bc.GetProperties(context.Background(), nil); err == nil { 
                meta := map[string]*string{}
                now  := fmt.Sprintf("%d", time.Now().Unix())
                for k,vp := range props.Metadata {
                    meta[k] = vp
                }
                meta["Dbat"] = &now
                if _, err := bc.SetMetadata(context.Background(), meta, nil); err != nil {
                    log("service", "SaveToDatabase", "acct=%s cntr=%s blob=%s: set blob metadata failed", time.Since(strt), err, account, cntr, blob)
                }
            } else {
                log("service", "SaveToDatabase", "acct=%s cntr=%s blob=%s: get blob metadata failed", time.Since(strt), err, account, cntr, blob)
                return err
            }
        } else {
            log("service", "SaveToDatabase", "acct=%s cntr=%s blob=%s: create blob client failed", time.Since(strt), err, account, cntr, blob)
            return err
        }
        return nil
    }
    downloadBlob := func(account, key, container, blob, file string) error {
        strt := time.Now()
        if fd, err := os.Create(file); err == nil {
            if client, err := newBlobClient(account, key); err == nil {
                opts := &azblob.DownloadFileOptions{
                    BlockSize:      4096,
                    Concurrency:    4,
                }
                if _, err := client.DownloadFile(context.Background(), container, blob, fd, opts); err == nil {
                    log("service", "SaveToDatabase", "acct=%s cntr=%s blob=%s downloaded file", time.Since(strt), nil, account, container, blob)
                    return nil
                }
                log("service", "SaveToDatabase", "acct=%s cntr=%s blob=%s: download blob failed", time.Since(strt), err, account, container, blob)
                return err
            } else {
                log("service", "SaveToDatabase", "acct=%s cntr=%s blob=%s: create az client failed", time.Since(strt), err, account, container, blob)
                return err
            }
        } else {
            log("service", "SaveToDatabase", "acct=%s cntr=%s blob=%s file=%s: create local file failed", time.Since(strt), err, account, container, blob, file)
            return err
        }
    }
    getMetadata := func(account, key, full string) (map[string]string, string, string, error) {
        strt := time.Now()
        meta := map[string]string{}
        toks := strings.SplitN(full, "/", 2)
        if len(toks) != 2 {
            return nil, "", "", fmt.Errorf("badly formatted blob name")
        }
        cntr := toks[0]
        blob := toks[1]
        if bc, err := newBlockBlobClient(account, key, cntr, blob); err == nil {
            if props, err := bc.GetProperties(context.Background(), nil); err == nil {
                for k,vp := range props.Metadata {
                    meta[k] = *vp
                }
            } else {
                log("service", "SaveToDatabase", "acct=%s cntr=%s blob=%s: get blob metadata failed", time.Since(strt), err, account, cntr, blob)
                return nil, cntr, blob, err
            }
        } else {
            log("service", "SaveToDatabase", "acct=%s cntr=%s blob=%s: create blob client failed", time.Since(strt), err, account, cntr, blob)
            return nil, cntr, blob, err
        }
        return meta, cntr, blob, nil
    }
    insertRows := func(ctx context.Context, pool *pgxpool.Pool, tbln, file string) error {
        return db_copyfrom(ctx, pool, tbln, file, 100000)
    }
    updateRows := func(ctx context.Context, pool *pgxpool.Pool, file string) error {
        return db_updates(ctx, pool, file)
    }
    
    os.Mkdir("to_datab", os.ModePerm)   // Okay if it fails (means it already exists, most likely).
    strt := time.Now()
    if q, err := newQueueClient(account, key, "todb"); err == nil {
        vto  := int32(60*30)    // 30 minute timeout.
        opts := &azqueue.DequeueMessageOptions{VisibilityTimeout: &vto}
        for {
            strt := time.Now()
            if resp, err := q.DequeueMessage(context.Background(), opts); err == nil {
                if len(resp.Messages) == 0 {
                    break
                }
                for _, msg := range resp.Messages {     // Should only be one message here. One at a time.
                    full := string(*msg.MessageText)
                    log("service", "SaveToDatabase", "read message with %s", time.Since(strt), nil, full)
                    if meta, cntr, blob, err := getMetadata(account, key, full); err == nil {
                        pln  := meta["Pool"]    // Apparently the AZ SDK capitalizes the first letter!
                        oper := meta["Oper"]
                        tbln := meta["Tbln"]
						pool := pools[pln]
                        if strings.HasSuffix(tbln, "attempts") {
                            q.DeleteMessage(context.Background(), *msg.MessageID, *msg.PopReceipt, nil)
                            continue
                        }
                        file := "to_datab/_downloaded"  // Since we only deal with one blob at a time, we just need the one file.
                        if err := downloadBlob(account, key, cntr, blob, file); err == nil {
							if oper == "insert" {
								err = insertRows(context.Background(), pool, tbln, file)
							} else {
								err = updateRows(context.Background(), pool, file)
							}
							if err == nil {
								// Delete message from queue if we were able to insert to database successfully.
								setBlobSaved(cntr, blob)
								q.DeleteMessage(context.Background(), *msg.MessageID, *msg.PopReceipt, nil)
							}
							// If we didn't insert to database, then just leave the message. It's visibility will stay hidden for
							// five minutes, after which we will try again (because the message becomes visible again).
							// The invisible message(s) is basically the DLQ, but managed automatically by the infra.
							// The TTL on the message itself is seven days. So we have seven days to fix the problem before losing the data.
							log("service", "SaveToDatabase", "acct=%s cntr=%s blob=%s oper=%s: completed", time.Since(strt), err, account, cntr, blob, oper)
							os.Remove(file)
                        }
                    }
                }
            }
        }
    } else {
		log("service", "SaveToDatabase", "acct=%s cannot create queue client", time.Since(strt), err, account)
        return err
    }
    return nil
}
