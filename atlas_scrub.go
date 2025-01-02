package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
)

type claim struct {
	clm  *Claim
	excl string
}
type scrub struct {
	sr   *Scrub
	cs   *cache_set
	spis *SPIs
	clms *cache
	scid int64
	slot string
	outf string
	// sr   *scrub_req
	hdrs []string
	rbtC chan *Rebate
	plcy *Policy
	atts *Attempts
	metr *Metrics
	lckM sync.Mutex
}

// type scrub_req struct {
// 	auth  string
// 	manu  string
// 	sort  string                 // The sort order for rebates (defaults to "indx", the generated ordinal position in the rebate file).
// 	slot  string                 // How we divide across rebate procs. All rebates with same value for "slot" must go to same rebate proc.
// 	outf  string                 // The output file.
// 	files map[string]*scrub_file // The set of input files (rebates, claims, LU tables).
// }

// type scrub_file struct {
// 	name  string // The name to use for the cache, like "rebates", "claims", "ndcs", etc.
// 	path  string // Disk file path. Either test dir or temp dir created by http upload.
// 	args  string
// 	csep  string            // If set, the hdr/col separator in the input (defaults to ",")
// 	hdrs  []string          // Original values for CSV headers or table column names.
// 	keys  []sort_key        // keyn;length;order,keyn,keyn;length (in policy definition).
// 	hdrm  map[string]string // Maps custom input name to proper short name (in policy definition for defining input).
// 	hdri  map[int]string    // CSV column index => proper_hdr
// 	shrt  map[string]string // Maps short name back to original name (CSV header or table column) (dynamic based on input found).
// 	full  map[string]string // Maps original name (CSV or table column) to short name (dynamic based on input found).
// 	lines int
// 	rderr error
// }

func new_scrub(s *Scrub, stop chan any) *scrub {
	scrb := &scrub{
		cs:   nil,
		spis: atlas.spis,
		scid: s.Scid,
		sr:   s,
		hdrs: []string{},
		rbtC: make(chan *Rebate),
		plcy: nil,
		atts: &Attempts{list: []*attempt{}},
		metr: &Metrics{},
		lckM: sync.Mutex{},
	}
	scrb.cs = atlas.ca.clone()
	if scrb.sr.Test != "" {
		setCache[Rebate](		scrb, &scrb.cs.rbts, "atlas.test_rebates",      stop)
		setCache[Claim](		scrb, &scrb.cs.clms, "atlas.test_claims", 		stop)
		setCache[Entity](		scrb, &scrb.cs.ents, "atlas.test_entities", 	stop)
		setCache[Pharmacy](		scrb, &scrb.cs.phms, "atlas.test_pharmacies",	stop)
		setCache[NDC](			scrb, &scrb.cs.ndcs, "atlas.test_ndcs", 		stop)
		setCache[SPI](			scrb, &scrb.cs.spis, "atlas.test_spis", 		stop)
		setCache[Designation](	scrb, &scrb.cs.desg, "atlas.test_desigs", 		stop)
		setCache[LDN](			scrb, &scrb.cs.ldns, "atlas.test_ldns", 		stop)
		setCache[ESP1PharmNDC](	scrb, &scrb.cs.esp1, "atlas.test_esp1", 		stop)
		setCache[Eligibility](	scrb, &scrb.cs.ledg, "atlas.test_eligibilities",stop)
	}
	if scrb.cs.spis != atlas.ca.spis {
		scrb.spis = newSPIs()
		scrb.spis.load(scrb.cs.spis)
	}
	scrb.clms = new_cache[Claim]() // TODO: load me
	return scrb
}
func setCache[T any](scrb *scrub, ca **cache, tbln string, stop chan any) {
	whr := fmt.Sprintf(" test = '%s' ", scrb.sr.Test)
	c   := new_cache[T]()
	*ca = c
	dbm := new_dbmap[T]()
	dbm.table(atlas.pools["atlas"], tbln)
	if chn, err := db_select[T](atlas.pools["atlas"], "atlas", tbln, dbm, whr, "", stop); err == nil {
		for obj := range chn {
			c.Add(obj)
		}
	}
}

func (sc *scrub) run() {
	wgrp := sync.WaitGroup{}
	wgrp.Add(2)
	go sc.prep_rebates(&wgrp) // Filters/prepares rebates based on policy.
	go sc.prep_claims(&wgrp)  // Filters/prepares claims based on policy.
	wgrp.Wait()

	chn1 := make(chan *Rebate, 100000) // Connects rebate database reader to the workers.
	chn2 := make(chan *Rebate, 100000) // Connects rebate workers to the rebate (db) saver.
	out1, in1 := (chan<- *Rebate)(chn1), (<-chan *Rebate)(chn1)
	out2, in2 := (chan<- *Rebate)(chn2), (<-chan *Rebate)(chn2)
	wgrp.Add(3)
	go sc.pull_rebates(&wgrp, out1) // Pull rebates from table in order specified by policy, and feed to rebate workers.
	go sc.work_rebates(&wgrp, in1, out2, 2, 20)
	go sc.save_rebates(&wgrp, in2, 5, 100) // Reads finished rebates from workers and updates the rebates table.
	wgrp.Wait()

	sc.file_rebates() // Reads the rebates table and generates the result file.

}

