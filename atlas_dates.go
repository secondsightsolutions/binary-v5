package main

import (
	"fmt"
	"time"
)

type Dates struct {
    hashToClear map[string]time.Time
    clearToHash map[time.Time]string
}
var dates *Dates

func init() {
    CryptInit(cert, cacr, "", pkey, salt, phrs)
    dates = NewDates(10)
}

func NewDates(yrs int) *Dates {
    add := func(now time.Time, offset, dir int) {
        tmOff := now.Add(time.Duration(offset) * time.Hour * 24 * time.Duration(dir)).UTC()
        tmFmt  := tmOff.UTC().Format("2006-01-02")
        hash,_ := Hash(tmFmt)
        dates.clearToHash[tmOff] = hash
        dates.hashToClear[hash]  = tmOff
    }
    dates = &Dates{
    	hashToClear: map[string]time.Time{},
    	clearToHash: map[time.Time]string{},
    }
    off := 365 * yrs
    now := time.Now()

    add(now, 0, 1)          // Today

    // Offsets from today
    for a := 0; a < off; a++ {
        add(now, a+1, 1)    // Going forward day by day
        add(now, a+1, -1)   // Going backward day by day
    }
    return dates
}

func (d *Dates) ToTime(obj any) (*time.Time, error) {
    switch val := obj.(type) {
    case string:
        if Is64bitHash(val) {
            if tm, ok := d.hashToClear[val]; ok {
                return &tm, nil
            } else {
                return nil, fmt.Errorf("hash not found")
            }
        } else if tm := ParseStrToTime(val); tm != nil {
            return tm, nil
        } else {
            return nil, fmt.Errorf("cannot convert string to time")
        }

    case int64:
        if tm := ParseI64ToTime(val); tm != nil {
            return tm, nil
        } else {
            return nil, fmt.Errorf("cannot convert int64 to time")
        }
    case *time.Time:
        return val, nil
    case time.Time:
        return &val, nil
    default:
        return nil, fmt.Errorf("unrecognized type %T", val)
    }
}

// Takes two strings that are date-ish, meaning they are clear dates or are hashes of
// 2006-01-02 and returns the duration where date1 is less than, equal, or greater
// than date2. If a date string is a hash, we reverse it using the date cache. 
func (d *Dates) Compare(dt1A, dt2A any) (time.Duration, error) {
    if tm1, err := d.ToTime(dt1A); err != nil {
        if tm2, err := d.ToTime(dt2A); tm2 != nil {
            return tm2.Sub(*tm1), nil
        } else {
            return 0, err
        }
    } else {
        return 0, err
    }
}
