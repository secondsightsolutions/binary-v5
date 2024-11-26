package main

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CA struct {
    clms *cache
	esp1 *cache
	ents *cache
	ledg *cache
	ndcs *cache
	phms *cache
	spis *cache
    done bool
}
type cache struct {
    sync.Mutex
    views map[string]*view  // Each view is an indexed representation of the set of elements.
    rows  []*row            // All elements in this table.
    //sfile *scrub_file       // If we were loaded via a scrub_file.
}
type view struct {
    rows map[string][]*row
    vkey view_key
    skey sort_key
}
type row struct {
    elem any
    indx int
}

type sort_key struct {
    keyn string
    keyf string
    desc bool
}
type view_key struct {
    keyn string
    keyf string
    keyl int
}
type sort_elem struct {
    index   int
    strVal  string
    dtVal   int64
    intVal  int64
    fltVal  float64
    boolVal bool
    row     *row
}

func new_cache[T any](name string, list []*T) *cache {
    strt := time.Now()
    ca := &cache{
    	views: map[string]*view{},
    	rows:  []*row{},
    }
    for _, obj := range list {
        ca.Add(obj)
    }
    // At this point the main list is naturally sorted by the index.
    log("atlas", "new_cache", "%-21s / %-20s / rows=%d", time.Since(strt), nil, "cache loaded", name, len(ca.rows))
    return ca
}

func (c *cache) Find(keyn, full string) []*row {
    if c == nil {
        return nil
    }
    c.Lock()
    defer c.Unlock()
    vw := c.views[keyn]
    if vw == nil {
        return nil
    }
    rows := vw.find(full)
    return rows
}
// Sorts the main list based on this key.
func (c *cache) Sort(keyn string, keyf string, keyl int, desc bool) {
    c.Lock()
    defer c.Unlock()
    c.rows = sort_list(keyn, keyf, desc, c.rows)
}
func (c *cache) Add(a any) {
    c.Lock()
    defer c.Unlock()
    row := &row{indx: len(c.rows), elem: a}
    c.rows = append(c.rows, row)
    for _, view := range c.views {
        view.add(row)
    }
}
func (c *cache) CreateView(keyn string, keyf string, keyl int, desc bool) {
    c.Lock()
    defer c.Unlock()
    if _, ok := c.views[keyn]; ok {
        return
    }
    view := &view{rows: map[string][]*row{}}
    view.vkey.keyn = keyn
    view.vkey.keyf = keyf
    view.vkey.keyl = keyl
    view.skey.keyn = keyn
    view.skey.keyf = keyf
    view.skey.desc = desc
    c.views[keyn] = view
    for _, row := range c.rows {
        view.add(row)
    }
}

func (v *view) viewKeyVal(row *row) string {
    val := rowVal(row, v.vkey.keyf)
    vkv := viewKeyVal(val, v.vkey.keyl)
    return vkv
}
func (v *view) add(r *row) {
    val := v.viewKeyVal(r)
    if _,ok := v.rows[val];!ok {
        v.rows[val] = make([]*row, 0, 1)
    }
    v.rows[val] = append(v.rows[val], r)
}
func (v *view) find(full string) []*row {
    fnd  := []*row{}
    srch := viewKeyVal(full, v.vkey.keyl)
    all  := v.rows[srch]
    for _, row := range all {
        val := rowVal(row, v.vkey.keyf)
        if strings.EqualFold(val, full) {
            fnd = append(fnd, row)
        }
    }
    return fnd
}

