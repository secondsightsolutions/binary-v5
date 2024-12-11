package main

import (
	"time"
)

func amgenPrepRebates(sc *scrub) {

}
func amgenPrepClaims(sc *scrub) {
	dt509 := time.Date(2023, time.May, 9, 0, 0, 0, 0, time.UTC)
	for _, row := range sc.clms.rows {
		clm := row.elem.(*claim)
		Clm := clm.clm
		if IsGrantee(Clm.I340) || len(Clm.Ihph) > 0 {
			continue
		}
		clmTm := ParseI64ToTime(Clm.Doc)
		if CheckOnAfter(clmTm, &dt509) {
			if !Clm.Cnfm {
				clm.excl = "clm_not_cnfm"
				continue
			}
			if !Clm.Elig && !Clm.Susp {
				clm.excl = "not_elig_at_sub"
			}
		}
	}
}
func amgenScrubRebate(sc *scrub, rbt *Rebate) {
	chains := false
	stacks := false
	rbtTm  := ParseStrToTime(rbt.Dos)
	tm411  := time.Date(2023, time.April, 11, 0, 0, 0, 0, time.UTC)
	tm509  := time.Date(2023, time.May,    9, 0, 0, 0, 0, time.UTC)

	rxn    := rbt.Rxn
	hrxn,_ := Hash(rxn)
	
	clms1 := sc.clms.Find(Fields.Rxn,  rxn)	// Copies of claims. Only do this if policy updates the claim.
	clms2 := sc.clms.Find(Fields.Frxn, rxn)	// Currently this policy does not, so could use false here.
	clms3 := sc.clms.Find(Fields.Rxn,  hrxn)
	clms4 := sc.clms.Find(Fields.Frxn, hrxn)
	rows  := sort_lists("Doc", false, clms1, clms2, clms3, clms4)
	for _, row := range rows {
		clm := row.elem.(*Claim)
		// if clm.Excl != "" {
		// 	sc.atts.Add(rbt, clm, "skip excluded", clm.Excl, "")
		// 	continue
		// }
		clmTm := ParseI64ToTime(clm.Doc)

		if CheckBefore(clmTm, rbtTm) {
			sc.atts.Add(rbt, clm, "claim before rebate", "dos_rbt_aft_clm", "")
			continue
		}
		if CheckBefore(rbtTm, &tm411) {
			if CheckOnAfter(clmTm, &tm509) {
				sc.atts.Add(rbt, clm, "rebate/claim dates", "old_rbt_new_clm", "rbt bef 4/11/23, clm on/aft 5/9/23")
				continue
			}
		} else {
			if CheckBefore(clmTm, &tm509) {
				sc.atts.Add(rbt, clm, "rebate/claim dates", "new_rbt_old_clm", "rbt on/aft 4/11/23, clm bef 5/9/23")
				continue
			}
		}
		if CheckOnAfter(rbtTm, &tm411) {
			if res := CheckRange(rbtTm, clmTm, 30, 45); res != "" {
				sc.atts.Add(rbt, clm, "rebate/claim dates", res, "rbt on/aft 4/11/23, clm not -30/+45") // "above_range", "below_range"
				continue
			}
			if !clm.Cnfm {
				sc.atts.Add(rbt, clm, "claim cnfm", "clm_not_cnfm", "rbt on/aft 4/11/23, clm not conforming")
				continue
			}
			if !clm.Elig && !clm.Susp {
				sc.atts.Add(rbt, clm, "claim elig", "not_elig_at_sub", "rbt on/aft 4/11/23, clm not elig and not susp")
				continue
			}
			sc.atts.Add(rbt, clm, "rebate/claim dates", "", "rbt on/aft 4/11/23, clm within -30/+45, is conforming and eligible")
		}
		if rbt.Hrxn != clm.Hrxn && rbt.Hrxn != clm.Hfrx {
			sc.atts.Add(rbt, clm, "test RX match", "no_match_rx", "no match")
			continue
		}
		sc.atts.Add(rbt, clm, "test RX match", "", "matched")

		if yes, how := CheckSPI(sc, rbt.Fspd, clm.Spid, chains, stacks);yes {
			rbt.Spmt = how
			sc.update_spi_counts(how)
			sc.atts.Add(rbt, clm, "test SPI match", "", "matched using " + how)
		} else {
			sc.atts.Add(rbt, clm, "test SPI match", "no_match_spi", "no match")
			continue
		}
		rbt.Stat = "matched"
		sc.update_rbt_clm(rbt, clm)
		return
	}
	rbt.Stat = "nomatch"
}
func amgenResult(sc *scrub, rbt *Rebate) string {
	//return fmt.Sprintf("%s,%s,%s,%s", rbt.Stat, renderCols(sc.hdrs, rbt.Flds), rbt.Errc, rbt.Errm)
	return ""
}
