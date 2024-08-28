package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type cache struct {
    sync.Mutex
    name  string            // Name of the cache ("rebates", "claims", "ndcs", etc)
    views map[string]*view  // Each view is an indexed representation of the set of elements.
    elems []*elem           // All elements in this table.
    inmem int               // Count of elems that have their Data loaded in memory.
    head *chunk             // Front of elem LRU (newest).
    tail *chunk             // Back of elem LRU (oldest, next to go!).
    max  int                // Maximum we'd like to keep in memory (-1 means no max).
    hdrs []string
}

type view struct {
    keyn    string		    // The name of the attribute (the index) splitting the chunks.
    keyl    int             // The length of the key value (like 3). Negative means start from end.
    chunks  map[string]*chunk
}

type chunk struct {
    prfx    string          // The substring (start or end) of the index key's value - all elems in this chunk share this same index key value substring.
    elems   []*elem         // Point to entries in cache.elems.
    next    *chunk          // LRU pointer to next chunk.
    prev    *chunk          // LRU pointer to prev chunk.
    loaded  bool            // The chunk is loaded, although individual elements may no longer be! Different views (and even chunks) can share elems.
}

type elem struct {
    Data    data            // The data row itself - this will be nillified when we unload the elem (the elem structures and their chunks remain in memory).
    index   int             // The position within the global list (need this for resorting).
    dirty   bool            // The row was written to, and must be flushed out to disk before we can nillify.
}

type sort_elem struct {
    index   int
    value   string
    dtVal   int64
    intVal  int64
    fltVal  float64
}

type data = map[string]string

