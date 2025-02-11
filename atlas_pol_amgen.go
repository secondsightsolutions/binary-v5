package main

import (
	"time"
)

type pol_amgen_default struct {
	IPolicy
}
func (p *pol_amgen_default) prep_rebates(*scrub) {

}
func (p *pol_amgen_default) prep_claims(sc *scrub) {
	dt509 := time.Date(2023, time.May, 9, 0, 0, 0, 0, time.UTC)
	for _, sclm := range sc.clms.all {	// No need to lock sclm yet.
		gclm := sclm.gclm
		gclm.Lock()
		clm := gclm.clm
		if IsGrantee(clm.I340) || len(clm.Ihph) > 0 {
			gclm.Unlock()
			continue
		}
		clmTm := ParseI64ToTime(clm.Doc)
		if CheckOnAfter(clmTm, &dt509) {
			if !clm.Cnfm {
				sclm.excl = "clm_not_cnfm"
			} else if !clm.Elig && !clm.Susp {
				sclm.excl = "not_elig_at_sub"
			}
		}
		gclm.Unlock()
	}
}
func (p *pol_amgen_default) scrub_rebate(sc *scrub, rbt *rebate) {
	rbtTm  := ParseI64ToTime(rbt.rbt.Dos)
	tm411  := time.Date(2023, time.April, 11, 0, 0, 0, 0, time.UTC)
	tm509  := time.Date(2023, time.May,    9, 0, 0, 0, 0, time.UTC)
	sclms  := sc.clms.FindRXN(rbt.rbt.Rxn, rbt.rbt.Hrxn)
	
	for _, sclm := range sclms {
		sclm.Lock()
		gclm := sclm.gclm
		gclm.Lock()
		clm := gclm.clm
		
		clmTm := ParseI64ToTime(clm.Doc)
		cnfm  := clm.Cnfm
		elig  := clm.Elig
		susp  := clm.Susp
		spid  := clm.Spid
		gclm.Unlock()

		if CheckBefore(clmTm, rbtTm) {
			goto next	// dos_rbt_aft_clm
		}
		if CheckBefore(rbtTm, &tm411) {
			if CheckOnAfter(clmTm, &tm509) {
				goto next	// old_rbt_new_clm
			}
		} else {
			if CheckBefore(clmTm, &tm509) {
				goto next	// new_rbt_old_clm
			}
		}
		if CheckOnAfter(rbtTm, &tm411) {
			if res := CheckRange(rbtTm, clmTm, 30, 45); res != "" {
				goto next	// res
			}
			if !cnfm {
				goto next	// clm_not_cnfm
			}
			if !elig && !susp {
				goto next	// not_elig_at_sub
			}
		}
		// if rbt.rbt.Hrxn != clm.Hrxn && rbt.rbt.Hrxn != clm.Hfrx {
		// 	continue	// no_match_rx
		// }

		if yes, _ := CheckSPI(sc, rbt.rbt.Spid, spid, p.options().chains, p.options().stacks);!yes {
			goto next	// no_match_spi
		}
		rbt.sr.Stat = "matched"
		return

		next:
		sclm.Unlock()
	}
	rbt.sr.Stat = "nomatch"
}
