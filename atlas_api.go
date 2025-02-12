package main


func CheckBefore(d1, d2 int64) string {
	if d1 > 1000000 {
		d1 = d1 / 1000000
	}
	if d2 > 1000000 {
		d2 = d2 / 1000000
	}
	if d1 < d2 {
		return ""
	}
	return "old_rbt_new_clm"
}

func CheckOnAfter(d1, d2 int64) string {
	if d1 > 1000000 {
		d1 = d1 / 1000000
	}
	if d2 > 1000000 {
		d2 = d2 / 1000000
	}
	if d1 >= d2 {
		return ""
	}
	return "new_rbt_old_clm"
}

func CheckDayRange(d1, d2 int64, bef, aft int64) string {
	if d1 > 1000000 {
		d1 = d1 / 1000000
	}
	if d2 > 1000000 {
		d2 = d2 / 1000000
	}
	if bef > -1 && d1 < (d2 - bef) {
		return "below_range"
	}
	if aft > -1 && d1 > (d2 + aft) {
		return "above_range"
	}
	return ""
}

func CheckClaimAvail(rbt *rebate, sclm *sclaim, allowMult bool, dayDiff int, baseDay int64) string {
	if len(sclm.rbts) > 0 {
		if !allowMult {
			return "clm_used"
		}
		// TODO: finish
	}
	return ""
}

func CheckSPI(sc *scrub, spiA, spiB string, chains, stacks bool) (bool, string) {
	return atlas.spis.match(spiA, spiB, chains, stacks)
}
