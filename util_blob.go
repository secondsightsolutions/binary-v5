package main

// import (
// 	context "context"
// 	"fmt"
// 	"os"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
// 	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
// 	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
// )

// func run_save_to_azure(wg *sync.WaitGroup, stop chan any, intv int, account, key string) {
// 	defer wg.Done()
// 	for {
// 		select {
// 		case <-time.After(time.Duration(intv) * time.Second):
// 			saveToAzure(account, key)
// 		case <-stop:
// 			return
// 		}
// 	}
// }

// func newBlobClient(account, key string) (*azblob.Client, error) {
//     url := fmt.Sprintf("https://%s.blob.core.windows.net", account)
//     if cred, err := azblob.NewSharedKeyCredential(account, key); err == nil {
//         return azblob.NewClientWithSharedKeyCredential(url, cred, nil)
//     } else {
//         return nil, err
//     }
// }
// func newBlockBlobClient(account, key, container, name string) (*blockblob.Client, error) {
//     url := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", account, container, name)
//     if cred, err := azblob.NewSharedKeyCredential(account, key); err == nil {
//         return blockblob.NewClientWithSharedKeyCredential(url, cred, nil)
//     } else {
//         return nil, err
//     }
// }
// func newQueueClient(account, key, que string) (*azqueue.QueueClient, error) {
//     url := fmt.Sprintf("https://%s.queue.core.windows.net", account)
//     if cred, err := azqueue.NewSharedKeyCredential(account, key); err == nil {
//         if svc, err := azqueue.NewServiceClientWithSharedKeyCredential(url, cred, nil); err == nil {
//             svc.CreateQueue(context.Background(), que, nil)
//             return svc.NewQueueClient(que), nil
//         } else {
//             return nil, err
//         } 
//     } else {
//         return nil, err
//     }
// }

// func saveToAzure(account, key string) {
// 	readScrubDir := func(root string) []string {
// 		list := []string{}
// 		if dirs, err := os.ReadDir(root); err == nil {
// 			for _, dir := range dirs {									// dir:  "111_12345678_brg_amgen"
// 				prnt := fmt.Sprintf("%s/%s", root, dir.Name())			// prnt: "root/111_12345678_brg_amgen"
// 				if files, err := os.ReadDir(prnt); err == nil {
// 					for _, file := range files {						// file: "citus_insert_rbtbin.attempts"
// 						full := fmt.Sprintf("%s/%s", prnt, file.Name()) // full: "root/111_12345678_brg_amgen/citus_insert_rbtbin.attempts"
// 						list = append(list, full)
// 					}
// 				}
// 			}
// 		}
// 		return list
// 	}
// 	scrubDirParse := func(str string) (scid, secs, proc, manu string) {
// 		// 111_12345678_brg_novo-nordisk
// 		toks := strings.Split(str, "_")
// 		if len(toks) >= 1 {
// 			scid = toks[0]
// 			if len(toks) >= 2 {
// 				secs = toks[1]
// 				if len(toks) >= 3 {
// 					proc = strings.ReplaceAll(toks[2], "-", "_")
// 					if len(toks) >= 4 {
// 						manu = strings.ReplaceAll(toks[3], "-", "_")
// 					}
// 				}
// 			}
// 		}
// 		return
// 	}
// 	scrubFileParse := func(str string) (pool, oper, tbln string) {
// 		toks := strings.Split(str, "_")
// 		if len(toks) >= 1 {
// 			pool = toks[0]
// 			if len(toks) >= 2 {
// 				oper = toks[1]
// 				if len(toks) >= 3 {
// 					tbln = strings.ReplaceAll(toks[2], "-", "_")
// 				}
// 			}
// 		}
// 		return
// 	}
// 	parsePath := func(path string) (root, dir, file string) {
// 		toks := strings.SplitN(path, "/", 3)
// 		root = toks[0]
// 		dir  = toks[1]
// 		file = toks[2]
// 		return
// 	}
// 	makeMeta := func(manu, proc, scid, pool, oper, tbln, secs string) map[string]string {
// 		meta := map[string]string{}
// 		meta["manu"]  = manu
// 		meta["proc"]  = proc
//         meta["scid"]  = scid
//         meta["pool"]  = pool
//         meta["oper"]  = oper
//         meta["tbln"]  = tbln
//         meta["crat"]  = secs
//         meta["dbat"]  = ""
// 		return meta
// 	}
// 	makeTags := func(manu, proc, scid, secs string) map[string]string {
// 		tags := map[string]string{}
// 		sctm := time.Unix(StrDecToInt64(secs), 0).UTC()
// 		tags["manu"]  = manu
// 		tags["proc"]  = proc
// 		tags["scid"]  = scid
// 		tags["date"]  = sctm.Format("2006-01-02")
// 		tags["time"]  = sctm.Format("15:04:05")
//         tags["day"]   = sctm.Format("02")
// 		tags["month"] = sctm.Format("01")
// 		tags["year"]  = sctm.Format("2006")
//         tags["indb"]  = "false"
// 		return tags
// 	}
//     uploadBlob := func(account, key, container, blob, full string, meta, tags map[string]string) error {
//         strt := time.Now()
//         if fd, err := os.Open(full); err == nil {
//             defer fd.Close()
//             metp := map[string]*string{}
//             for k,v := range meta {
//                 str := string(v)
//                 metp[k] = &str
//             }
//             if client, err := newBlobClient(account, key); err == nil {
//                 opts := &azblob.UploadFileOptions{
//                     BlockSize:      0,
//                     Metadata:       metp,
//                     Tags:           tags,
//                     Concurrency:    2,
//                 }
//                 client.CreateContainer(context.Background(), container, nil)
//                 if _, err := client.UploadFile(context.Background(), container, blob, fd, opts); err != nil {
// 					log("titan", "blob upload", "acct=%s cntr=%s blob=%s file=%s: upload file failed", time.Since(strt), err, account, container, blob, full)
//                     return err
//                 }
//             } else {
//                 log("titan", "blob upload", "acct=%s cntr=%s blob=%s file=%s: create az client failed", time.Since(strt), err, account, container, blob, full)
//                 return err
//             }
//         } else {
//             log("titan", "blob upload", "acct=%s cntr=%s blob=%s file=%s: open local file failed", time.Since(strt), err, account, container, blob, full)
//             return err
//         }
//         log("titan", "blob upload", "acct=%s cntr=%s blob=%s file=%s: upload file succeeded", time.Since(strt), nil, account, container, blob, full)
//         return nil
//     }
//     sendMessage := func(cntr, blob string) error {
//         if qc, err := newQueueClient(account, key, "todb"); err == nil {
//             _, err := qc.EnqueueMessage(context.Background(), cntr + "/" + blob, nil)
//             return err
//         } else {
//             return err
//         }
//     }

