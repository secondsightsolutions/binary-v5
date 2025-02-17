package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type scrub struct {
	scrb *Scrub
	cs   *cache_set
	rbts []*rebate
	clms *scache
	spis *SPIs
	scid int64
	ivid int64
	plcy IPolicy
	metr *Metrics
	lckM sync.Mutex
	done bool
}

func new_scrub(s *Scrub) *scrub {
	scrb := &scrub{
		scrb: s,
		cs:   atlas.ca.clone(),
		clms: atlas.claims.new_scache(),
		spis: atlas.spis,
		scid: s.Scid,
		ivid: s.Ivid,
		rbts: nil,
		plcy: GetPolicy(s.Manu, s.Plcy),
		metr: &Metrics{},
		lckM: sync.Mutex{},
	}
	if scrb.scrb.Test != "" {
		// testCache[Rebate](		scrb, &scrb.cs.rbts, "atlas.test_rebates")
		// testCache[Claim](		scrb, &scrb.cs.clms, "atlas.test_claims")
		testCache[Entity](		scrb, &scrb.cs.ents, "atlas.test_entities")
		testCache[Pharmacy](	scrb, &scrb.cs.phms, "atlas.test_pharmacies")
		testCache[NDC](			scrb, &scrb.cs.ndcs, "atlas.test_ndcs")
		testCache[SPI](			scrb, &scrb.cs.spis, "atlas.test_spis")
		testCache[Designation](	scrb, &scrb.cs.desg, "atlas.test_desigs")
		testCache[LDN](			scrb, &scrb.cs.ldns, "atlas.test_ldns")
		testCache[ESP1PharmNDC](scrb, &scrb.cs.esp1, "atlas.test_esp1")
		testCache[Eligibility](	scrb, &scrb.cs.ledg, "atlas.test_eligibilities")
		scrb.spis = new_spis()
		scrb.spis.load(scrb.cs.spis, nil)
	}
	return scrb
}

func testCache[T any](scrb *scrub, ca **cache, tbln string) {
	whr := fmt.Sprintf(" test = '%s' ", scrb.scrb.Test)
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
	wrks := 10
	wqsz := 20
	wgrp := sync.WaitGroup{}
	wgrp.Add(2)
	go sc.load_rebates(&wgrp)			// If this policy wants rebates loaded (most do), then load/prep them here.
	go sc.prep_claims(&wgrp)  			// Filters/prepares claims based on policy.
	wgrp.Wait()							// Wait until all rebates pulled/prepped and claims are prepped (based on policy).

	chn1 := make(chan *rebate, 1000)	// Rebate puller will feed us all the rebates by sending to this channel. (see next)
	chn2 := make(chan *rebate, 1000)

	wgrp.Add(2)
	go sc.pull_rebates(&wgrp, chn1)			// This thread will close the rbts channel once all rebates have been written to it.
	go sc.work_rebates(&wgrp, chn1, chn2, wrks, wqsz)
	
	for rbt := range chn2 {
		sc.update_metrics(rbt)
	}
	sc.done = true
}

