package main

import (
	"bufio"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type claim struct {
	clm  *Claim
	excl string
}
type scrub struct {
	ca   CA
	clms []*claim
	scid int64
	sr   *scrub_req
	hdrs []string
	rbtC chan *Rebate
	plcy *Policy
	atts *Attempts
	metr *Metrics
	lckM sync.Mutex
}

type scrub_req struct {
	auth  string
	manu  string
	sort  string                 // The sort order for rebates (defaults to "indx", the generated ordinal position in the rebate file).
	slot  string                 // How we divide across rebate procs. All rebates with same value for "slot" must go to same rebate proc.
	outf  string                 // The output file.
	files map[string]*scrub_file // The set of input files (rebates, claims, LU tables).
}

type scrub_file struct {
	name  string // The name to use for the cache, like "rebates", "claims", "ndcs", etc.
	path  string // Disk file path. Either test dir or temp dir created by http upload.
	args  string
	csep  string            // If set, the hdr/col separator in the input (defaults to ",")
	hdrs  []string          // Original values for CSV headers or table column names.
	keys  []sort_key        // keyn;length;order,keyn,keyn;length (in policy definition).
	hdrm  map[string]string // Maps custom input name to proper short name (in policy definition for defining input).
	hdri  map[int]string    // CSV column index => proper_hdr
	shrt  map[string]string // Maps short name back to original name (CSV header or table column) (dynamic based on input found).
	full  map[string]string // Maps original name (CSV or table column) to short name (dynamic based on input found).
	lines int
	rderr error
}

type rbt_sort struct {
	num int
	rbt *Rebate
}

// func new_scrub(scid int64, manu string) *Scrub {
// 	sc := &Scrub{
// 		scid: scid,
// 		sr:   &scrub_req{manu: manu, files: map[string]*scrub_file{}},
// 		atts: NewAttempts(),
// 		metr: &Metrics{},
// 		plcy: GetPolicy(manu),
// 	}
	// sc.sr.files["rebates"]  = &scrub_file{name: "rebates", csep: ",", hdrm: "rxnum=rxn;hrxnum=hrxn"}
	// sc.sr.files["claims"]   = &scrub_file{name: "claims",   pool: "citus",  tbln: "submission_rows"}
	// sc.sr.files["ndcs"]     = &scrub_file{name: "ndcs",     pool: "esp",    tbln: "ndcs"}
	// sc.sr.files["spis"]     = &scrub_file{name: "spis",     pool: "esp",    tbln: "ncpdp_providers"}
	// sc.sr.files["pharms"]   = &scrub_file{name: "pharms",   pool: "esp",    tbln: "contracted_pharmacies"}
	// sc.sr.files["ents"]     = &scrub_file{name: "ents",     pool: "esp",    tbln: "covered_entities"}
	// sc.sr.files["elig"]     = &scrub_file{name: "elig",     pool: "esp",    tbln: "eligibility_ledger"}
	// sc.sr.files["esp1"]     = &scrub_file{name: "esp1",     pool: "esp",    tbln: "esp1_providers"}
// 	return sc
// }

func (sc *scrub) run() {
	wgrp := sync.WaitGroup{}
	wgrp.Add(4)
	chn1 := make(chan *Rebate, 100000) // Connects rebate reader/workers to the rebate re-sorter.
	chn2 := make(chan *Rebate, 100000) // Connects rebate re-sorter to the rebate sender.
	chn3 := make(chan *Rebate, 100000) // Connects rebate sender to the result writer.

	go sc.read_rebates(&wgrp, chn1)       // Reads rebates from input source.
	go sc.sort_rebates(&wgrp, chn1, chn2) // Re-sorts the completed rebates coming from multiple workers.
	go sc.send_rebates(&wgrp, chn2, chn3) // Sends rebates up to server.
	go sc.save_rebates(&wgrp, chn3)       // Writes rebates to output file.

	wgrp.Wait()
}

func (sc *scrub) read_rebates(wgrp *sync.WaitGroup, out chan *Rebate) {
	hashIndex := func(indx int64, modulo int) int {
		return int(indx % int64(modulo))
	}
	hashString := func(str string, modulo int) int {
		max := 18 // Can only fit ~20 hex digits into an int64, so let's not overflow the int64 when we parse.
		if len(str) < 18 {
			max = len(str)
		}
		str = str[0:max]
		if i, err := strconv.ParseInt(str, 10, 64); err == nil {
			return int(i % int64(modulo))
		} else if i, err := strconv.ParseInt(str, 16, 64); err == nil {
			return int(i % int64(modulo))
		} else {
			return 0
		}
	}
	cgrp := &sync.WaitGroup{}
	thrs := 1
	chns := make([]chan *Rebate, thrs)
	// Create the rebate workers.
	for a := 0; a < len(chns); a++ {
		cgrp.Add(1)
		chns[a] = make(chan *Rebate, 20)
		go sc.read_rebates_worker(cgrp, chns[a], out)
	}
	// Now that all rebate worker threads are started, feed them rebates.
	indx := int64(0)
	sc.sr.slot = strings.ToLower(sc.sr.slot)
	for rbt := range sc.rbtC {
		rbt.Indx = indx
		slot := hashIndex(indx, thrs)
		if sc.sr.slot != "" {
			slot = hashString(rbt.GetRxn(), thrs)	// TODO: fix this! Field should be dynamic!
		}
		chns[slot] <- rbt
	}
	// All rebates have been distributed to the worker channels. Now close our end.
	for _, chn := range chns {
		close(chn)
	}
	cgrp.Wait()
	close(out) // Tell the next step in the pipe (the sort thread) that no more data coming.
	wgrp.Done()
}
func (sc *scrub) read_rebates_worker(wgrp *sync.WaitGroup, in, out chan *Rebate) {
	for rbt := range in {
		sc.plcy.scrubRebate(sc, rbt)
		sc.update_rbt(rbt)
		out <- rbt
	}
	wgrp.Done()
}
func (sc *scrub) sort_rebates(wgrp *sync.WaitGroup, in, out chan *Rebate) {
	// The sort thread. Reads the common scrub output channel written by input workers.
	// Re-sort the rebates and write to the result writer input channel.
	sortQ := []*rbt_sort{}
	last := -1
	for rbt := range in {
		i64 := rbt.Indx
		num := int(i64)
		rsort := &rbt_sort{num: num, rbt: rbt}
		sortQ = append(sortQ, rsort)                  // Just put it onto our outbound queue.
		sort.SliceStable(sortQ, func(i, j int) bool { // Sort them to make sure they're still in order.
			return sortQ[i].num < sortQ[j].num
		})

		// The next one to go needs to be the next rnum in sequence. Otherwise can't send.
		for (len(sortQ) > 0 && sortQ[0].num == last+1) || len(sortQ) > 1000000 {
			rbt := sortQ[0].rbt
			sortQ = sortQ[1:] // Trims first entry off.
			out <- rbt
			last++
		}
		// We've sent out what we can. Go back to top and get another completed rebate/claim.
	}
	// No more writers. Take whatever we have buffered and send them.
	for _, rbtS := range sortQ {
		out <- rbtS.rbt
	}
	close(out) // Tell the next step in the pipe (the result writer thread) no more data coming.
	wgrp.Done()
}
func (sc *scrub) send_rebates(wgrp *sync.WaitGroup, in, out chan *Rebate) {
	// pool  := sc.sr.files["rebates"].pool
	// tbln  := sc.sr.files["rebates"].tbln
	// do,di := putData(pool, tbln, "insert")  // do - data out (send to grpc), di - data in (back from grpc)
	// for {
	//     select {
	//     case rbt := <-in:       // Get rebate from previous stage.
	//         if rbt != nil {     // Not nil, means channel still open. Previous stage still active and sending.
	//             do <-rbt        // Send to grpc.
	//         } else {
	//             close(do)       // Close this side - tells grpc nothing more is coming.
	//         }

	//     case rbt := <-di:       // Get rebate back from grpc (was sent up to server, so now we can print it).
	//         if rbt != nil {     // Not nil, means channel still open.
	//             out <-rbt
	//         } else {
	//             close(out)      // Tell next stage that we're done - nothing more coming.
	//             goto done       // Is nil, means grpc has no work left to do, and closed the channel.
	//         }
	//     }
	// }
	for {
		select {
		case rbt := <-in: // Get rebate from previous stage.
			if rbt != nil { // Not nil, means channel still open. Previous stage still active and sending.
				out <- rbt // Send to grpc.
			} else {
				close(out) // Tell next stage that we're done - nothing more coming.
				goto done
			}
		}
	}
done:
	wgrp.Done()
}
func (sc *scrub) save_rebates(wgrp *sync.WaitGroup, in chan *Rebate) {
	if fd, err := os.Create(sc.sr.outf); err == nil {
		w := bufio.NewWriter(fd)
		hdrs := strings.Join(sc.hdrs, ",")
		w.Write([]byte("stat," + hdrs + "\n"))
		for rbt := range in {
			str := sc.plcy.result(sc, rbt)
			w.Write([]byte(str + "\n"))
		}
		w.Flush()
		fd.Close()
	} else {
		exit(sc, 1, "cannot create output file (%s): %s", sc.sr.outf, err.Error())
	}
	wgrp.Done()
}
