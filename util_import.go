package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func import_file[T any](file string, csep string, hdrm map[string]string) ([]string, chan *T, error) {
	fmt.Println("import_file starting")
	if fd, err := os.Open(file); err == nil {
		fmt.Println("import_file opened file")
		defer fd.Close()
		br := bufio.NewReader(fd)
		return import_bufio[T](br, csep, hdrm)
	} else {
		fmt.Printf("import_file failed to open=%s\n", err.Error())
		return nil, nil, err
	}
}
// func import_data[T any](data string, csep string, hdrm map[string]string) ([]string, chan *T, error) {
// 	br := bufio.NewReader(bytes.NewReader([]byte(data)))
// 	return import_bufio[T](br, csep, hdrm)
// }
func import_bufio[T any](br *bufio.Reader, csep string, hdrm map[string]string) ([]string, chan *T, error) {
	objs := make(chan *T, 1000)
	rf   := &rflt{}
	flds := rf.fields(new(T))			// List of fields on target object.
	if hdrs, fldi, err := import_header(br, csep, flds, hdrm); err == nil {
		go func() {
			for {
				if cols, err := import_line(br, csep); err == nil {
					obj := new(T)
					for i, col := range cols {
						if fld, ok := fldi[i]; ok {
							rf.setFieldValue(obj, fld, col)
						}
					}
					fmt.Println("import_bufio - pushing to channel")
					objs <-obj
				} else {
					return
					//close(objs)
				}
			}
		}()
		return hdrs, objs, nil
	} else {
		fmt.Printf("import_bufio - read header failed %s\n", err.Error())
		return nil, nil, err
	}
}
func import_header(br *bufio.Reader, csep string, flds []string, hdrm map[string]string) ([]string, map[int]string, error) {
	if hdrs, err := import_line(br, csep); err == nil {
		fldi := map[int]string{}
		cust := 1
		Case := cases.Title(language.AmericanEnglish)
		for i, hdr := range hdrs {				// The list of original headers on CSV file.
			if fld, ok := hdrm[hdr];ok {		// Requested mapping CUSTNAME=>PROPERNAME
				fldi[i] = Case.String(fld)		// Set fld to be the mapped name.		
			} else {
				fldi[i] = Case.String(hdr)		// Set fld to the value of the hdr.
			}
			// Now fldi[i] is either a proper name (matching a field name), or it's not (so it's custom).
			if !slices.Contains(flds, fldi[i]) {
				// This header (mapped or not), converted to Title case, does not have a field in the object.
				// So the column values for this column will go into one of the "ColN" fields.
				fldi[i] = Case.String(fmt.Sprintf("col%d", cust))
				cust++
			}
		}
		return hdrs, fldi, nil
	} else {
		return nil, nil, err
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