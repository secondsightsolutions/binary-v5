package main

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// func import_file[T any](file string, csep string, hdrm map[string]string) []*T {
// 	if fd, err := os.Open(file); err == nil {
// 		defer fd.Close()
// 		br := bufio.NewReader(fd)
// 		return import_bufio[T](br, csep, hdrm)
// 	}
// 	return nil
// }
func import_data[T any](data string, csep string, hdrm map[string]string) []*T {
	br := bufio.NewReader(bytes.NewReader([]byte(data)))
	return import_bufio[T](br, csep, hdrm)
}
func import_bufio[T any](br *bufio.Reader, csep string, hdrm map[string]string) []*T {
	rf  := &rflt{}
	lst := new_slice_ptr[T](0, 10)
	if hdri, err := import_header(br, csep, hdrm); err == nil {
		for {
			if cols, err := import_line(br, csep); err == nil {
				obj := new(T)
				for i, col := range cols {
					if fld, ok := hdri[i]; ok {
						rf.setFieldValue(obj, fld, col)
					}
				}
				lst = append(lst, obj)
			} else if err == io.EOF {
				break
			} else {
				break
			}
		}
	}
	return lst
}
func import_header(br *bufio.Reader, csep string, hdrm map[string]string) (map[int]string, error) {
	if hdrs, err := import_line(br, csep); err == nil {
		hdri := map[int]string{}
		for i, hdr := range hdrs {				// The list of original headers on CSV file.
			if prop, ok := hdrm[hdr];ok {		// Requested mapping CUSTNAME=>PROPERNAME
				hdr = prop				
			}
			hdri[i] = hdr
		}
		return hdri, nil
	} else {
		return nil, err
	}
}
func import_line(br *bufio.Reader, csep string) ([]string, error) {
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