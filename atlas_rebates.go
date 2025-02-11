package main

import "fmt"

type rebate struct {
	rbt *Rebate
	sr  *ScrubRebate
	scs []*sclaim
	rcs []*ScrubRebateClaim
}

func new_rebate(rbt *Rebate) *rebate {
	return &rebate{rbt: rbt, sr: &ScrubRebate{Ivid: rbt.Ivid, Rbid: rbt.Rbid}, scs: []*sclaim{}, rcs: []*ScrubRebateClaim{}}
}

func (r rebate) String() string {
	return fmt.Sprintf("rbid=%d spid=%s rxn=%s dos=%d ndc=%s prid=%s hrxn=%s hdos=%s scid=%d ivid=%d indx=%d errc=%s errm=%s", r.rbt.Rbid, r.rbt.Spid, r.rbt.Rxn, r.rbt.Dos, r.rbt.Ndc, r.rbt.Prid, r.rbt.Hrxn, r.rbt.Hdos, r.sr.Scid, r.sr.Ivid, r.sr.Indx, r.sr.Errc, r.sr.Errm)
}