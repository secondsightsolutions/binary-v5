package main

import (
	"time"
)

type pol_amgen_default struct {
	IPolicy
}
func (p *pol_amgen_default) options() *policy_opts {
	return &policy_opts{stacks: false, chains: false, load_rebates: true}
}
func (p *pol_amgen_default) rebate_order() []string {
	return nil
}
func (p *pol_amgen_default) prep_rebates(*scrub) {

}
func (p *pol_amgen_default) prep_claims(sc *scrub) {
	dt509  := time.Date(2023, time.May, 9, 0, 0, 0, 0, time.UTC)
	dt509s := dt509.Unix()
	for _, sclm := range sc.clms.all {	// No need to lock sclm yet.
		gclm := sclm.gclm
		gclm.Lock()
		clm := gclm.clm
		if IsGrantee(clm.I340) || len(clm.Ihph) > 0 {
			gclm.Unlock()
			continue
		}
		if res := CheckOnAfter(clm.Doc, dt509s, true); res != "" {
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
	tm411  := time.Date(2023, time.April, 11, 0, 0, 0, 0, time.UTC)
	tm509  := time.Date(2023, time.May,    9, 0, 0, 0, 0, time.UTC)
	sclms  := sc.clms.FindRXN(rbt.rbt.Rxn, rbt.rbt.Hrxn)
	tm411s := tm411.Unix()
	tm509s := tm509.Unix()
	
	for _, sclm := range sclms {
		clm := sclm.gclm.clm
		
		if res := CheckBefore(clm.Doc, rbt.rbt.Dos, true); res != "" {
			rbt.attempt(sclm, res)
			continue
		}
		if CheckBefore(rbt.rbt.Dos, tm411s, true) == "" {
			if res := CheckOnAfter(clm.Doc, tm509s, true); res != "" {
				rbt.attempt(sclm, res)
				continue
			}
		} else {
			if res := CheckBefore(clm.Doc, tm509s, true); res != "" {
				rbt.attempt(sclm, res)
				continue
			}
		}
		if CheckOnAfter(rbt.rbt.Dos, tm411s, true) == "" {
			if res := CheckDayRange(rbt.rbt.Dos, clm.Doc, 30, 45); res != "" {
				rbt.attempt(sclm, res)
				continue
			}
			if !clm.Cnfm {
				rbt.attempt(sclm, "clm_not_cnfm")
				continue
			}
			if !clm.Elig && !clm.Susp {
				rbt.attempt(sclm, "not_elig_at_sub")
				continue
			}
		}
		// if rbt.rbt.Hrxn != clm.Hrxn && rbt.rbt.Hrxn != clm.Hfrx {
		// 	continue	// no_match_rx
		// }

		if yes, which := CheckSPI(sc, rbt.rbt.Spid, rbt.rbt.Spid, p.options().chains, p.options().stacks);!yes {
			rbt.attempt(sclm, "no_match_spi")
			continue
		} else {
			rbt.spmt = which
			rbt.match(sclm)
		}
		rbt.stat = "matched"
		return
	}
	rbt.stat = "nomatch"
}
