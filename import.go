package main

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"
)

func (c *cache) getData(pool, tbln, manu string, filts map[string]string) {
    for row := range getData(pool, tbln, manu, filts) {
        c.toShortNames(row)
        c.Add(row)
    }
}

func (c *cache) getFile(name, path, csep string) error {
    lines := 0
    level := ScreenLevel.All
    if strings.EqualFold(Type, "proc") {
        if name != "rebates" {
            level = ScreenLevel.None
        }
    }
    if fd, err := os.Open(path); err == nil {
        if cnt, err := lineCounter(fd); err == nil {
            lines = cnt
            if csep != "" {     // If not FW (so there's a header row)
                if lines > 0 {
                    lines--
                }
            }
        } else {
            return err
        }
    } else {
        return err
    }
    now := time.Now()
    bar := 0
    if fd, err := os.Open(path); err == nil {
        defer fd.Close()
        br := bufio.NewReader(fd)
        hd := true
        rn := 0
        for {
            if line, _, err := br.ReadLine(); err == nil {
                str  := string(line)
                str  = strings.ReplaceAll(str, " ",  "")
                str  = strings.ReplaceAll(str, "\t", "")
                if len(str) == 0 {
                    continue
                }
                toks := strings.Split(str, csep)
                row  := map[string]string{}
                
                if hd {
                    // First line is the header row. Process the column headers.
                    c.hdrs = toks
                    for i, hdr := range toks {
                        if prop, ok := c.hdrm[hdr];ok {	// If this hdr is mapped to another value
                            c.hdri[i] = prop            // then assume the other value is the 
                        } else {						// proper value to use.
                            c.hdri[i] = hdr				// If no custom mapping, could still be a
                        }								// custom column - just keep name as is.
                    }
                    hd = false
                    continue
                }
                // The rest of the rows are data rows.
                for i, fld := range toks {
                    if i < len(c.hdri) {
                        row[c.hdri[i]] = fld
                    }
                }
                c.toShortNames(row)
                c.Add(row)
                
                //if (rn)%5000 == 0 {
                    screen(now, &bar, rn, lines, ScreenLevel.All, level, false, "loading %s", name)
                //}
                rn++
            } else if err == io.EOF {
                if rn > 0 {
                    rn--
                }
                break
            } else {
                screen(now, &bar, rn, lines, ScreenLevel.All, level, true, "loading %s", name)
                return err
            }
        }
        screen(now, &bar, rn, lines, ScreenLevel.All, level, true, "loading %s", name)
    } else {
        screen(now, &bar, 0, lines, ScreenLevel.All, level, true, "loading %s", name)
    }
    return nil
}
