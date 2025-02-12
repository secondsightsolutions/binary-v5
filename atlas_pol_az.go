package main

import "time"

type pol_astrazeneca_default struct {
	IPolicy
	dt801 int64
	dt916 int64
}

func new_pol_astrazeneca_default() *pol_astrazeneca_default {
	pol := &pol_astrazeneca_default{}
	dt916 := time.Date(2023, time.September, 16, 0, 0, 0, 0, time.UTC) // policy buffer
	dt801 := time.Date(2023, time.August,     1, 0, 0, 0, 0, time.UTC) // policy effective date
	pol.dt801 = dt801.Unix()
	pol.dt916 = dt916.Unix()
	return pol
}
func (p *pol_astrazeneca_default) options() *policy_opts {
	return &policy_opts{stacks: false, chains: false, load_rebates: true}
}
func (p *pol_astrazeneca_default) rebate_order() []string {
	return nil
}
func (p *pol_astrazeneca_default) prep_rebates(*scrub) {

}
func (p *pol_astrazeneca_default) prep_claims(sc *scrub) {
	for _, sclm := range sc.clms.all {
		clm := sclm.gclm.clm
		if clm.Ihph != "" {
			continue
		}
		if clm.Doc < p.dt916 {
			continue
		}
		if !clm.Elig {
			sclm.excl = "not_elig_at_sub"
		}
	}
}
func (p *pol_astrazeneca_default) scrub_rebate(sc *scrub, rbt *rebate) {
	sclms := sc.clms.FindRXN(rbt.rbt.Rxn, rbt.rbt.Hrxn)
	for _, sclm := range sclms {
		if res := CheckDayRange(rbt.rbt.Dos, sclm.gclm.clm.Doc, 30, -1); res != "" {
			rbt.attempt(sclm, res)
			continue
		}
		if res := CheckClaimAvail(rbt, sclm, true, 0, 0); res != "" {
			rbt.attempt(sclm, res)
			continue
		}
		if sclm.gclm.clm.Ihph != "" {
			rbt.match(sclm)
			return
		}
		if res := CheckBefore(rbt.rbt.Dos, p.dt801); res == "" {
			if res := CheckBefore(sclm.gclm.clm.Doc, p.dt916); res == "" {
				rbt.match(sclm)
				return
			} else {
				rbt.attempt(sclm, res)
				continue
			}
		} else {
			if res := CheckBefore(sclm.gclm.clm.Doc, p.dt801); res != "" {
				rbt.attempt(sclm, res)
				continue
			} else {
				if !sclm.gclm.clm.Elig {
					rbt.attempt(sclm, "not_elig_at_sub")
					continue
				} else {
					rbt.match(sclm)
					return
				}
			}
		}
	}
}