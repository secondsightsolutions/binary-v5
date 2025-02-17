package main


func seconds(val int64) int64 {
	if val > 1000000 {
		val = val / 1000000
	}
	return val
}
func days(val int64) int64 {
	secs := seconds(val)
	days := secs / (60*60*24)
	return days
}

func CheckBefore(t1, t2 int64, rndDays bool) string {
	v1 := t1
	v2 := t2
	if rndDays {
		v1 = days(v1)
		v2 = days(v2)
	}
	if v1 < v2 {
		return ""
	}
	return "old_rbt_new_clm"
}

func CheckOnAfter(t1, t2 int64, rndDays bool) string {
	v1 := t1
	v2 := t2
	if rndDays {
		v1 = days(v1)
		v2 = days(v2)
	}
	if v1 >= v2 {
		return ""
	}
	return "new_rbt_old_clm"
}

func CheckDayRange(t1, t2 int64, bef, aft int64) string {
	v1 := days(t1)
	v2 := days(t2)
	if bef > -1 && v1 < (v2 - bef) {
		return "below_range"
	}
	if aft > -1 && v1 > (v2 + aft) {
		return "above_range"
	}
	return ""
}
func DayDiff(t1, t2 int64) int64 {
	v1 := days(t1)
	v2 := days(t2)
	return v2 - v1
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
