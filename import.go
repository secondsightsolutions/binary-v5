package main

import (
	"bufio"
	"io"
	"strings"
)

func (sf *scrub_file) import_file() *cache {
	cach := &cache{views: map[string]*view{}, elems: []*elem{}, hdrs: []string{}}
	hdrM := map[string]string{}		// cust_hdr => proper_hdr
	hdrI := map[int]string{}		// CSV indx => proper_hdr
	if sf.hdrs != "" {					// rid=rebate_id,rxn=rx_number
		toks := strings.Split(sf.hdrs, ",")
		for _, tok := range toks {
			nvp := strings.Split(tok, "=")
			if len(nvp) == 2 {
				cust := nvp[0]
				prop := nvp[1]
				hdrM[cust] = prop
			}
		}
	}
	if sf.keyn == "" {
		sf.keyn = "indx"
		sf.keyl = -3
	}
	if sf.keyl == 0 {
		sf.keyl = 3
	}
	if sf.csep == "" {
		sf.csep = ","
	}
	rdr := bufio.NewReader(sf.rdr)
	for {
		if line, _, err := rdr.ReadLine(); err == nil {
			str  := string(line)
			str  = strings.ReplaceAll(str, " ",  "")
			str  = strings.ReplaceAll(str, "\t", "")
			toks := strings.Split(str, sf.csep)
			if len(cach.hdrs) == 0 {
				cach.hdrs = toks
				for i, hdr := range cach.hdrs {
					if prop, ok := hdrM[hdr];ok {	// If this hdr is mapped to another value
						hdrI[i] = prop				// then assume the other value is the 
					} else {						// proper value to use.
						hdrI[i] = hdr				// If no custom mapping, could still be a
					}								// custom column - just keep name as is.
				}
			} else {
				row := map[string]string{}
				for i, fld := range toks {
					if i < len(hdrI) {
						row[hdrI[i]] = fld
					}
				}
				cach.Add(row)
			}
		} else if err == io.EOF {
			break
		} else {
			break
		}
	}
	cach.Index(sf.keyn, sf.keyl)
	return cach
}