// 	root := "to_azure"
// 	list := readScrubDir(root)
// 	dirs := map[string]any{}
//     keep := map[string]any{}
//     blbP := 0
//     blbF := 0
//     msgP := 0
//     msgF := 0
//     strt := time.Now()
//     if len(list) == 0 {
//         return
//     }
// 	for _, full := range list {							// full: "root/111_12345678_brg_amgen/citus_insert_rbtbin.attempts"
// 		_, dir, file := parsePath(full)					// dir:  "111_12345678_brg_amgen" file: "citus_insert_rbtbin.attempts"
// 		scid, secs, proc, manu := scrubDirParse(dir)	// scid: "111" secs: "12345678", proc: "brg" manu: "amgen"
//         pool, oper, tbln := scrubFileParse(file)        // pool: "citus" oper: "insert" tbln: "rbtbin.attempts"
// 		dirs[dir] = nil
// 		cntr := proc									// cntr: "johnson_n_johnson"
//         cntr = strings.ReplaceAll(cntr, "_", "-")       // Azure blob container name limitation ("johnson-n-johnson").
//         cntr = strings.ToLower(cntr)                    // Azure blob container name limitation.
//         blob := fmt.Sprintf("%s_%s_%s.csv", scid, oper, tbln)
// 		meta := makeMeta(manu, proc, scid, pool, oper, tbln, secs)
// 		tags := makeTags(manu, proc, scid, secs)
// 		if err := uploadBlob(account, key, cntr, blob, full, meta, tags); err == nil {
//             blbP++
//             if sendMessage(cntr, blob) == nil {
//                 msgP++
//             } else {
//                 msgF++
//             }
// 			if err := os.Remove(full); err != nil {
// 				log("titan", "SaveToAzure", "failed to remove file (%s)", time.Since(strt), err, full)
// 			}
// 		} else {
// 			log("titan", "SaveToAzure", "blob upload failed (%s)", time.Since(strt), err, full)
//             blbF++
//             keep[dir] = nil    // At least one upload failed. Do not delete the parent directory (below)
//         }
// 	}
//     log("titan", "SaveToAzure", "blob upload status - good (%d) fail (%d)", time.Since(strt), nil, blbP, blbF)
//     log("titan", "SaveToAzure", "msgs upload status - good (%d) fail (%d)", time.Since(strt), nil, msgP, msgF)
// 	for dir := range dirs {
//         if _,ok := keep[dir];!ok {  // Only remove dirs that have not been marked for keeping (os.Remove would probably fail anyway...)
//             strt := time.Now()
//             full := root + "/" + dir
//             if err := os.Remove(full); err != nil {
//                 log("titan", "SaveToAzure", "failed to remove directory (%s)", time.Since(strt), err, full)
//             }
//         }
// 	}
// }
