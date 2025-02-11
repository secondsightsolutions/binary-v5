package main

import (
	"bufio"
	"bytes"
	"os"
	"slices"
	"strings"
)

func import_file[T any](file string, csep string) ([]string, chan *T, error) {
	stdm := map[string]string{
		"rx_number": 			"rxn",
		"service_provider_id": 	"spid",
		"prescriber_id":		"prid",
		"date_of_service":		"dos",
	}
	dates := []string{"dos", "dop", "dof", "doc", "crat", "strt", "term", "xdat", "dlat", "xsat", "cpat"}
	if fd, err := os.Open(file); err == nil {
		objs := make(chan *T, 1000)
		rf   := &rflt{}
		flds := rf.fields(new(T))			// List of fields on target object.
		data := findAny("data", flds)
		bufr := bufio.NewReader(fd)
		if hdrs, hdrm := import_header(bufr, csep); hdrs != nil {
			setClearOrHashed := func(obj *T, hdr, val, clrF, hshF string, flds []string, done *bool) {
				// Prior call had identified/handled this field. Do not overwrite. Done.
				if *done {
					return
				}
				// For example, maybe hdr=rxn and clrF=rxn and hshF=hrxn
				if findAny(hdr, []string{hshF, clrF}) == "" {		// Is the requested header either the clear or hashed header? If not, we're done.
					return
				}
				hshd := Is64bitHash(val)							// The value is hashed. Can only write into hashed form of the header (hshF).
				_fld := clrF										// Start with one field. We'll switch over to the other if we need to.
				if hshd {											// The value is hashed, so can only go into the hashed field.
					_fld = hshF
				}
				if fld := findAny(_fld, flds); fld != "" {			// Does this field (clear or hashed) exist on the object?
					if slices.Contains(dates, _fld) {				// Is this (lower-cased) field/header expected to contain a date formatted string?
						if tm := ParseStrToTime(val); tm != nil {	// If this is a column that contains dates, try to parse it.
							rf.setFieldValue(obj, fld, tm)			// We have to convert to time.Time - the only way we can write into an int64 field.
						}											// If we don't switch to time.Time - will fail when we try to convert string ("2024-01-01") to int64.
					} else {										// Not one of the date formatted strings (mapped to int64), so either mapped to non-int64 or value *is* an int. 
						rf.setFieldValue(obj, fld, val)				// Write as a string to the field. If the field is int-ish, will be parsed to an int.
					}
					*done = true
				}
			}
			go func() {
				defer fd.Close()
				defer close(objs)
				for {
					if vals := import_line(bufr, csep); len(vals) > 0 {
						obj := new(T)
						for i, val := range vals {
							hdr := hdrm[i]
							if mhdr, ok := stdm[strings.ToLower(hdr)];ok {					// If this CSV column header is mapped to something else, switch it in now.
								hdr = mhdr
							}
							done := false
							setClearOrHashed(obj, hdr, val, "rxn",  "hrxn", flds, &done)	// If hdr is rxn or hrxn,  write into rxn (if not hashed)  or hrxn (if hashed).
							setClearOrHashed(obj, hdr, val, "spid", "hspd", flds, &done)	// If hdr is spid or hspd, write into spid (if not hashed) or hspd (if hashed).
							setClearOrHashed(obj, hdr, val, "dos",  "hdos", flds, &done)
							setClearOrHashed(obj, hdr, val, "prid", "hprd", flds, &done)
							if !done {
								if fld := findAny(hdr, flds); fld != "" {					// This header (CSV column header) has a corresponding field in the object.									
									rf.setFieldValue(obj, fld, val)
								}
							}
						}
						if data != "" {
							list := make([]string, 0, len(vals))
							for _, val := range vals {
								list = append(list, strings.ReplaceAll(val, ",", ""))
							}
							rf.setFieldValue(obj, data, strings.Join(list, ","))
						}
						objs <-obj
					} else {
						return
					}
				}
			}()
			return hdrs, objs, nil
		} else {
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}
}
func import_header(br *bufio.Reader, csep string) ([]string, map[int]string) {
	if hdrs := import_line(br, csep); hdrs != nil {
		fldi := map[int]string{}
		for i, hdr := range hdrs {				// The list of original headers on CSV file.
			fldi[i] = hdr
		}
		return hdrs, fldi
	} else {
		return nil, nil
	}
}
func import_line(br *bufio.Reader, csep string) []string {
	var sb bytes.Buffer
	read_line:
	for {
		if line, hasMore, err := br.ReadLine(); err == nil {
			sb.Write(line)
			if !hasMore {
				break
			}
		} else {
			return nil
		}
	}
	line := sb.String()
	if len(line) > 0 {
		toks := strings.Split(line, csep)
		vals := make([]string, 0, len(toks))
		for _, tok := range toks {
			tok  = strings.ReplaceAll(tok, " ",  "")
			tok  = strings.ReplaceAll(tok, "\t", "")
			vals = append(vals, tok)
		}
		return vals
	} else {
		goto read_line
	}
}