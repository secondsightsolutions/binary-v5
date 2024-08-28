package main

func amgenPrepRebates(sc *scrub) {

}
func amgenPrepClaims(sc *scrub) {

}
func amgenScrubRebate(sc *scrub, rbt data) {
	clms := sc.cs["claims"].Find("rxn", rbt["rxn"])
	for _, clm := range clms {
		if clm["rxn"] != rbt["rxn"] {
			continue
		}
		if clm["dos"] != rbt["dos"] {
			continue
		}
		if clm["ndc"] != rbt["ndc"] {
			continue
		}
		rbt["stat"] = "matched"
		return
	}
	rbt["stat"] = "nomatch"
}
func amgenResult(sc *scrub, rbt data) string {
	return rbt["stat"] + "," + rbt["rbid"] + "," + rbt["dos"] + "," + rbt["rxn"] + "," + rbt["ndc"]
}
