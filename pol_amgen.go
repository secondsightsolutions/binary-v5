package main

import (
	"bytes"
	"time"
)

func amgenPrepRebates(sc *scrub) {

}
func amgenPrepClaims(sc *scrub) {
	dt509 := time.Date(2023, time.May,    9, 0, 0, 0, 0, time.UTC)
	for _, clm := range sc.cs["claims"].rows {
		if TestTF(clm, Fields.Isgr) {
			continue
		}
		if TestTF(clm, Fields.Isih) {
			continue
		}
		clmTm := ParseStrToTime(clm[Fields.Doc])
		if CheckOnAfter(clmTm, dt509) {
			if !TestTF(clm, Fields.Cnfm) {
				clm[Fields.Excl] = "clm_not_cnfm"
				continue
			}
			if !TestTF(clm, Fields.Elig) && !TestTF(clm, Fields.Susp) {
				clm[Fields.Excl] = "not_elig_at_sub"
			}
		}
	}
}
func amgenScrubRebate(sc *scrub, rbt data) {
	chains := false
	stacks := false
	rbtTm  := ParseStrToTime(rbt[Fields.Dos])
	tm411  := time.Date(2023, time.April, 11, 0, 0, 0, 0, time.UTC)
	tm509  := time.Date(2023, time.May,    9, 0, 0, 0, 0, time.UTC)

	rxn    := rbt[Fields.Rxn]
	hrxn,_ := Hash(rbt[Fields.Rxn])
	
	clms1 := sc.cs["claims"].Find(Fields.Rxn,  rxn,  true)	// Copies of claims. Only do this if policy updates the claim.
	clms2 := sc.cs["claims"].Find(Fields.Frxn, rxn,  true)	// Currently this policy does not, so could use false here.
	clms3 := sc.cs["claims"].Find(Fields.Rxn,  hrxn, true)
	clms4 := sc.cs["claims"].Find(Fields.Frxn, hrxn, true)
	clms  := sort_lists(Fields.Dos, false, "indx", clms1, clms2, clms3, clms4)
	for _, clm := range clms {
		if clm[Fields.Excl] != "" {
			sc.atts.add(rbt, clm, "skip excluded", clm[Fields.Excl], "")
			continue
		}
		clmTm, err := TryParseStrToTime(clm[Fields.Doc])
		if err != nil {
			sc.atts.add(rbt, clm, "rebate/claim dates", "missing_field", "missing doc")
			continue
		}
		
		if CheckBefore(*clmTm, rbtTm) {
			sc.atts.add(rbt, clm, "claim before rebate", "dos_rbt_aft_clm", "")
			continue
		}
		if CheckBefore(rbtTm, tm411) {
			if CheckOnAfter(*clmTm, tm509) {
				sc.atts.add(rbt, clm, "rebate/claim dates", "old_rbt_new_clm", "rbt bef 4/11/23, clm on/aft 5/9/23")
				continue
			}
		} else {
			if CheckBefore(*clmTm, tm509) {
				sc.atts.add(rbt, clm, "rebate/claim dates", "new_rbt_old_clm", "rbt on/aft 4/11/23, clm bef 5/9/23")
				continue
			}
		}
		if CheckOnAfter(rbtTm, tm411) {
			if res := CheckRange(rbtTm, *clmTm, 30, 45); res != "" {
				sc.atts.add(rbt, clm, "rebate/claim dates", res, "rbt on/aft 4/11/23, clm not -30/+45") // "above_range", "below_range"
				continue
			}
			if !TestTF(clm, Fields.Cnfm) {
				sc.atts.add(rbt, clm, "claim cnfm", "clm_not_cnfm", "rbt on/aft 4/11/23, clm not conforming")
				continue
			}
			if !TestTF(clm, Fields.Elig) && !TestTF(clm, Fields.Susp) {
				sc.atts.add(rbt, clm, "claim elig", "not_elig_at_sub", "rbt on/aft 4/11/23, clm not elig and not susp")
				continue
			}
			sc.atts.add(rbt, clm, "rebate/claim dates", "", "rbt on/aft 4/11/23, clm within -30/+45, is conforming and eligible")
		}
		if rbt[Fields.Rxn] != clm[Fields.Rxn] && rbt[Fields.Rxn] != clm[Fields.Frxn] {
			sc.atts.add(rbt, clm, "test RX match", "no_match_rx", "no match")
			continue
		}
		sc.atts.add(rbt, clm, "test RX match", "", "matched")

		if yes, how := CheckSPI(sc, rbt[Fields.Spid], clm[Fields.Spid], chains, stacks);yes {
			rbt[Fields.Spmt] = how
			sc.update_spi_counts(how)
			sc.atts.add(rbt, clm, "test SPI match", "", "matched using " + how)
		} else {
			sc.atts.add(rbt, clm, "test SPI match", "no_match_spi", "no match")
			continue
		}
		rbt["stat"] = "matched"
		sc.metr.update_rbt_clm(rbt, clm)
		return
	}
	rbt["stat"] = "nomatch"
}
func amgenResult(sc *scrub, rbt data) string {
	var sb bytes.Buffer
	rbts := sc.cs["rebates"]
	sb.WriteString(rbt["stat"])
	sb.WriteString(",")
	for i, hdr := range rbts.hdrs {		// These are the original headers we received on import.
		shrn := rbts.GetShortName(hdr)
		sb.WriteString(rbt[shrn])
		if i < len(rbts.hdrs)-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}