// Sorts the main list and the chunks based on this key.
func (c *cache) Sort(keyn string, asc bool) {
    // Since we want to be careful not to pull in too much data, we'll read in the elems just
    // long enough to grab their current index and the value of the keyn attribute.
    // Once we have the full list of sort_elem objects we'll sort them, then use their new
    // order to rewrite the full list (and then the sublists within the chunks).
    c.Lock()
    defer c.Unlock()
    elms := []*sort_elem{}  // This is the list of tiny elems - contains just enough to do a sort.
    // Since we can only load by chunk, grab the first view (any will do) and go through their
    // chunks, loading each chunk and then grabbing the current global index of that elem along
    // with the value we're sorting by.
    dtOK  := true           // Continue to convert as datetime
    fltOK := true           // Continue to convert as float
    intOK := true           // Continue to convert as integer
    dtFmt := ""
    for _, vw := range c.views {
        for _, chnk := range vw.chunks {
            c.load(chnk, c.chunkFile(vw.keyn, chnk.prfx))
            for _, elem := range chnk.elems {
                val := strings.ToLower(elem.Data[keyn])
                se  := &sort_elem{index: elem.index, value: val}
                // Skip conversions if the value is empty string. We'll deal with this later during sort.
                if val != "" {
                    if dtOK {
                        // If so far all non-empty values have been parseable as datetime, keep doing it.
                        if dt, _fmt, err := TryParseDateToUnix(val, dtFmt);err == nil {
                            se.dtVal = dt
                            dtFmt = _fmt    // Save time. Once we parse the first one, continue with just that format.
                        } else {
                            // This value is not a datetime, do not continue trying to parse as dt.
                            dtOK = false
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
                }
                elms = append(elms, se)     // Add to list of sort_elem (in current sort order).
            }
        }
        break   // After the first view, that was all we needed. Break out so we can sort.
    }
    // We now have the full list of sort_elem objects. Sort this list based on saved values.
    if dtOK {
        // All elems have datetimes (or blanks) as values. So sort as datetimes (converted to int64s).
        sort.SliceStable(elms, func(i, j int) bool {
            if asc {
                return elms[i].dtVal < elms[j].dtVal
            } else {
                return elms[i].dtVal > elms[j].dtVal
            }
        })
    } else if fltOK {
        // All elems have floats (or blanks) as values. So sort as floats.
        sort.SliceStable(elms, func(i, j int) bool {
            if asc {
                return int64(elms[i].fltVal) < int64(elms[j].fltVal)
            } else {
                return int64(elms[i].fltVal) > int64(elms[j].fltVal)
            }
        })
    } else if intOK {
        // All elems have int64s (or blanks) as values. So sort as int64s.
        sort.SliceStable(elms, func(i, j int) bool {
            if asc {
                return elms[i].intVal < elms[j].intVal
            } else {
                return elms[i].intVal > elms[j].intVal
            }
        })
    } else {
        // None of the above. They're just lowly strings!
        sort.SliceStable(elms, func(i, j int) bool {
            if asc {
                return strings.Compare(elms[i].value, elms[j].value) == -1
            } else {
                return strings.Compare(elms[i].value, elms[j].value) != -1
            }
        })
    }
    // The sort_elem list has been sorted. Now build the new master list using the new order.
    newElems := make([]*elem, 0, len(elms))
    for i, se := range elms {
        elm := c.elems[se.index]            // This is the elem in the current master list.
        elm.index = i                       // This is the new index value.
        newElems  = append(newElems, elm)   // Add it in sequence into the new master list.
    }
    c.elems = newElems                      // Set the cache to use this new master list.
    // Finally, let's reorder our chunks! This is easier, now that we have the new index values.
    for _, vw := range c.views {
        for _, chnk := range vw.chunks {
            c.load(chnk, c.chunkFile(vw.keyn, chnk.prfx))
            sort.SliceStable(chnk.elems, func(i, j int) bool {
                return chnk.elems[i].index < chnk.elems[j].index
            })
        }
    }
}

func (c *cache) Add(d data) {
    c.Lock()
    defer c.Unlock()
    e := &elem{Data: d}
    e.index = len(c.elems)
    e.Data["indx"] = fmt.Sprintf("%d", e.index)
    c.elems = append(c.elems, e)
    for keyn := range c.views {
        c.indexElem(keyn, e)
    }
}
func (c *cache) Index(keyn string, keyl int) {
    c.Lock()
    defer c.Unlock()
    if _, ok := c.views[keyn]; ok {
        return
    }
    vw := &view{keyn: keyn, keyl: keyl, chunks: map[string]*chunk{}}
    c.views[keyn] = vw
    for _, e := range c.elems {
        c.indexElem(keyn, e)
    }
}
func (c *cache) indexElem(keyn string, e *elem) {
    vw := c.views[keyn]
    val := e.Data[keyn]
    prf := vw.prefix(val)
    if _, ok := vw.chunks[prf];!ok {
        vw.chunks[prf] = &chunk{prfx: prf, elems: []*elem{}}
    }
    chnk := vw.chunks[prf]
    chnk.elems = append(chnk.elems, e)
    c.toFront(chnk)
}

func (c *cache) Find(keyn, val string) []data {
    // Note that reading in a sequence of elems from a chunk file will actually update the cache elems.
    // The elems slice in the chunk is always populated - the elements in this slice are pointers to the
    // elements in the main cache elems slice.
    // So the chunk elems are pointers to the elems in the cache elems, and unloading a chunk means going
    // into the elems in the cache and nilling out their Data.
    // This means that over time it's possible that a chunk may be "loaded" and still have a couple missing
    // elems - this would happen if two chunks shared some common elems (very likely when more than one view) and one chunk gets "unloaded".
    // One chunk might say loaded, the other unloaded, and some elems in the loaded chunk may actually be nilled out,
    // and some elems in the unloaded chunk may have its elem Data.
    // Load and unload broadly at the chunk level, but double-check each elem to make sure it's loaded.
    // The threshold counts are on elems, not chunks.
    // It would be *much* more accurate if we kept an LRU of elems, not chunks. But in practice, trying for such
    // fine-grained LRU management would become ridiculously expensive, especially considering the increase in disk I/O.
    // Disk path names
    // {base}/{cachename}/{viewkeyn}/{chunkprefix}.json
    // ./.hidden/claims/rx_number/123.json
    // ./.hidden/claims/formatted_rx/123.json

    c.Lock()
    defer c.Unlock()

    if _, ok := c.views[keyn];!ok {                     // Key might be "rx_number"
        c.Unlock()
        c.Index(keyn, 3)
        c.Lock()
    }
    rows := []data{}
    vw   := c.views[keyn]
    prfx := vw.prefix(val)                          // Prefix might be "1234" (of "123456-78")
    file := c.chunkFile(keyn, val)
    if chnk, ok := vw.chunks[prfx];ok {             // If the elem exists at all (disk or mem), we'll see its chunk here.
        c.load(chnk, file)                          // If chunk already loaded, does nothing. But will load elem gaps if any.
        c.toFront(chnk)                             // We're using the chunk so refresh it to the front of the LRU list.
        for _, e := range chnk.elems {              // Now look at each elem to find one or more matches. Prefix matches, so maybe?
            if v, ok := e.Data[vw.keyn]; ok {
                if strings.EqualFold(v, val) {
                    rows = append(rows, e.Data)
                }
            }
        }
    }
    // We may have brought some chunk data into memory. If so, need to flush out chunk(s) at the tail
    for tc := c.tail; tc != nil; tc = tc.prev {
        if c.inmem >= c.max && c.max != -1 {
            file := fmt.Sprintf("%s/%s/%s/%s.json", ".hidden", c.name, vw.keyn, tc.prfx)
            c.unload(tc, file)
        } else {
            break
        }
    }
    return rows
}

func (c *cache) chunkFile(keyn, val string) string {
    vw   := c.views[keyn]
    prfx := vw.prefix(val)
    file := fmt.Sprintf("%s/%s/%s/%s.json", ".hidden", c.name, vw.keyn, prfx)
    return file
}
func (c *cache) toFront(chnk *chunk) {
    // First handle the head and tail pointers.
    if c.head == nil {
        c.head = chnk
    } else if c.head != chnk {
        chnk.next   = c.head
        c.head.prev = chnk
        c.head      = chnk
    }
    if c.tail == nil {
        c.tail = chnk
    } else if c.tail == chnk {
        if chnk.prev != nil {   // If no one behind us, we must stay as the tail (as well as the head).
            chnk.prev.next = nil
            c.tail = chnk.prev
        }
    }
    // Since we're at the front of the list, no one is before us.
    chnk.prev = nil
}

func (v *view) prefix(val string) string {
    if v.keyl == 0 {
        v.keyl = 3      // Hopefully this is unnecessary!
    }
    if v.keyl > 0 {
        if len(val) >= v.keyl {
            return val[:v.keyl]
        } else {
            return val
        }
    } else {
        keyl := v.keyl * -1
        if len(val) >= keyl {
            return val[len(val)-keyl:]
        } else {
            return val
        }
    }
}

func (c *cache) load(chnk *chunk, file string) {
    _, unloaded, _ := chnk.counts()
    if unloaded > 0 {                   // If all elems are actually loaded, no need to go to disk at all.
        data := chnk.read(file)         // At least one elem is not loaded, so read from disk.
        chnk.fill(data)                 // Fill in the gaps (the unloaded elems).
    }
}
func (c *cache) unload(chnk *chunk, file string) {
    loaded, unloaded, clean := chnk.counts()
    if !clean {                         // Do we have data to be written out?
        if unloaded > 0 {               // Do we have any gaps that first need to be pulled in from the chunk on disk?
            data := chnk.read(file)     // This reads the whole chunk, without gaps - but we still have in-mem elems that are dirty.
            chnk.fill(data)             // Fill in the gaps (the unloaded elems)
        }
        chnk.write(file)                // Any gaps in-mem were filled, so write the full chunk (with the dirty changes)
        for _, e := range chnk.elems {
            e.dirty = false             // Can't be dirty. We just wrote it, and now we're purging the in-mem data.
            e.Data  = nil
        }
    }
    chnk.loaded = false
    c.inmem -= loaded
}

func (chnk *chunk) fill(diskData []data) {
    for i, e := range chnk.elems {
        if !e.dirty {                   // Can update the in-mem if it's dirty (okay, cuz we'd fail the next test anyway).
            if e.Data == nil {          // We're not dirty, but do we already have the data? If so, skip (no diff really).
                e.Data = diskData[i]    // Not in-mem, so put the elem Data we read from disk back into the cache elem.
            }
        }
    }
}

func (chnk *chunk) counts() (loaded, unloaded int, clean bool) {
    clean = true
    for _, e := range chnk.elems {
        if e.dirty {
            e.dirty = false
            clean   = false
        }
        if e.Data == nil {
            unloaded++
        } else {
            loaded++
        }
    }
    return
}

func (chnk *chunk) read(file string) []map[string]string {
    rows := []data{}
    if fd, err := os.Open(file); err == nil {
        defer fd.Close()
        if bytes, err := io.ReadAll(fd); err == nil {
            if err := json.Unmarshal(bytes, &rows); err == nil {
                return rows
            } else {
                fmt.Printf("chunk read of %s failed: %s\n", file, err.Error())
            }
        } else {
            fmt.Printf("chunk read of %s failed: %s\n", file, err.Error())
        }
    } else {
        fmt.Printf("chunk read of %s failed: %s\n", file, err.Error())
    }
    return rows
}
func (chnk *chunk) write(file string) {
    if fd, err := os.Open(file); err == nil {
        defer fd.Close()
        if bytes, err := json.Marshal(chnk.elems); err == nil {
            if n, err := fd.Write(bytes); err == nil {
                if n != len(bytes) {
                    fmt.Printf("chunk write to %s failed: wrote %d of %d bytes\n", file, n, len(bytes))
                }
            } else {
                fmt.Printf("chunk write to %s failed: %s\n", file, err.Error())
            }
        } else {
            fmt.Printf("chunk write to %s failed: %s\n", file, err.Error())
        }
    } else {
        fmt.Printf("chunk open of %s failed: %s\n", file, err.Error())
    }
}