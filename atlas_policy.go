package main

import "fmt"

type IPolicy interface {
	options() *policy_opts
	rebate_order() []string
	prep_rebates(*scrub)
	prep_claims(*scrub)
	scrub_rebate(*scrub, *rebate)
}

type policy_opts struct {
	stacks bool
	chains bool
	load_rebates bool
}

var policies = map[string]map[string]IPolicy {
	"abbvie": {
		"default": nil,
	},
	"amgen": {
		"default": &pol_amgen_default{},
	},
	"astrazeneca": {
		"default": &pol_simple{},
	},
	"bayer": {
		"default": nil,
	},
	"gilead": {
		"default": &pol_gilead_default{},
	},
	"johnson_n_johnson": {
		"default": nil,
	},
	"eli_lilly": {
		"default": nil,
	},
	"novartis": {
		"default": nil,
	},
	"novo_nordisk": {
		"default": nil,
	},
	"organon": {
		"default": nil,
	},
	"pfizer": {
		"default": nil,
	},
	"sanofi": {
		"default": nil,
	},
	"takeda": {
		"default": nil,
	},
	"teva": {
		"default": nil,
	},
	"ucb": {
		"default": nil,
	},
}

func GetPolicy(manu, name string) IPolicy {
	if m,ok := policies[manu];ok {
		if p,ok := m[name];ok {
			return p
		}
	}
	return &pol_simple{}
}


type pol_simple struct {
	IPolicy
}
func (p *pol_simple) options() *policy_opts {
	return &policy_opts{chains: false, stacks: false, load_rebates: true}
}
func (p *pol_simple) rebate_order() []string {
	return []string{"rbid"}
}
func (p *pol_simple) prep_rebates(*scrub) {
}
func (p *pol_simple) prep_claims(sc *scrub) {
	
}

func (p *pol_simple) scrub_rebate(sc *scrub, rbt *rebate) {
	fmt.Printf("scrubbing rebate %v\n", rbt)
	// sclms := sc.clms.FindRXN(rbt.rbt.Rxn, rbt.rbt.Hrxn)
	// rbt.sr.Stat = Stats.nomatch
	// for _, sclm := range sclms {
	// 	sclm.Lock()
	// 	gclm := sclm.gclm
	// 	gclm.Lock()
	// 	clm  := gclm.clm

	// 	spid := clm.Spid
	// 	ndc  := clm.Ndc
	// 	dos  := clm.Dos
	// 	gclm.Unlock()

	// 	if spid != rbt.rbt.Spid {
	// 		goto next
	// 	}
	// 	if ndc != rbt.rbt.Ndc {
	// 		goto next
	// 	}
	// 	if dos != rbt.rbt.Dos {
	// 		goto next
	// 	}
	// 	rbt.scs = append(rbt.scs, sclm)
	// 	rbt.sr.Stat = Stats.matched
	// 	break

	// 	next:
	// 	sclm.Unlock()
	// }
}