package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"

)

func import_file[T any](file string, csep string) ([]string, chan *T, error) {
	if fd, err := os.Open(file); err == nil {
		objs := make(chan *T, 1000)
		rf   := &rflt{}
		flds := rf.fields(new(T))			// List of fields on target object.
		data := findAny("data", flds)
		bufr := bufio.NewReader(fd)
		if hdrs, hdrm := import_header(bufr, csep); hdrs != nil {
			go func() {
				defer fd.Close()
				defer close(objs)
				for {
					if vals := import_line(bufr, csep); len(vals) > 0 {
						obj := new(T)
						for i, val := range vals {
							hdr := hdrm[i]
							if fld := findAny(hdr, flds); fld != "" {	// This header (CSV column header) has a corresponding field in the object.
								rf.setFieldValue(obj, fld, val)
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