func (sc *scrub) work_rebates(wgrp *sync.WaitGroup, in, out chan *rebate, wrks, qsiz int) {
	defer wgrp.Done()
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
			rns := []rune(str)
			tot := int32(0)
			for a := 0; a < len(rns); a++ {
				tot += rns[a]
			}
			return int(int64(tot) % int64(modulo))
		}
	}
	getSlot := func(rbt *rebate, fld string, wrks int) int {
		slot := 0
		rflt := &rflt{}
		if fld != "" { // Rebate field that we use to find the correct worker (eg., all with same rx go to same worker).
			val := rflt.getFieldValueAsString(rbt, fld)
			slot = hashString(val, wrks)
		} else {
			slot = hashIndex(rbt.rbt.Rbid, wrks)
		}
		return slot
	}
	worker := func(wgrp *sync.WaitGroup, in <-chan *rebate, out chan<- *rebate) {
		defer wgrp.Done()
		for {
			select {
			case <-stop:
				return

			case rbt, ok := <-in:
				if ok {
					sc.plcy.scrub_rebate(sc, rbt)
					out <-rbt
				} else {
					return
				}
			}
		}
	}
	chns := make([]chan *rebate, wrks)
	cgrp := &sync.WaitGroup{}
	for a := 0; a < len(chns); a++ {
		cgrp.Add(1)
		chns[a] = make(chan *rebate, qsiz)
		go worker(cgrp, chns[a], out)	// The worker must always call cgrp.Done() !!!!
	}
	for {
		select {
		case <-stop:
		case rbt, ok := <-in:
			if !ok {
				for _, chn := range chns {	// Tell the workers that we're done. 
					close(chn)				// Workers will return once they drain their channels.
				}
				goto done					// Leave here to go below and wait for all workers to finish.
			}
			slot := getSlot(rbt, "dos", wrks)
			chns[slot] <- rbt				// Worker gets rebate on short per-worker queue.
		}
	}
	done:
	cgrp.Wait()					// Wait here for workers to return.
}
func (sc *scrub) save_rebates(qsiz int) {
	Log("atlas", "save_rebates", fmt.Sprintf("scid=%d", sc.scid), "starting", 0, nil, nil)
	strt := time.Now()
	
	scrb_rbts := make(chan *ScrubRebate,    qsiz)
	scrb_mtcs := make(chan *ScrubMatch,		qsiz)
	scrb_atts := make(chan *ScrubAttempt, 	qsiz)
	scrb_clms := make(chan *ScrubClaim,     qsiz)

	save_rbts := time.Duration(0)
	save_mtcs := time.Duration(0)
	save_atts := time.Duration(0)
	save_clms := time.Duration(0)
	
	dgrp := sync.WaitGroup{}
	dgrp.Add(4)
	db_insert_run(&dgrp, atlas.pools["atlas"], "atlas", "atlas.scrub_rebates",	nil, scrb_rbts, qsiz, "", false, true, nil, nil, nil, &save_rbts, stop)
	db_insert_run(&dgrp, atlas.pools["atlas"], "atlas", "atlas.scrub_matches", 	nil, scrb_mtcs, qsiz, "", false, true, nil, nil, nil, &save_mtcs, stop)
	db_insert_run(&dgrp, atlas.pools["atlas"], "atlas", "atlas.scrub_attempts",	nil, scrb_atts, qsiz, "", false, true, nil, nil, nil, &save_atts, stop)
	db_insert_run(&dgrp, atlas.pools["atlas"], "atlas", "atlas.scrub_claims",	nil, scrb_clms, qsiz, "", false, true, nil, nil, nil, &save_clms, stop)

	for _, rbt := range sc.rbts {
		sr  := rbt.new_scrub_rebate(sc)
		sms := rbt.new_scrub_matches(sc)
		sas := rbt.new_scrub_attempts(sc)

		scrb_rbts <- sr				// Writes ScrubRebate to the database writer.
		for _, sm := range sms {
			scrb_mtcs <- sm			// Writes the ScrubMatchs to the database writer.
		}
		for _, sa := range sas {
			scrb_atts <- sa			// Writes the ScrubAttempts to the database writer.
		}
	}
	close(scrb_rbts)	// Flushes any buffered db writes.
	close(scrb_mtcs)	// Flushes any buffered db writes.
	close(scrb_atts)	// Flushes any buffered db writes.

	for _, sclm := range sc.clms.all {
		scrb := &ScrubClaim{Scid: sc.scid, Clid: sclm.gclm.clm.Clid, Excl: sclm.excl}
		scrb_clms <-scrb
	}
	close(scrb_clms)	// Flushes any buffered db writes.

	dgrp.Wait()

	sc.metr.SaveScrubRebates  = int64(save_rbts)
	sc.metr.SaveScrubMatches  = int64(save_mtcs)
	sc.metr.SaveScrubAttempts = int64(save_atts)
	sc.metr.SaveScrubClaims   = int64(save_clms)

	Log("atlas", "save_rebates", fmt.Sprintf("scid=%d", sc.scid), "completed", time.Since(strt), map[string]any{"rbts": sc.metr.SaveScrubRebates, "mtcs": sc.metr.SaveScrubMatches, "atts": sc.metr.SaveScrubAttempts, "clms": sc.metr.SaveScrubClaims}, nil)
}

func (sc *scrub) load_rebates(wgrp *sync.WaitGroup) {
	defer wgrp.Done()
	if !sc.plcy.options().load_rebates {	// The option to load rebates into memory must be stated. Otherwise we stream.
		return
	}
	pool := atlas.pools["atlas"]
	whr  := fmt.Sprintf("manu = '%s' AND ivid = %d", sc.scrb.Manu, sc.ivid)
	if cnt, err := db_count(context.Background(), pool, fmt.Sprintf("FROM atlas.rebates WHERE manu = '%s' AND ivid = %d", sc.scrb.Manu, sc.ivid)); err == nil {
		sc.rbts = make([]*rebate, 0, cnt)
	}
	if chn, err := db_select[Rebate](pool, "atlas", "atlas.rebates", nil, whr, "rbid", stop); err == nil {
		for rbt := range chn {
			sc.rbts = append(sc.rbts, new_rebate(rbt))	// All rebates into memory.
		}
	}
	sc.plcy.prep_rebates(sc)	// Filters/prepares rebates based on policy.
}

func (sc *scrub) prep_claims(wgrp *sync.WaitGroup) {
	defer wgrp.Done()
	sc.plcy.prep_claims(sc)	// Claims always in memory. Scrub has a claim cache that wraps the global claims cache.
}

func (sc *scrub) pull_rebates(wgrp *sync.WaitGroup, out chan<- *rebate) {
	defer wgrp.Done()
	// Only one thread here. Make sure we pull rebates in their proper order (and queue them to next stage in order).
	defer close(out)		// Starts the completion sequence downstream.
	if sc.rbts != nil {		// If we've loaded into memory (and prepped them!) then the rebates come from memory.
		for _, rbt := range sc.rbts {
			out <- rbt		// Sends rebate to next stage (work_rebates)
		}
	} else {				// Rebates not pre-loaded into memory, so stream them in from the database.
		whr  := fmt.Sprintf("manu = '%s' AND ivid = %d", sc.scrb.Manu, sc.ivid)
		sort := strings.Join(sc.plcy.rebate_order(), ",")
		if sort == "" {
			sort = "rbid"
		}
		if chn, err := db_select[Rebate](atlas.pools["atlas"], "atlas", "atlas.rebates", nil, whr, sort, stop); err == nil {
			for rbt := range chn {
				out <- new_rebate(rbt)	// Sends rebate to next stage (work_rebates)
			}
		}
	}
}