func (sc *scrub) prep_rebates(wgrp *sync.WaitGroup) {
	defer wgrp.Done()
	sc.plcy.prepRebates(sc)
}

func (sc *scrub) prep_claims(wgrp *sync.WaitGroup) {
	defer wgrp.Done()
	sc.plcy.prepClaims(sc)
}

func (sc *scrub) pull_rebates(wgrp *sync.WaitGroup, out chan<- *Rebate) {
	defer wgrp.Done()
	defer close(out)
	whr := fmt.Sprintf("scid = %d", sc.scid)
	sort := sc.plcy.rebateOrder
	if chn, err := db_select[Rebate](atlas.pools["atlas"], "atlas", "atlas.rebates", nil, whr, sort, nil); err == nil {
		for rbt := range chn {
			out <- rbt
		}
	}
}
func (sc *scrub) work_rebates(wgrp *sync.WaitGroup, in1 <-chan *Rebate, out2 chan<- *Rebate, wrks, size int) {
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
	worker := func(cgrp *sync.WaitGroup, in <-chan *Rebate, out chan<- *Rebate) {
		for rbt := range in {
			sc.plcy.scrubRebate(sc, rbt)
			sc.update_rbt(rbt)
			out <- rbt
		}
		cgrp.Done()
	}
	defer wgrp.Done()
	defer close(out2)
	cgrp := &sync.WaitGroup{}
	chns := make([]chan *Rebate, wrks)
	// Create the rebate workers.
	for a := 0; a < len(chns); a++ {
		cgrp.Add(1)
		chns[a] = make(chan *Rebate, size)
		go worker(cgrp, chns[a], out2)
	}
	// Now that all rebate worker threads are started, feed them rebates.
	rflt := &rflt{}
	indx := int64(0)
	sc.slot = strings.ToLower(sc.slot)
	for rbt := range in1 {
		rbt.Indx = indx
		slot := 0
		if sc.slot != "" { // Rebate field that we use to find the correct worker (eg., all with same rx go to same worker).
			val := rflt.getFieldValueAsString(rbt, sc.slot)
			slot = hashString(val, wrks)
		} else {
			slot = hashIndex(indx, wrks)
		}
		chns[slot] <- rbt
	}
	// All rebates have been distributed to the worker channels. Now close our end.
	for _, chn := range chns {
		close(chn)
	}
	cgrp.Wait()
}
func (sc *scrub) save_rebates(wgrp *sync.WaitGroup, in2 <-chan *Rebate, wrks, size int) {
	defer wgrp.Done()
	cgrp := &sync.WaitGroup{}
	pool := atlas.pools["atlas"]
	opts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	whr  := map[string]string{"scid": fmt.Sprintf("%d", sc.scid)}
	dbm  := new_dbmap[Rebate]()
	dbm.table(pool, "atlas.rebates")
	for a := 0; a < wrks; a++ { // Create the workers
		cgrp.Add(1)
		go func() { 			// Each worker runs separately
			defer cgrp.Done()
			cnt := 0
			tx, _ := pool.BeginTx(context.Background(), opts) // Create the first transaction for the batch of updates.
			for rbt := range in2 {                            // Keep reading rebates from the common input queue.
				if rbt != nil { 							  // When we get nil, the sending side is closed - we're done (almost).
					cnt++
					whr["rbid"] = fmt.Sprintf("%d", rbt.Rbid)
					db_update(context.Background(), rbt, tx, nil, "atlas.rebates", dbm, whr)
					if cnt%size == 0 { 						  // If we reached our size number of updates, commit the transaction and create another.
						tx.Commit(context.Background())
						tx, _ = pool.BeginTx(context.Background(), opts)
					}
				} else {	// No more rebates. But very likely we have uncommitted updates.
					break
				}
			}
			if cnt%size != 0 {
				tx.Commit(context.Background())
			} else {
				tx.Rollback(context.Background())
			}
		}()
	}
	cgrp.Wait()
}
func (sc *scrub) file_rebates() {
	if fd, err := os.Create(sc.outf); err == nil {
		w := bufio.NewWriter(fd)
		hdrs := strings.Join(sc.hdrs, ",")
		w.Write([]byte("stat," + hdrs + "\n"))

		whr := fmt.Sprintf("scid = %d", sc.scid)
		sort := "indx"
		if chn, err := db_select[Rebate](atlas.pools["atlas"], "atlas", "atlas.rebates", nil, whr, sort, nil); err == nil {
			for rbt := range chn {
				str := sc.plcy.result(sc, rbt)
				w.Write([]byte(str + "\n"))
			}
		}
		w.Flush()
		fd.Close()
	} else {
		Log("atlas", "file_rebates", "", "output file", 0, nil, err)
		//log("atlas", "file_rebates", "output file", 0, err)
	}
}
