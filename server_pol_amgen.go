package main

import (
	"fmt"
	"time"
)

func amgenPrepRebates(sc *Scrub) {

}
func amgenPrepClaims(sc *Scrub) {
	dt509 := time.Date(2023, time.May,    9, 0, 0, 0, 0, time.UTC)
	for _, row := range sc.ca.clms.rows {
		clm := row.elem.(*Claim)
		if clm.Isgr || clm.Isih {
			continue
		}
		clmTm := ParseI64ToTime(clm.Doc)
		if CheckOnAfter(clmTm, &dt509) {
			if !clm.Cnfm {
				clm.Excl = "clm_not_cnfm"
				continue
			}
			if !clm.Elig && !clm.Susp {
				clm.Excl = "not_elig_at_sub"
			}
		}
	}
}
func amgenScrubRebate(sc *Scrub, rbt *Rebate) {
	chains := false
	stacks := false
	rbtTm  := ParseStrToTime(rbt.Flds[Fields.Dos])
	tm411  := time.Date(2023, time.April, 11, 0, 0, 0, 0, time.UTC)
	tm509  := time.Date(2023, time.May,    9, 0, 0, 0, 0, time.UTC)

	rxn    := rbt.Flds[Fields.Rxn]
	hrxn,_ := Hash(rxn)
	
	clms1 := sc.ca.clms.Find(Fields.Rxn,  rxn)	// Copies of claims. Only do this if policy updates the claim.
	clms2 := sc.ca.clms.Find(Fields.Frxn, rxn)	// Currently this policy does not, so could use false here.
	clms3 := sc.ca.clms.Find(Fields.Rxn,  hrxn)
	clms4 := sc.ca.clms.Find(Fields.Frxn, hrxn)
	rows  := sort_lists(Fields.Doc, "GetDoc", 0, false, clms1, clms2, clms3, clms4)
	for _, row := range rows {
		clm := row.elem.(*Claim)
		if clm.Excl != "" {
			sc.atts.add(rbt, clm, "skip excluded", clm.Excl, "")
			continue
		}
		clmTm := ParseI64ToTime(clm.Doc)

		if CheckBefore(clmTm, rbtTm) {
			sc.atts.add(rbt, clm, "claim before rebate", "dos_rbt_aft_clm", "")
			continue
		}
		if CheckBefore(rbtTm, &tm411) {
			if CheckOnAfter(clmTm, &tm509) {
				sc.atts.add(rbt, clm, "rebate/claim dates", "old_rbt_new_clm", "rbt bef 4/11/23, clm on/aft 5/9/23")
				continue
			}
		} else {
			if CheckBefore(clmTm, &tm509) {
				sc.atts.add(rbt, clm, "rebate/claim dates", "new_rbt_old_clm", "rbt on/aft 4/11/23, clm bef 5/9/23")
				continue
			}
		}
		if CheckOnAfter(rbtTm, &tm411) {
			if res := CheckRange(rbtTm, clmTm, 30, 45); res != "" {
				sc.atts.add(rbt, clm, "rebate/claim dates", res, "rbt on/aft 4/11/23, clm not -30/+45") // "above_range", "below_range"
				continue
			}
			if !clm.Cnfm {
				sc.atts.add(rbt, clm, "claim cnfm", "clm_not_cnfm", "rbt on/aft 4/11/23, clm not conforming")
				continue
			}
			if !clm.Elig && !clm.Susp {
				sc.atts.add(rbt, clm, "claim elig", "not_elig_at_sub", "rbt on/aft 4/11/23, clm not elig and not susp")
				continue
			}
			sc.atts.add(rbt, clm, "rebate/claim dates", "", "rbt on/aft 4/11/23, clm within -30/+45, is conforming and eligible")
		}
		if rbt.Hrxn != clm.Hrxn && rbt.Hrxn != clm.Hfrx {
			sc.atts.add(rbt, clm, "test RX match", "no_match_rx", "no match")
			continue
		}
		sc.atts.add(rbt, clm, "test RX match", "", "matched")

		if yes, how := CheckSPI(sc, rbt.Fspd, clm.Spid, chains, stacks);yes {
			rbt.Spmt = how
			sc.update_spi_counts(how)
			sc.atts.add(rbt, clm, "test SPI match", "", "matched using " + how)
		} else {
			sc.atts.add(rbt, clm, "test SPI match", "no_match_spi", "no match")
			continue
		}
		rbt.Stat = "matched"
		sc.update_rbt_clm(rbt, clm)
		return
	}
	rbt.Stat = "nomatch"
}
func amgenResult(sc *Scrub, rbt *Rebate) string {
	return fmt.Sprintf("%s,%s,%s,%s", rbt.Stat, renderCols(sc.hdrs, rbt.Flds), rbt.Errc, rbt.Errm)
}
