package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var prefDateFmts = map[string]any{}

var allDateFmts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05.000",
	"2006-01-02 15:04:05.000000",
	"2006-01-02 15:04:05.999",
	"2006-01-02 15:04:05.999999",
	"2006/01/02 15:04:05",
	"2006/01/02 15:04:05.000",
	"2006/01/02 15:04:05.000000",
	"2006/01/02 15:04:05.999",
	"2006/01/02 15:04:05.999999",

	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05.000",
	"2006-01-02T15:04:05.000000",
	"2006-01-02T15:04:05.999",
	"2006-01-02T15:04:05.999999",
	"2006/01/02T15:04:05",
	"2006/01/02T15:04:05.000",
	"2006/01/02T15:04:05.000000",
	"2006/01/02T15:04:05.999",
	"2006/01/02T15:04:05.999999",

	"2006-01-02 15:04:05 MST",
	"2006-01-02 15:04:05.000 MST",
	"2006-01-02 15:04:05.000000 MST",
	"2006-01-02 15:04:05.999 MST",
	"2006-01-02 15:04:05.999999 MST",
	"2006/01/02 15:04:05 MST",
	"2006/01/02 15:04:05.000 MST",
	"2006/01/02 15:04:05.000000 MST",
	"2006/01/02 15:04:05.999 MST",
	"2006/01/02 15:04:05.999999 MST",

	"2006-01-02T15:04:05 MST",
	"2006-01-02T15:04:05.000 MST",
	"2006-01-02T15:04:05.000000 MST",
	"2006-01-02T15:04:05.999 MST",
	"2006-01-02T15:04:05.999999 MST",
	"2006/01/02T15:04:05 MST",
	"2006/01/02T15:04:05.000 MST",
	"2006/01/02T15:04:05.000000 MST",
	"2006/01/02T15:04:05.999 MST",
	"2006/01/02T15:04:05.999999 MST",

	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05.000 -0700",
	"2006-01-02 15:04:05.000000 -0700",
	"2006-01-02 15:04:05.999 -0700",
	"2006-01-02 15:04:05.999999 -0700",
	"2006/01/02 15:04:05 -0700",
	"2006/01/02 15:04:05.000 -0700",
	"2006/01/02 15:04:05.000000 -0700",
	"2006/01/02 15:04:05.999 -0700",
	"2006/01/02 15:04:05.999999 -0700",

	"2006-01-02T15:04:05 -0700",
	"2006-01-02T15:04:05.000 -0700",
	"2006-01-02T15:04:05.000000 -0700",
	"2006-01-02T15:04:05.999 -0700",
	"2006-01-02T15:04:05.999999 -0700",
	"2006/01/02T15:04:05 -0700",
	"2006/01/02T15:04:05.000 -0700",
	"2006/01/02T15:04:05.000000 -0700",
	"2006/01/02T15:04:05.999 -0700",
	"2006/01/02T15:04:05.999999 -0700",

	"20060102",
	
	"2006-01-02",
	"2006-01-02T15:04",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04",
	"2006-01-02 15:04:05",
	"2006/01/02",
	"2006/01/02 15:04",
	"2006/01/02 15:04:05",

	"1/2/2006",

	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Stamp,
	time.StampMicro,
	time.StampMilli,
	time.StampNano,
	"1/2/06",
}

func TryParseStrToUnix(val, useFmt string) (int64, string, error) {
	if useFmt != "" {
		if tm, err := time.Parse(useFmt, val); err == nil {
			prefDateFmts[useFmt] = nil
			return tm.UTC().Unix(), useFmt, nil
		}
	}
	for _fmt := range prefDateFmts {
		if tm, err := time.Parse(_fmt, val); err == nil {
			return tm.UTC().Unix(), _fmt, nil
		}
	}
	for _, _fmt := range allDateFmts {
		if tm, err := time.Parse(_fmt, val); err == nil {
			prefDateFmts[_fmt] = nil
			return tm.UTC().Unix(), _fmt, nil
		}
	}
	return 0, "", fmt.Errorf("not a date")
}

