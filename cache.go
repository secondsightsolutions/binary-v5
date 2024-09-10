package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type cache struct {
    sync.Mutex
    views map[string]view   // Each view is an indexed representation of the set of elements.
    rows  []data            // All elements in this table.
    hdrs  []string          // Original values for CSV headers or table column names.
    keys  []sort_key        // keyn;length;order,keyn,keyn;length (in policy definition).
    hdrm  map[string]string // Maps custom input name to proper short name (in policy definition for defining input).
    hdri  map[int]string    // CSV column index => proper_hdr
    shrt  map[string]string // Maps short name back to original name (CSV header or table column) (dynamic based on input found).
    full  map[string]string // Maps original name (CSV or table column) to short name (dynamic based on input found).
}

type sort_key struct {
    keyn string
    keyl int
    desc bool
}
type sort_elem struct {
    index   int
    strVal  string
    dtVal   int64
    intVal  int64
    fltVal  float64
    row     data
}

type data = map[string]string   // Indexed by keyname, or attr name - like "rxn", "ndc", etc.
type view = map[string][]data   // Indexed by keyvalue, like "123456789"

func new_cache(sf *scrub_file) *cache {
    ca := &cache{
    	views: map[string]map[string][]map[string]string{},
    	rows:  []map[string]string{},
    	hdrs:  []string{},
    	keys:  []sort_key{},
    	hdrm:  map[string]string{},
    	hdri:  map[int]string{},
    	shrt:  map[string]string{},
    	full:  map[string]string{},
    }
    if sf.keys != "" {
        keyl := strings.Split(sf.keys, ",")     // keyn;length;order,keyn,keyn;length
        for _, key := range keyl {
            toks := strings.Split(key, ";")     // keyn;length;order
            keyn := toks[0]
            keyl := int(3)                      // default key length is 3
            desc := false                       // default sort order is ascending
            if len(toks) > 1 {                  // could be just keyn, or keyn;length, or keyn;length;order
                if lng, err := strconv.ParseInt(toks[1], 10, 64); err == nil {
                    keyl = int(lng)
                }
                if len(toks) > 2 {
                    ordr := toks[2]
                    if strings.EqualFold(ordr, "desc") {
                        desc = true
                    }
                }
            }
            ca.keys = append(ca.keys, sort_key{keyn: keyn, keyl: keyl, desc: desc})
        }
    }
    if sf.hdrs != "" {
        toks := strings.Split(sf.hdrs, ",")
        for _, tok := range toks {
            csp := strings.Split(tok, "=")
            if len(csp) == 2 {
                cust := csp[0]
                shrt := csp[1]
                ca.hdrm[cust] = shrt
            }
        }
    }
    return ca
}

// Sorts the main list and the chunks based on this key.
func (c *cache) Sort(keyn string, desc bool) {
    // Since we want to be careful not to pull in too much data, we'll read in the elems just
    // long enough to grab their current index and the value of the keyn attribute.
    // Once we have the full list of sort_elem objects we'll sort them, then use their new
    // order to rewrite the full list (and then the sublists within the chunks).
    c.Lock()
    defer c.Unlock()
    
    c.rows = sort_list(keyn, desc, c.rows)

    // Finally, let's reorder our lists! This is easier, now that we have the new index values.
    for _, vw := range c.views {
        for keyv, list := range vw {
            vw[keyv] = sort_list(keyn, desc, list)
        }
    }
}

func (c *cache) Add(d data) {
    c.Lock()
    defer c.Unlock()
    d["indx"] = fmt.Sprintf("%d", len(c.rows))
    c.rows = append(c.rows, d)
    for keyn := range c.views {
        c.index_data(keyn, d)
    }
}
func (c *cache) Index(keyn string, keyl int) {
    c.Lock()
    defer c.Unlock()
    if _, ok := c.views[keyn]; ok {
        return
    }
    vw := view{}
    c.views[keyn] = vw
    for _, d := range c.rows {
        c.index_data(keyn, d)
    }
}
func (c *cache) index_data(keyn string, d data) {
    vw   := c.views[keyn]
    keyv := d[keyn]
    if vw[keyv] == nil {
        vw[keyv] = []data{}
    }
    vw[keyv] = append(vw[keyv], d)
}

func (c *cache) Find(keyn, val string, copy bool) []data {
    c.Lock()
    defer c.Unlock()

    if _, ok := c.views[keyn];!ok {   // Key might be "rxn"
        c.Unlock()
        c.Index(keyn, 3)
        c.Lock()
    }
    vw   := c.views[keyn]
    if vw == nil {
        vw = view{}
        c.views[keyn] = vw
        for _, d := range c.rows {
            c.index_data(keyn, d)
        }
    }
    rows := vw[val]
    if copy {
        if len(rows) > 0 {
            cps := make([]map[string]string, 0, len(rows))
            for _, row := range rows {
                cp := map[string]string{}
                for k,v := range row {
                    cp[k] = v
                }
                cps = append(cps, cp)
            }
            return cps
        }
    }
    return rows
}
func (c *cache) Update(upd data) {
    c.Lock()
    defer c.Unlock()
    if vw := c.views["indx"];vw != nil {    // Special view always created - a view indexed by the row insertion index.
        if list, ok := vw[upd["indx"]];ok { // Get the row(s) that have this same lookup value (upd["index"])
            if len(list) == 1 {             // A view usually has only one row per lookup value (hopefully always true for indx!)
                row := list[0]              // Should only be one!
                for k,v := range upd {      // Take every k/v in the updated row...
                    row[k] = v              // ...and write it into the row in the cache.
                }
                for k := range row {        // Now look at every key in the cache row...
                    if _,ok := upd[k];!ok { // If it's not in the update we received, then the k/v was removed.
                        delete(row, k)      // So remove it from the row in the cache.
                    }
                }
            }
        }
    }
}

