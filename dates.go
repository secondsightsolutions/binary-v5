package main

import (
	"fmt"
	"time"

)

type Dates struct {
    c  *cache
}
var dates *Dates


func init() {
    CryptInit(cert, cacr, "", pkey, salt, phrs)
    dates = NewDates(10)
}

func (d *Dates) FromHash(hashedVal string) (*time.Time, error) {
    if !Is64bitHash(hashedVal) {
        return nil, fmt.Errorf("not_hash")
    }
	rows := d.c.Find("hash", hashedVal)
	if len(rows) > 0 {
		if tm, err := TryParseStrToTime(rows[0]["clear"]); err == nil {
			return tm, nil
		} else {
			return nil, fmt.Errorf("not_date")
		}
	} else {
		return nil, fmt.Errorf("not_found")
	}
}

func NewDates(yrs int) *Dates {
    dates = &Dates{
    	c: &cache{
    		views: map[string]map[string][]map[string]string{},
    		rows:  []map[string]string{},
    	},
    }
	dates.c.Index("hash",  4)
	dates.c.Index("clear", -4)

    off := 365 * yrs

    // Today
    tmNow  := time.Now()
    tmOff  := tmNow.Add(time.Duration(0) * time.Hour * 24).UTC()
    tmFmt  := tmOff.UTC().Format("2006-01-02")
    hash,_ := Hash(tmFmt)
    elem   := data{"clear": tmFmt, "hash": hash}
	dates.c.Add(elem)

    // Offsets from today
    for a := 0; a < off; a++ {
        // Going forward day by day
        tmOff  = tmNow.Add(time.Duration(a+1) * time.Hour * 24).UTC()
        tmFmt  = tmOff.UTC().Format("2006-01-02")
        hash,_ = Hash(tmFmt)
        elem   = data{"clear": tmFmt, "hash": hash}
		dates.c.Add(elem)

        // Going backward day by day
        tmOff  = tmNow.Add(time.Duration(a+1) * time.Hour * 24 * -1).UTC()
        tmFmt  = tmOff.UTC().Format("2006-01-02")
        hash,_ = Hash(tmFmt)
        elem   = data{"clear": tmFmt, "hash": hash}
		dates.c.Add(elem)
    }
    return dates
}

// Takes two strings that are date-ish, meaning they are clear dates or are hashes of
// 2006-01-02 and returns the count of days where date1 is less than, equal, or greater
// than date2. If a date string is a hash, we reverse it using the date cache. 
func (d *Dates) Compare(dt1, dt2 string) (time.Duration, error) {
    var tm1, tm2 *time.Time
    var hsh1, hsh2 string
    var err error

    // Is the first argument a hash already? No? Then try to parse it as a date.
    if !Is64bitHash(dt1) {
        if tm1, err = TryParseStrToTime(dt1); err != nil {
            return 0, err
        }
        // It was a valid date. Let's get its hash.
        if hsh1, err = Hash(tm1.Format("2006-01-02")); err != nil {
            return 0, err
        }
    } else {
        hsh1 = dt1  // Already a hash. Let's hope it's a date!
    }
    // Same for the second argument. Check if it's a hash. If not, convert to date and hash.
    if !Is64bitHash(dt2) {
        if tm2, err = TryParseStrToTime(dt2); err != nil {
            return 0, err
        }
        if hsh2, err = Hash(tm2.Format("2006-01-02")); err != nil {
            return 0, err
        }
    } else {
        hsh2 = dt2
    }
    // Is nil only if the first date was a hash (in which case we need to look it up).
    if tm1 == nil {
		if clr, err := d.FromHash(hsh1); err == nil {
			tm1 = clr
		} else {
			return 0, fmt.Errorf("date1 hash not in date cache (%s)", err.Error())
		}
    }
    // Is nil only if the second date was a hash (in which case we need to look it up).
    if tm2 == nil {
		if clr, err := d.FromHash(hsh2); err == nil {
			tm2 = clr
		} else {
			return 0, fmt.Errorf("date2 hash not in date cache (%s)", err.Error())
		}
    }
    return tm2.Sub(*tm1), nil
}

// FindHash() takes a date string, converts it into its normal form of 2006-01-02, and then pulls out
// the hash for it.
// This function currently only being used from test code.
func (d *Dates) FindHash(dtStr string) (string, error) {
    if tm, err := TryParseStrToTime(dtStr); err != nil {
        return "", err
    } else {
        str  := tm.Format("2006-01-02")
		rows := d.c.Find("clear", str)
		if len(rows) > 0 {
			return rows[0]["hash"], nil
		}
        return "", fmt.Errorf("time not found in table")
    }
}

