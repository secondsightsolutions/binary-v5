package main

import (
	"time"
)

type Dates struct {
    hashToTime map[string]time.Time
    timeToHash map[time.Time]string
    hashToDays map[string]int64
    daysToHash map[int64]string
}

func new_dates(yrs int) *Dates {
    dates := &Dates{
        hashToTime: map[string]time.Time{},
        timeToHash: map[time.Time]string{},
        hashToDays: map[string]int64{},
        daysToHash: map[int64]string{},
    }
    add := func(now time.Time, offset, dir int) {
        tmOff := now.Add(time.Duration(offset) * time.Hour * 24 * time.Duration(dir))
        tmFmt  := tmOff.Format("2006-01-02")
        tmDays := tmOff.Unix()
        hash,_ := Hash(tmFmt)
        dates.daysToHash[tmDays] = hash
        dates.timeToHash[tmOff]  = hash
        dates.hashToDays[hash]   = tmDays
        dates.hashToTime[hash]   = tmOff
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

