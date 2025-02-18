package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type gclaim struct {
	sync.Mutex
	clm  *Claim
}
type gcache struct {
	sync.Mutex
	all []*gclaim
	rxn map[string][]*gclaim
	seq int64
}

type sclaim struct {
	sync.Mutex
	gclm *gclaim
	excl string
	rbts []*rebate
}
type scache struct {
	sync.Mutex
	all []*sclaim
	rxn map[string][]*sclaim
}

func new_gcache() *gcache {
	return &gcache{all: []*gclaim{}, rxn: map[string][]*gclaim{}, seq: -1}
}

func load_gclms(done *sync.WaitGroup, gc *gcache, dur *time.Duration) {
	defer done.Done()
	strt := time.Now()
	cnt  := 0
	whr  := fmt.Sprintf("manu = '%s' AND seq > %d", manu, gc.seq)
	if chn, err := db_select[Claim](atlas.pools["atlas"], "atlas", "atlas.claims", nil, whr, "", stop); err == nil {
		for clm := range chn {
			atlas.claims.add(clm)
			cnt++
		}
	}
	atlas.claims.sort()
	if dur != nil {
		*dur = time.Since(strt)
	}
	Log("atlas", "load_gclms", name, "claims loaded/added", time.Since(strt), map[string]any{"cnt": cnt, "manu": manu}, nil)
}

func (gc *gcache) add(clm *Claim) {
	gc.Lock()
	defer gc.Unlock()
	gclm := &gclaim{clm: clm}
	gclm.Lock()
	defer gclm.Unlock()
	gc.all = append(gc.all, gclm)
	gc.rxn[clm.Hfrx] = append(gc.rxn[clm.Hfrx], gclm)
	if clm.Hfrx != clm.Hrxn {
		gc.rxn[clm.Hrxn] = append(gc.rxn[clm.Hrxn], gclm)
	}
	if gclm.clm.Seq > gc.seq {
		gc.seq = gclm.clm.Seq
	}
}

func (gc *gcache) sort() {
	gc.Lock()
	defer gc.Unlock()
	sort.SliceStable(gc.all, func(i, j int) bool {
		return gc.all[i].clm.Doc < gc.all[j].clm.Doc
	})
	for _, list := range gc.rxn {
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].clm.Doc < list[j].clm.Doc
		})
	}
}

func (gc *gcache) new_scache() *scache {
	sc := &scache{all: make([]*sclaim, 0, len(gc.all)), rxn: make(map[string][]*sclaim)}
	gc.Lock()
	defer gc.Unlock()
	for _, gclm := range gc.all {
		sclm := &sclaim{gclm: gclm, rbts: []*rebate{}}
		sc.all = append(sc.all, sclm)
		sc.rxn[gclm.clm.Hfrx] = append(sc.rxn[gclm.clm.Hfrx], sclm)
		if gclm.clm.Hfrx != gclm.clm.Hrxn {
			sc.rxn[gclm.clm.Hrxn] = append(sc.rxn[gclm.clm.Hrxn], sclm)
		}
	}
	return sc
}

func (sc *scache) FindRXN(rxns ...string) []*sclaim {
	sc.Lock()
	defer sc.Unlock()
	list := []*sclaim{}
	set  := map[string]*sclaim{}
	for _, rxn := range rxns {
		for _, sclm := range sc.rxn[rxn] {	// Lock each gclaim?
			set[sclm.gclm.clm.Clid] = sclm	// Clid is unique - create a unique claim set
		}
	}
	for _, sclm := range set {
		list = append(list, sclm)
	}
	sort.SliceStable(list, func(i, j int) bool {	// Lock each gclaim?
		return list[i].gclm.clm.Doc < list[j].gclm.clm.Doc
	})
	return list		// Nothing locked when we return. Caller must do locking.
}