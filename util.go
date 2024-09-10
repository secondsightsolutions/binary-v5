package main

import (
	"fmt"
	"strconv"
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

