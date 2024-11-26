package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

type rebate_file struct {
	path  string 			// Disk file.
	csep  string            // If set, the hdr/col separator in the input (defaults to ",")
	hdrs  []string          // Original values for CSV headers or table column names.
	hdrm  map[string]string // Maps custom input name to proper short name.
	hdri  map[int]string    // CSV column index => proper_hdr
	lines int
	rderr error
}

func new_rebate_file(opts *Opts) *rebate_file {
	rf := &rebate_file{
		path:  opts.fileIn,
		csep:  opts.csep,
		hdrm:  map[string]string{},
	}
	if opts.hdrm != "" {
		toks1 := strings.Split(opts.hdrm, ",")
		for _, tok1 := range toks1 {
			if tok1 != "" {
				toks2 := strings.Split(tok1, "=")
				if len(toks2) == 2 {
					cust := toks2[0]
					stnd := toks2[1]
					rf.hdrm[cust] = stnd
				}
			}
		}
	}
	return rf
}

func (rf *rebate_file) read() (chan *Rebate, error) {
    count := func(path string) int {
        if fd, err := os.Open(path); err == nil {
            defer fd.Close()
            buf := make([]byte, 32*1024)
            count := 0
            lineSep := []byte{'\n'}
            var last byte
            for {
                c, err := fd.Read(buf)
                count += bytes.Count(buf[:c], lineSep)
                if c != 0 {
                    last = buf[c-1]
                }
                switch {
                case err == io.EOF:
                    if last != lineSep[0] {
                        count++
                    }
                    return count
                case err != nil:
                    return count
                default:
                }
            }
        }
        return 0
    }
    readl := func(br *bufio.Reader, csep string) ([]string, error) {
        if line, _, err := br.ReadLine(); err == nil {
            str := string(line)
            if len(str) > 0 {
                toks := strings.Split(str, csep)
                vals := make([]string, 0, len(toks))
                for _, tok := range toks {
                    tok  = strings.ReplaceAll(tok, " ",  "")
                    tok  = strings.ReplaceAll(tok, "\t", "")
                    vals = append(vals, tok)
                }
                return vals, nil
            } else {
                return []string{}, nil
            }
        } else {
            return nil, err
        }
    }
	rf.hdrs  = []string{}
	rf.hdri  = map[int]string{}
    rf.lines = count(rf.path)
	stdh := []string{ "rxn", "hrxn", "ndc", "spid", "prid", "dos"}
    if fd, err := os.Open(rf.path); err == nil {
        defer fd.Close()
        br := bufio.NewReader(fd)
        if rf.hdrs, err = readl(br, rf.csep); err != nil {
            return nil, err
        }
		cust := 1
		for i, hdr := range rf.hdrs {			// The list of original headers on rebate CSV file.
			if prop, ok := rf.hdrm[hdr];ok {	// Requested mapping CUSTNAME=>PROPERNAME
				prop = strings.ToLower(prop)
				if slices.Contains(stdh, prop) {// Is PROPERNAME in standard list?
					rf.hdri[i] = prop			// Store the proper name. This is where we'll save.
				} else {						// Cannot map a custom name to another custom name!
					if cust <= 50 {				// Only keep up to 50 custom columns. But don't break out.
						rf.hdri[i] = fmt.Sprintf("col%d", cust)
						cust++
					}
				}
			} else {							// No custom mapping for this one. Use as is.
				hdr = strings.ToLower(hdr)
				if slices.Contains(stdh, hdr) {	// Is it a standard column?
					rf.hdri[i] = hdr
				} else {						// It's a custom column. Keep it in colX.
					if cust <= 50 {				// Only keep up to 50 custom columns. But don't break out.
						rf.hdri[i] = fmt.Sprintf("col%d", cust)
						cust++
					}
				}
			}
		}

        // Now that we've gotten the header stuff figured out we can start reading rows in the background.
        rbts := make(chan *Rebate, 1000)
        go func() {
            // The rest of the rows are data rows.
            for {
                if toks, err := readl(br, rf.csep); err == nil {
                    rbt  := &Rebate{}
                    for i, fld := range toks {
						if i >= len(rf.hdri) {
							break
						}
						col := rf.hdri[i]
						switch col {
						case "spid":
							rbt.Spid = fld
						case "prid":
							rbt.Prid = fld
						case "dos":
							rbt.Dos = fld
						case "rxn":
							rbt.Rxn = fld
						case "hrxn":
							rbt.Hrxn = fld
						case "ndc":
							rbt.Ndc = fld
						case "col1":
							rbt.Col1 = fld
						case "col2":
							rbt.Col2 = fld
						case "col50":
							rbt.Col50 = fld
						default:
						}
                    }
                    rbts <-rbt
                } else if err == io.EOF {
                    close(rbts)
                    break
                } else {
                    rf.rderr = err
                    close(rbts)
                    break
                }
            }
        }()
        return rbts, nil
    } else {
        return nil, err
    }
}
