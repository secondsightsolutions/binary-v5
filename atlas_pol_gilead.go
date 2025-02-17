package main

import (
)

type pol_gilead_default struct {
	IPolicy
}
func (p *pol_gilead_default) options() *policy_opts {
	return &policy_opts{stacks: false, chains: false, load_rebates: true}
}
func (p *pol_gilead_default) rebate_order() []string {
	return nil
}
func (p *pol_gilead_default) prep_rebates(*scrub) {

}
func (p *pol_gilead_default) prep_claims(sc *scrub) {
	
}
func (p *pol_gilead_default) scrub_rebate(sc *scrub, rbt *rebate) {
	
}
