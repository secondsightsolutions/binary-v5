package main

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type scrub struct {
    sr *scrub_req
    cs map[string]*cache
    plcy *policy
}

type scrub_req struct {
    auth string
    manu string
    sort string         // The sort order for rebates (defaults to "indx", the generated ordinal position in the rebate file).
    uniq string         // How we divide across rebate procs. All rebates with same value for "uniq" must go to same rebate proc. Defaults to "dos".
    files map[string]*scrub_file // The set of input files (rebates, claims, LU tables).
}

type scrub_file struct {
    name string         // The name to use for the cache, like "rebates", "claims", "ndcs", etc.
    file string         // Optional disk file name. Used by test code, http does not use it.
    hdrs string         // If set, the mapping from input hdrs to attr names (hdr1=attr1;hdr2=attr2)
    csep string         // If set, the hdr/col separator in the input (defaults to ",")
    keyn string         // If set, the attr name of the key (the index). Can be more than one (rxn;frxn).
    keyl int            // The length of the key. Typically 3, meaning first three chars of key value (negative means start from end of key).
    rdr  io.Reader      // The file reader.
}

type rbt_sort struct {
    num int
    rbt data
}

func new_scrub() *scrub {
    return &scrub{
    	sr:   &scrub_req{files: map[string]*scrub_file{}},
    	cs:   map[string]*cache{},
    }
}
func (sc *scrub) run(w io.Writer) {
    for _, sf := range sc.sr.files {
        sc.cs[sf.name] = sf.import_file()
    }
    hdrs := strings.Join(sc.cs["rebates"].hdrs, ",")
    w.Write([]byte("stat,"+hdrs+"\n"))
    plcy := getPolicy(sc.sr.manu)
    wgrp := sync.WaitGroup{}
    wgrp.Add(3)
    chn1  := make(chan data, 100000)        // Connects rebate reader/workers to the rebate re-sorter.
    chn2  := make(chan data, 100000)        // Connects rebate re-sorter to the result writer.
    sc.plcy = plcy

    go sc.read_rebates(&wgrp, chn1)
    go sc.sort_rebates(&wgrp, chn1, chn2)
    go sc.save_rebates(&wgrp, chn2, w)      // The result writer thread. Reads the output from the rebate re-sorter.

    wgrp.Wait()
}

func (sc *scrub) read_rebates(wgrp *sync.WaitGroup, out chan data) {
    cgrp := &sync.WaitGroup{}
    thrs := 1
    chns := make([]chan data, thrs)
    // Create the rebate workers.
    for a := 0; a < len(chns); a++ {
        cgrp.Add(1)
        chns[a] = make(chan data, 20)
        go sc.read_rebates_worker(cgrp, chns[a], out, a)
    }
    // Now that all rebate worker threads are started, feed them rebates.
    if rbtc := sc.cs["rebates"]; rbtc != nil {      // *Should* always have this!
        for i := 0; i < len(rbtc.elems); i++ {
            indx := fmt.Sprintf("%d", i)
            rbts := rbtc.Find("indx", indx)
            if len(rbts) > 0 {
                rbt := rbts[0]
                if sc.sr.uniq != "" {
                    indx = rbt[sc.sr.uniq]
                }
                slot := hashIndex(indx, thrs)
                chns[slot] <-rbt
            }
        }
    }
    // All rebates have been distributed to the worker channels. Now close our end.
    for _, chn := range chns {
        close(chn)
    }
    cgrp.Wait()
    close(out)     // Tell the next step in the pipe (the sort thread) that no more data coming.
    wgrp.Done()
}
func (sc *scrub) read_rebates_worker(wgrp *sync.WaitGroup, in, out chan data, wid int) {
    for rbt := range in {
        sc.plcy.scrubRebate(sc, rbt)
        out <-rbt
    }
    wgrp.Done()
}
func (sc *scrub) sort_rebates(wgrp *sync.WaitGroup, in, out chan data) {
    // The sort thread. Reads the common scrub output channel written by input workers.
    // Re-sort the rebates and write to the result writer input channel.
    sortQ := []*rbt_sort{}
    last  := -1
    for rbt := range in {
        i64,_ := strconv.ParseInt(rbt["indx"], 10, 64)
        num   := int(i64)
        rsort := &rbt_sort{num: num, rbt: rbt}
        sortQ = append(sortQ, rsort)                    // Just put it onto our outbound queue.
        sort.SliceStable(sortQ, func(i, j int) bool {   // Sort them to make sure they're still in order.
            return sortQ[i].num < sortQ[j].num
        })

        // The next one to go needs to be the next rnum in sequence. Otherwise can't send.
        for (len(sortQ) > 0 && sortQ[0].num == last + 1) || len(sortQ) > 1000000 {
            rbt := sortQ[0].rbt
            sortQ = sortQ[1:] // Trims first entry off.
            out <-rbt
            last++
        }
        // We've sent out what we can. Go back to top and get another completed rebate/claim.
    }
    // No more writers. Take whatever we have buffered and send them.
    for _, rbtS := range sortQ {
        out <-rbtS.rbt
    }
    close(out)     // Tell the next step in the pipe (the result writer thread) no more data coming.
    wgrp.Done()
}
func (sc *scrub) save_rebates(wgrp *sync.WaitGroup, in chan data, w io.Writer) {
    for rbt := range in {
        str := sc.plcy.result(sc, rbt)
        w.Write([]byte(str+"\n"))
    }
    wgrp.Done()
}

func hashIndex(str string, modulo int) int {
	max := 18	// Can only fit ~20 hex digits into an int64, so let's not overflow the int64 when we parse.
	if len(str) < 18 {
		max = len(str)
	}
	str = str[0:max]
	if i, err := strconv.ParseInt(str, 10, 64); err == nil {
		return int(i%int64(modulo))
	} else if i, err := strconv.ParseInt(str, 16, 64); err == nil {
		return int(i%int64(modulo))
	} else {
		return 0
	}
}