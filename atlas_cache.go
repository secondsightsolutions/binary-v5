package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type keyn = string

type cache_set struct {
	esp1 *cache
	ents *cache
	ledg *cache
	ndcs *cache
	phms *cache
	spis *cache
	desg *cache
	ldns *cache
}
type cache struct {
	views map[keyn]*view	// Each view is an indexed representation of the set of elements.
	rows  []*row            // All elements in this table.
	seqn  int64				// Largest seq number in cache. Starting point for requesting new rows from source.
}
type view struct {
	rows map[any][]*row
	keyn string
	skey sort_key
}
type row struct {
	elem any
	indx int
}

type sort_key struct {
	keyn string
	desc bool
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

func (cs *cache_set) clone() *cache_set {
    ncs := &cache_set{
    	esp1: cs.esp1,
    	ents: cs.ents,
    	ledg: cs.ledg,
    	ndcs: cs.ndcs,
    	phms: cs.phms,
    	spis: cs.spis,
		desg: cs.desg,
		ldns: cs.ldns,
    }
    return ncs
}
func new_cache[T any]() *cache {
	ca := &cache{
		views: map[string]*view{},
		rows:  []*row{},
		seqn:  -1,
	}
	return ca
}
func load_cache[T any](done *sync.WaitGroup, c **cache, name string, f func(chan any, int64) chan *T) {
	var ca *cache
	if *c == nil {
		ca = new_cache[T]()
		*c = ca
	} else {
		ca = *c
	}
	go func() {
		defer done.Done()
		cnt  := 0
		seq  := ca.seqn
		strt := time.Now()
		rfl  := &rflt{}
		fm   := f(stop, seq)
		for {
			select {
			case <-stop:
				Log("atlas", "load_cache", name, "received stop signal, returning", time.Since(strt), map[string]any{"cnt": cnt, "seq": seq}, nil)
				return
			case obj, ok := <-fm:
				if !ok {
					Log("atlas", "load_cache", name, "cache loaded", time.Since(strt), map[string]any{"cnt": cnt, "manu": manu, "seq": seq}, nil)
					return
				}
				cnt++
				seq = rfl.getFieldValueAsInt64(obj, "Seq")
				ca.Add(obj)
				if seq > ca.seqn {
					ca.seqn = seq
				}
			}
		}
	}()
}

func (c *cache) Find(keyn, keyv string) []*row {
	if c == nil {
		return nil
	}
	vw := c.views[keyn]
	if vw == nil {
		return nil
	}
	rows := vw.find(keyv)
	return rows
}
// Sorts the main list based on this key.
func (c *cache) Sort(keyn string, desc bool) {
	c.rows = sort_list(keyn, desc, c.rows)
}
func (c *cache) Add(a any) {
	row := &row{indx: len(c.rows), elem: a}
	c.rows = append(c.rows, row)
	for _, view := range c.views {
		view.add(row)
	}
}
func (c *cache) CreateView(keyn string, desc bool) {
	if _, ok := c.views[keyn]; ok {
		return
	}
	view := &view{rows: map[any][]*row{}}
	view.keyn = keyn
	view.skey.keyn = keyn
	view.skey.desc = desc
	c.views[keyn] = view
	for _, row := range c.rows {
		view.add(row)
	}
}

func (v *view) add(r *row) {
	keyv := r.value(v.keyn)
	if _,ok := v.rows[keyv];!ok {
		v.rows[keyv] = make([]*row, 0, 1)
	}
	v.rows[keyv] = append(v.rows[keyv], r)
}
func (v *view) find(keyv any) []*row {
	return v.rows[keyv]
}

func (r *row) value(name string) any {
	var rfl rflt
	switch obj := r.elem.(type) {
	default:
		return rfl.getFieldValue(obj, name)
	}
}

// func sort_lists(keyn string, desc bool, lists ...[]*row) []*row {
// 	all := []*row{}
// 	set := map[int]any{}
// 	for _, list := range lists {                    // Look at each list passed in.
// 		all = append(all, list...)                  // Add each (sub) list into the big list.
// 	}
// 	all = sort_list(keyn, desc, all)          		// Sort the big list.
// 	unq := []*row{}                                 // Possibly a smaller list, to contain big list with dups removed.
// 	for _, row := range all {
// 		if _, ok := set[row.indx];!ok {             // We identify dups by their indx. Note we are not identifying dups by uniq attr!
// 			unq = append(unq, row)
// 			set[row.indx] = nil
// 		}
// 	}
// 	return unq
// }
func sort_list(keyn string, desc bool, list []*row) []*row {
	// Since we want to be careful not to pull in too much data, we'll read in the elems just
	// long enough to grab their current index and the value of the keyn attribute.
	// Once we have the full list of sort_elem objects we'll sort them, then use their new
	// order to create a new full list.
	// This function does not deal with views.

	elms  := []*sort_elem{}	// This is the list of tiny elems - contains just enough to do a sort.
	dtOK  := true           // Continue to convert as datetime
	fltOK := true           // Continue to convert as float
	intOK := true           // Continue to convert as integer
	blOK  := true           // Continue to convert as boolean
	dtFmt := ""
	for i, row := range list {
		se := &sort_elem{index: i, row: row}
		elms = append(elms, se)

		// Set one of the type values, whichever it is.
		if keyn == "indx" { // The list *should* be in indx order. Do we need this?
			se.intVal = int64(row.indx)
			dtOK = false
			fltOK = false
			blOK = false
		} else { // We're not using the index, we're using some other attribute to sort on.
			objv := row.value(keyn)
			switch val := objv.(type) {
			case string:
				// Convert the string to as many of the more refined types that we can.
				// However, once we see a type that we cannot convert to, stop trying to convert the remaining rows to those types.
				// This saves us on processing time.
				// For instance, converting a string to a time can be expensive. So as soon as we hit a row whose value cannot be
				// converted to a time, time is no longer ubiquitous across all rows, so stop trying to save value as a time. Won't
				// be able to sort on it when done, so no point in continuing with that type.
				if dtOK {
					if dt, _fmt, err := TryParseStrToUnix(val, dtFmt); err == nil {
						se.dtVal = dt // This value is a datetime! Save it. Will sort on this when done (if all rows have dt).
						dtFmt = _fmt  // Save time. Once we parse the first one, continue with just that format.
					} else {
						dtOK = false // This value is not a datetime, do not continue trying to parse as dt.
					}
				}
				if fltOK {
					// If so far all non-empty values have been parseable as floats, keep doing it.
					if flt64, err := strconv.ParseFloat(val, 64); err == nil {
						se.fltVal = flt64
					} else {
						fltOK = false
					}
				}
				if intOK {
					// If so far all non-empty values have been parseable as ints, keep doing it.
					if i64, err := strconv.ParseInt(val, 10, 64); err == nil {
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
				se.strVal = val // It's always at least a string.

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
	if dtOK { // All elems have datetimes (or blanks) as values. So sort as datetimes (converted to int64s).
		sort.SliceStable(elms, func(i, j int) bool {
			if desc {
				return elms[i].dtVal > elms[j].dtVal
			} else {
				return elms[i].dtVal < elms[j].dtVal
			}
		})
	} else if fltOK { // All elems have floats (or blanks) as values. So sort as floats.
		sort.SliceStable(elms, func(i, j int) bool {
			if desc {
				return int64(elms[i].fltVal) > int64(elms[j].fltVal)
			} else {
				return int64(elms[i].fltVal) < int64(elms[j].fltVal)
			}
		})
	} else if intOK { // All elems have int64s (or blanks) as values. So sort as int64s.
		sort.SliceStable(elms, func(i, j int) bool {
			if desc {
				return elms[i].intVal > elms[j].intVal
			} else {
				return elms[i].intVal < elms[j].intVal
			}
		})
	} else if blOK { // All elems have booleans (or blanks) as values. So sort as booleans.
		sort.SliceStable(elms, func(i, j int) bool {
			if desc { // true, then false
				return elms[i].boolVal && !elms[j].boolVal
			} else { // false, then true
				return !elms[i].boolVal && elms[j].boolVal
			}
		})
	} else { // None of the above. They're just lowly strings!
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
	for _, se := range elms { // Here, this is the new correct order (these are the sort elements).
		currRow := list[se.index]                          // This is the elem in the current master list. Get it using the new sort order.
		newRow := &row{elem: currRow.elem, indx: se.index} // Create a new row, but share the data element! Lets us have a new indx.
		newRows = append(newRows, newRow)                  // Add it in sequence into the new master list.
	}
	return newRows
}
