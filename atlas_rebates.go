package main

import "fmt"

type rebate struct {
	rbt  *Rebate
	stat string
	spmt string		// service provider match type (exact, xwalk, stack, chain)
	errc string
	errm string
	clms []*sclaim
	atts []*attempt
}
type attempt struct {
	sclm *sclaim
	excl string
}

func new_rebate(rbt *Rebate) *rebate {
	return &rebate{rbt: rbt, clms: []*sclaim{}, atts: []*attempt{}}
}

func (r *rebate) attempt(sclm *sclaim, excl string) {
	r.atts = append(r.atts, &attempt{sclm: sclm, excl: excl})
}
func (r *rebate) match(sclm *sclaim) {
	r.clms = append(r.clms, sclm)
}
func (r *rebate) new_scrub_rebate(sc *scrub) *ScrubRebate {
	return &ScrubRebate{
		Manu: sc.scrb.Manu,
		Scid: sc.scid,
		Ivid: r.rbt.Ivid,
		Rbid: r.rbt.Rbid,
		Stat: r.stat,
		Spmt: r.spmt,
		Errc: r.errc,
		Errm: r.errm,
	}
}
func (r *rebate) new_scrub_matches(sc *scrub) []*ScrubMatch {
	srcs := make([]*ScrubMatch, 0, len(r.clms))
	for _, sclm := range r.clms {
		src := &ScrubMatch{}
		src.Clid = sclm.gclm.clm.Clid
		src.Ivid = r.rbt.Ivid
		src.Manu = sc.scrb.Manu
		src.Rbid = r.rbt.Rbid
		src.Scid = sc.scid
		srcs = append(srcs, src)
	}
	return srcs
}
func (r *rebate) new_scrub_attempts(sc *scrub) []*ScrubAttempt {
	sas := make([]*ScrubAttempt, 0, len(r.clms))
	for _, att := range r.atts {
		sa := &ScrubAttempt{}
		sa.Clid = att.sclm.gclm.clm.Clid
		sa.Ivid = r.rbt.Ivid
		sa.Manu = sc.scrb.Manu
		sa.Rbid = r.rbt.Rbid
		sa.Scid = sc.scid
		sa.Excl = att.excl
		sas = append(sas, sa)
	}
	return sas
}
func (r *rebate) String() string {
	return fmt.Sprintf("rbid=%d spid=%s rxn=%s dos=%d ndc=%s prid=%s hrxn=%s hdos=%s ivid=%d errc=%s errm=%s", r.rbt.Rbid, r.rbt.Spid, r.rbt.Rxn, r.rbt.Dos, r.rbt.Ndc, r.rbt.Prid, r.rbt.Hrxn, r.rbt.Hdos, r.rbt.Ivid, r.errc, r.errm)
}