func sort_lists(keyn string, desc bool, uniq string, lists ...[]data) []data {
    all := []data{}
    set := map[string]any{}
    for _, list := range lists {
        sub := sort_list(keyn, desc, list)
        all = append(all, sub...)
    }
    all = sort_list(keyn, desc, all)
    unq := []data{}
    for _, row := range all {
        if _, ok := set[row[uniq]];!ok {
            unq = append(unq, row)
            set[row[uniq]] = nil
        }
    }
    return unq
}
func sort_list(keyn string, desc bool, list []data) []data {
    elms := []*sort_elem{}  // This is the list of tiny elems - contains just enough to do a sort.
    // Since we can only load by chunk, grab the first view (any will do) and go through their
    // chunks, loading each chunk and then grabbing the current global index of that elem along
    // with the value we're sorting by.
    dtOK  := true           // Continue to convert as datetime
    fltOK := true           // Continue to convert as float
    intOK := true           // Continue to convert as integer
    dtFmt := ""
    for i, d := range list {
        se := &sort_elem{index: i}
        elms = append(elms, se)
        se.strVal = d[keyn]
        se.row    = d
        if se.strVal != "" {
            if dtOK {
                // If so far all non-empty values have been parseable as datetime, keep doing it.
                if dt, _fmt, err := TryParseStrToUnix(se.strVal, dtFmt);err == nil {
                    se.dtVal = dt
                    dtFmt = _fmt    // Save time. Once we parse the first one, continue with just that format.
                } else {
                    // This value is not a datetime, do not continue trying to parse as dt.
                    dtOK = false
                }
            }
            if fltOK {
                // If so far all non-empty values have been parseable as floats, keep doing it.
                if flt64, err := strconv.ParseFloat(se.strVal, 64);err == nil {
                    se.fltVal = flt64
                } else {
                    fltOK = false
                }
            }
            if intOK {
                // If so far all non-empty values have been parseable as ints, keep doing it.
                if i64, err := strconv.ParseInt(se.strVal, 10, 64);err == nil {
                    se.intVal = i64
                } else {
                    intOK = false
                }
            }
        }
    }

    // We now have the full list of sort_elem objects. Sort this list based on saved values.
    if dtOK {
        // All elems have datetimes (or blanks) as values. So sort as datetimes (converted to int64s).
        sort.SliceStable(elms, func(i, j int) bool {
            if desc {
                return elms[i].dtVal > elms[j].dtVal
            } else {
                return elms[i].dtVal < elms[j].dtVal
            }
        })
    } else if fltOK {
        // All elems have floats (or blanks) as values. So sort as floats.
        sort.SliceStable(elms, func(i, j int) bool {
            if desc {
                return int64(elms[i].fltVal) > int64(elms[j].fltVal)
            } else {
                return int64(elms[i].fltVal) < int64(elms[j].fltVal)
            }
        })
    } else if intOK {
        // All elems have int64s (or blanks) as values. So sort as int64s.
        sort.SliceStable(elms, func(i, j int) bool {
            if desc {
                return elms[i].intVal > elms[j].intVal
            } else {
                return elms[i].intVal < elms[j].intVal
            }
        })
    } else {
        // None of the above. They're just lowly strings!
        sort.SliceStable(elms, func(i, j int) bool {
            if desc {
                return strings.Compare(elms[i].strVal, elms[j].strVal) != -1
            } else {
                return strings.Compare(elms[i].strVal, elms[j].strVal) == -1
            }
        })
    }
    // The sort_elem list has been sorted. Now build the new master list using the new order.
    newRows := make([]data, 0, len(elms))
    for _, se := range elms {               // Here, this is the new correct order.
        elm := list[se.index]               // This is the elem in the current master list.
        newRows  = append(newRows, elm)     // Add it in sequence into the new master list.
    }
    return newRows
}

func (c *cache) toShortNames(row data) {
    // convert column names to short names, if possible.
    for coln := range row {
        // First look for a custom mapping
        if shrn, ok := c.hdrm[coln];ok {
            c.shrt[shrn] = coln
            c.full[coln] = shrn
            row[shrn] = row[coln]
            delete(row, coln)
            continue
        }
        shrn := ToShortName(coln)
        if shrn != "" && shrn != coln {
			c.shrt[shrn] = coln
			c.full[coln] = shrn
            row[shrn] = row[coln]
            delete(row, coln)
        }
    }
}

func (c *cache) toFullNames(row data) {
    // convert column names to full names, if possible.
    for shrn := range row {
		if fuln, ok := c.shrt[shrn];ok {
			row[fuln] = row[shrn]
			delete(row, shrn)
		}
    }
}