func ParseStrToTime(val string) time.Time {
	if i64, err := strconv.ParseInt(val, 10, 64); err == nil {
		if i64 > int64(time.Second) {
			return time.Unix(i64 / int64(time.Second), i64 % int64(time.Second))
		} else if i64 > int64(time.Millisecond) {
			return time.Unix(i64 / int64(time.Millisecond), i64 % int64(time.Millisecond))
		} else if i64 > int64(time.Microsecond) {
			return time.Unix(i64 / int64(time.Microsecond), i64 % int64(time.Microsecond))
		} else if i64 > int64(time.Nanosecond) {
			return time.Unix(i64 / int64(time.Nanosecond), i64 % int64(time.Nanosecond))
		}
	}
	for _fmt := range prefDateFmts {
		if tm, err := time.Parse(_fmt, val); err == nil {
			return tm
		}
	}
	for _, _fmt := range allDateFmts {
		if tm, err := time.Parse(_fmt, val); err == nil {
			prefDateFmts[_fmt] = nil
			return tm
		}
	}
	return time.Time{}
}
func TryParseStrToTime(val string) (*time.Time, error) {
	zt := time.Time{}
	tm := ParseStrToTime(val)
	if tm != zt {
		return &tm, nil
	} else {
		return nil, fmt.Errorf("not_date")
	}
}
func StrDecToInt(val string) int {
	if d, err := strconv.ParseInt(val, 10, 64); err == nil {
		return int(d)
	}
	return 0
}

func TestTF(d data, attr string) bool {
	if _tf, err := strconv.ParseBool(d[attr]); err == nil {
		return _tf
	}
	return false
}

func CheckBefore(t1, t2 time.Time) bool {
	return t1.Before(t2)
}
func CheckOnAfter(t1, t2 time.Time) bool {
	return t1.Equal(t2) || t1.After(t2)
}
func CheckRange(t1, t2 time.Time, bef, aft int) string {
	days := t2.Sub(t1) / (time.Hour * 24)
	if days < 0 {
		days *= -1
		if days > time.Duration(bef) {
			return "below_range"
		}
	} else {
		if days > time.Duration(aft) {
			return "above_range"
		}
	}
	return ""
}
func CheckSPI(sc *scrub, spiA, spiB string, chains, stacks bool) (bool, string) {
	return sc.spis.match(spiA, spiB, chains, stacks)
}

func Is64bitHash(val string) bool {
	return len(val) == 64
}

var ScreenLevel = struct {
	None int
	Text int
	Bars int
	All  int
}{ 0, 1, 2, 3}

func screen(start time.Time, text string, current, max int, mfr, prc int, nl bool) {
	// Rows that are "hidden" actually just have the N of M hidden. Still display the row, along with the continually updating times (all platforms) and memory usages (on linux).
	_brg := strings.EqualFold("brg", X509ou())
	_mnu := strings.EqualFold(Type, "manu")
	_lvl := 0

	if _brg {
		_lvl = ScreenLevel.All
	} else if _mnu {
		_lvl = mfr
	} else {
		_lvl = prc
	}
	if _lvl == ScreenLevel.None {
		return
	}
	
	MemUse.Lock()
	valid := MemUse.valid
	ready := MemUse.ready
	inuse := MemUse.inuse
	avail := MemUse.avail
	total := MemUse.total
	used  := 0
	MemUse.Unlock()
	if total > 0 {
		used = 100 - (avail*100)/total
	}

	dur := time.Since(start)
	mil := dur.Milliseconds() % 1000
	sec := dur.Milliseconds() / 1000 % 60
	min := dur.Milliseconds() / 1000 / 60
	den := dur.Milliseconds()
	if den < 1000 {
		den = 1000
	}
	per := (int64(current) * 1000) / den

	fmt.Printf("\r%100s", " ")

	timS := fmt.Sprintf("%s (%02dm.%02ds.%03dms)", start.Format("2006-01-02 15:04:05"), min, sec, mil)
	memS := fmt.Sprintf(" (%5dM of %dM: %2d%% used)", inuse, total, used)
	perS := fmt.Sprintf(" (%6d/sec) ", per)
	barS := ""
	txtS := fmt.Sprintf(" %-30s", text)
	maxS := fmt.Sprintf(" %d of %d", current, max)
	prtS := "\r" + timS
	if !ready || !valid {
		memS = ""
	}
	if valid {
		prtS += memS
	}
	if _lvl == ScreenLevel.All {
		prtS += perS
	}
	if _lvl >= ScreenLevel.Bars {
		prtS += barS
	}
	if _lvl >= ScreenLevel.Text {
		prtS += txtS
	}
	if _lvl == ScreenLevel.All {
		prtS += maxS
	}
	fmt.Print(prtS)
	if nl {
		fmt.Println()
	}
}
