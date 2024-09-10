package main

import (
    "bufio"
    "io"
    "os"
    "strings"
)

func (c *cache) getData(pool, tbln, manu string, filts map[string]string) {
    for row := range getData(pool, tbln, manu, filts) {
        c.toShortNames(row)
        c.Add(row)
    }
}

func (c *cache) getFile(path, csep string) {
    if fd, err := os.Open(path); err == nil {
        defer fd.Close()
        br := bufio.NewReader(fd)
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
                
                if rn == 0 {
                    // First line is the header row. Process the column headers.
                    c.hdrs = toks
                    for i, hdr := range toks {
                        if prop, ok := c.hdrm[hdr];ok {	// If this hdr is mapped to another value
                            c.hdri[i] = prop            // then assume the other value is the 
                        } else {						// proper value to use.
                            c.hdri[i] = hdr				// If no custom mapping, could still be a
                        }								// custom column - just keep name as is.
                    }
                } else {
                    // The rest of the rows are data rows.
                    for i, fld := range toks {
                        if i < len(c.hdri) {
                            row[c.hdri[i]] = fld
                        }
                    }
                    c.toShortNames(row)
                    c.Add(row)
                }
                rn++
            } else if err == io.EOF {
                break
            } else {
                break
            }
        }
    }
}
