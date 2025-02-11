package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
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
	atts *Attempts
	metr *Metrics
	lckM sync.Mutex
	done chan any		// Tells the caller that the scrub is finished.
}

func new_scrub(s *Scrub, stop chan any) *scrub {
	scrb := &scrub{
		scrb: s,
		cs:   atlas.ca.clone(),
		clms: atlas.claims.new_scache(),
		spis: atlas.spis,
		scid: s.Scid,
		ivid: s.Ivid,
		rbts: nil,
		plcy: GetPolicy(s.Manu, s.Plcy),
		atts: &Attempts{list: []*attempt{}},
		metr: &Metrics{},
		lckM: sync.Mutex{},
		done: make(chan any, 1),
	}
	if scrb.scrb.Test != "" {
		// testCache[Rebate](		scrb, &scrb.cs.rbts, "atlas.test_rebates",      stop)
		// testCache[Claim](		scrb, &scrb.cs.clms, "atlas.test_claims", 		stop)
		testCache[Entity](		scrb, &scrb.cs.ents, "atlas.test_entities", 	stop)
		testCache[Pharmacy](	scrb, &scrb.cs.phms, "atlas.test_pharmacies",	stop)
		testCache[NDC](			scrb, &scrb.cs.ndcs, "atlas.test_ndcs", 		stop)
		testCache[SPI](			scrb, &scrb.cs.spis, "atlas.test_spis", 		stop)
		testCache[Designation](	scrb, &scrb.cs.desg, "atlas.test_desigs", 		stop)
		testCache[LDN](			scrb, &scrb.cs.ldns, "atlas.test_ldns", 		stop)
		testCache[ESP1PharmNDC](scrb, &scrb.cs.esp1, "atlas.test_esp1", 		stop)
		testCache[Eligibility](	scrb, &scrb.cs.ledg, "atlas.test_eligibilities",stop)
		scrb.spis = new_spis()
		scrb.spis.load(scrb.cs.spis)
	}
	return scrb
}

func testCache[T any](scrb *scrub, ca **cache, tbln string, stop chan any) {
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
	closeAndWait := func(wgrp *sync.WaitGroup, chns []chan *rebate, done chan *rebate) {
		for _, chn := range chns {
			close(chn)
		}
		wgrp.Wait()
		close(done)
	}

	wrks := 1
	qsiz := 20
	wgrp := sync.WaitGroup{}
	wgrp.Add(2)
	go sc.load_rebates(&wgrp)			// If this policy wants rebates loaded (most do), then load/prep them here.
	go sc.prep_claims(&wgrp)  			// Filters/prepares claims based on policy.
	wgrp.Wait()							// Wait until all rebates pulled/prepped and claims are prepped (based on policy).

	rbts := make(chan *rebate, 5000)	// Rebate puller will feed us all the rebates by sending to this channel. (see next)
	go sc.pull_rebates(rbts)			// This thread will close the rbts channel once all rebates have been written to it.

	scrb_rbts := make(chan *ScrubRebate,      500)
	scrb_rbcl := make(chan *ScrubRebateClaim, 500)
	scrb_clms := make(chan *ScrubClaim,       500)

	dgrp := sync.WaitGroup{}
	dgrp.Add(3)
	db_insert_run(&dgrp, atlas.pools["atlas"], "atlas", "atlas.scrub_rebates",        nil, scrb_rbts, 5000, "", false, true, nil, nil, nil)
	db_insert_run(&dgrp, atlas.pools["atlas"], "atlas", "atlas.scrub_rebates_claims", nil, scrb_rbcl, 5000, "", false, true, nil, nil, nil)
	db_insert_run(&dgrp, atlas.pools["atlas"], "atlas", "atlas.scrub_claims",         nil, scrb_clms, 5000, "", false, true, nil, nil, nil)

	done := make(chan *rebate, qsiz*wrks)
	chns := make([]chan *rebate, wrks)
	for a := 0; a < len(chns); a++ {
		wgrp.Add(1)
		chns[a] = make(chan *rebate, qsiz)
		go worker(&wgrp, chns[a], done)	// The worker must always call cgrp.Done() !!!!
	}
	for {
		select {
		case <-stop:
			// closeAndWait(&wgrp, chns, done)	// Closes workers, waits for them to finish, then closes done.
			// goto done

		case rbt, ok := <-rbts:				// Reads from main queue and distributes to workers.
			if !ok {
				closeAndWait(&wgrp, chns, done)
				rbts = nil
				continue
			}
			slot := getSlot(rbt, "dos", wrks)
			chns[slot] <- rbt				// Worker gets rebate on short per-worker queue.

		case rbt, ok := <-done:				// All workers return the rebates here once they've finished with them.
			if !ok {						// If no more rebates returned, and this channel was closed, we're done.
				goto done
			}
			sc.update_metrics(rbt)
			scrb_rbts <- rbt.sr				// Writes ScrubRebate to the database writer.
			for _, rc := range rbt.rcs {
				scrb_rbcl <- rc				// Writes the ScrubRebateClaims to the database writer.
			}
		}
	}
	done:
	close(scrb_rbts)	// Flushes any buffered db writes.
	close(scrb_rbcl)	// Flushes any buffered db writes.

	for _, sclm := range sc.clms.all {
		scrb := &ScrubClaim{Scid: sc.scid, Clid: sclm.gclm.clm.Clid, Excl: sclm.excl}
		scrb_clms <-scrb
	}
	close(scrb_clms)	// Flushes any buffered db writes.
	dgrp.Wait()
	sc.done <- nil		// Tells the caller (grpc server) that the scrub is done.
}

func (sc *scrub) load_rebates(wgrp *sync.WaitGroup) {
	defer wgrp.Done()
	if !sc.plcy.options().load_rebates {	// The option to load rebates into memory must be stated. Otherwise we stream.
		return
	}
	stop := make(chan any)
	pool := atlas.pools["atlas"]
	whr  := fmt.Sprintf("ivid = %d", sc.ivid)
	if cnt, err := db_count(context.Background(), pool, fmt.Sprintf("FROM atlas.rebates WHERE ivid = %d", sc.ivid)); err == nil {
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

func (sc *scrub) pull_rebates(out chan<- *rebate) {
	// Only one thread here. Make sure we pull rebates in their proper order (and queue them to next stage in order).
	defer close(out)
	if sc.rbts != nil {		// If we've loaded into memory (and prepped them!) then the rebates come from memory.
		for _, rbt := range sc.rbts {
			out <- rbt		// Sends rebate to next stage (work_rebates)
		}
	} else {				// Rebates not pre-loaded into memory, so stream them in from the database.
		whr  := fmt.Sprintf("ivid = %d", sc.ivid)
		sort := strings.Join(sc.plcy.rebate_order(), ",")
		if chn, err := db_select[Rebate](atlas.pools["atlas"], "atlas", "atlas.rebates", nil, whr, sort, nil); err == nil {
			for rbt := range chn {
				out <- new_rebate(rbt)	// Sends rebate to next stage (work_rebates)
			}
		}
	}
}