func sort_lists(keyn, keyf string, desc bool, lists ...[]*row) []*row {
    all := []*row{}
    set := map[int]any{}
    for _, list := range lists {                    // Look at each list passed in.
        all = append(all, list...)                  // Add each (sub) list into the big list.
    }
    all = sort_list(keyn, keyf, desc, all)          // Sort the big list.
    unq := []*row{}                                 // Possibly a smaller list, to contain big list with dups removed.
    for _, row := range all {
        if _, ok := set[row.indx];!ok {             // We identify dups by their indx. Note we are not identifying dups by uniq attr!
            unq = append(unq, row)
            set[row.indx] = nil
        }
    }
    return unq
}
func sort_list(keyn, keyf string, desc bool, list []*row) []*row {
    // Since we want to be careful not to pull in too much data, we'll read in the elems just
    // long enough to grab their current index and the value of the keyn attribute.
    // Once we have the full list of sort_elem objects we'll sort them, then use their new
    // order to create a new full list.
    // This function does not deal with views.

    elms  := []*sort_elem{} // This is the list of tiny elems - contains just enough to do a sort.
    dtOK  := true           // Continue to convert as datetime
    fltOK := true           // Continue to convert as float
    intOK := true           // Continue to convert as integer
    blOK  := true           // Continue to convert as boolean
    dtFmt := ""
    for i, row := range list {
        se := &sort_elem{index: i, row: row}
        elms = append(elms, se)

        // Set one of the type values, whichever it is.
        if keyn == "indx" {     // The list *should* be in indx order. Do we need this?
            se.intVal = int64(row.indx)
            dtOK  = false
            fltOK = false
            blOK  = false
        } else {                // We're not using the index, we're using some other attribute to sort on.
            objv := objVal(row.elem, keyf)
            switch val := objv.(type) {
            case string:
                // Convert the string to as many of the more refined types that we can.
                // However, once we see a type that we cannot convert to, stop trying to convert the remaining rows to those types.
                // This saves us on processing time.
                // For instance, converting a string to a time can be expensive. So as soon as we hit a row whose value cannot be
                // converted to a time, time is no longer ubiquitous across all rows, so stop trying to save value as a time. Won't
                // be able to sort on it when done, so no point in continuing with that type.
                if dtOK {
                    if dt, _fmt, err := TryParseStrToUnix(val, dtFmt);err == nil {
                        se.dtVal = dt   // This value is a datetime! Save it. Will sort on this when done (if all rows have dt).
                        dtFmt = _fmt    // Save time. Once we parse the first one, continue with just that format.
                    } else {
                        dtOK = false    // This value is not a datetime, do not continue trying to parse as dt.
                    }
                }
                if fltOK {
                    // If so far all non-empty values have been parseable as floats, keep doing it.
                    if flt64, err := strconv.ParseFloat(val, 64);err == nil {
                        se.fltVal = flt64
                    } else {
                        fltOK = false
                    }
                }
                if intOK {
                    // If so far all non-empty values have been parseable as ints, keep doing it.
                    if i64, err := strconv.ParseInt(val, 10, 64);err == nil {
                        se.intVal = i64
                    } else {
                        intOK = false
                    }
                }
                if blOK {
                    if bl, err := strconv.ParseBool(val); err == nil {
                        se.boolVal = bl
                    } else {
                        blOK = false
                    }
                }
                se.strVal = val     // It's always at least a string.

            case float32:
                se.fltVal = float64(val)
                dtOK  = false
                intOK = false
                blOK  = false
            case float64:
                se.fltVal = val
                dtOK  = false
                intOK = false
                blOK  = false
            case int:
                se.intVal = int64(val)
                dtOK  = false
                fltOK = false
                blOK  = false
            case int32:
                se.intVal = int64(val)
                dtOK  = false
                fltOK = false
                blOK  = false
            case int64:
                se.intVal = val
                dtOK  = false
                fltOK = false
                blOK  = false
            case bool:
                se.boolVal = val
                dtOK  = false
                fltOK = false
                intOK = false
            default:
            }
        }
    }

    // We now have the full list of sort_elem objects. Sort this list based on saved values.
    if dtOK {           // All elems have datetimes (or blanks) as values. So sort as datetimes (converted to int64s).
        sort.SliceStable(elms, func(i, j int) bool {
            if desc {
                return elms[i].dtVal > elms[j].dtVal
            } else {
                return elms[i].dtVal < elms[j].dtVal
            }
        })
    } else if fltOK {   // All elems have floats (or blanks) as values. So sort as floats.
        sort.SliceStable(elms, func(i, j int) bool {
            if desc {
                return int64(elms[i].fltVal) > int64(elms[j].fltVal)
            } else {
                return int64(elms[i].fltVal) < int64(elms[j].fltVal)
            }
        })
    } else if intOK {   // All elems have int64s (or blanks) as values. So sort as int64s.
        sort.SliceStable(elms, func(i, j int) bool {
            if desc {
                return elms[i].intVal > elms[j].intVal
            } else {
                return elms[i].intVal < elms[j].intVal
            }
        })
    } else if blOK {    // All elems have booleans (or blanks) as values. So sort as booleans.
        sort.SliceStable(elms, func(i, j int) bool {
            if desc {   // true, then false
                return elms[i].boolVal && !elms[j].boolVal
            } else {    // false, then true
                return !elms[i].boolVal && elms[j].boolVal
            }
        })
    } else {            // None of the above. They're just lowly strings!
        sort.SliceStable(elms, func(i, j int) bool {
            if desc {
                return strings.Compare(elms[i].strVal, elms[j].strVal) != -1
            } else {
                return strings.Compare(elms[i].strVal, elms[j].strVal) == -1
            }
        })
    }
    // The sort_elem list has been sorted. Now build the new master list using the new order.
    newRows := make([]*row, 0, len(elms))
    for _, se := range elms {                               // Here, this is the new correct order (these are the sort elements).
        currRow  := list[se.index]                          // This is the elem in the current master list. Get it using the new sort order.
        newRow   := &row{elem: currRow.elem, indx: se.index}// Create a new row, but share the data element! Lets us have a new indx.
        newRows  = append(newRows, newRow)                  // Add it in sequence into the new master list.
    }
    return newRows
}
func rowVal(row *row, keyf string) string {
    return strVal(row.elem, keyf)
}
func strVal(obj any, keyf string) string {
    str := ""
    val := objVal(obj, keyf)
    if val != nil {
        switch val := val.(type) {
        case string:
            str = val
        case bool:
            str = strconv.FormatBool(val)
        case int:
            str = fmt.Sprintf("%d", val)
        case int32:
            str = fmt.Sprintf("%d", val)
        case int64:
            str = fmt.Sprintf("%d", val)
        case time.Time:
            str = val.Format("2006/01/02 15:04:05.000000")
        case *time.Time:
            str = val.Format("2006/01/02 15:04:05.000000")
        default:
        }
    }
    return str
}
func objVal(obj any, keyf string) any {
    res := reflect.ValueOf(obj).MethodByName(keyf).Call([]reflect.Value{})
    if len(res) > 0 {
        return res[0].Interface()
    }
    return nil
}
func viewKeyVal(val string, keyl int) string {
    if keyl <= 0 {
        return val
    }
    if val == "" {
        return val
    }
    if len(val) > keyl {
        return val[len(val)-keyl:]
    }
    return val
}