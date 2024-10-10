package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

)

// scrubdir represents the scrub directory for all the uploaded tables for a scrub.
// Each uploaded table is represented by a CSV file with headers.
type scrubdir struct {
    root  string
    dirn  string
    files map[string]*scrubfile
}
type scrubfile struct {
	pool string
	oper string
	tbln string
    file *os.File
    hdrs []string
    bufw *bufio.Writer
    cnt  int64
}

func scrubDir(scid, root, manu, proc string) (*scrubdir) {
	 // 111_12345678_brg_amgen
    sd := &scrubdir{root: root, files: map[string]*scrubfile{}}
	sd.dirn = fmt.Sprintf("%s_%d_%s_%s", scid, time.Now().Unix(), strings.ReplaceAll(proc, "_", "-"), strings.ReplaceAll(manu, "_", "-"))
    os.Mkdir(sd.root, os.ModePerm)
	os.Mkdir(sd.root + "/" + sd.dirn, os.ModePerm)
	return sd
}
func (sd *scrubdir) scrubFile(pool, oper, tbln string, cols []string) (*scrubfile, error) {
	name := fmt.Sprintf("%s_%s_%s", pool, oper, strings.ReplaceAll(tbln, "_", "-"))
    file := sd.files[name]
    if file == nil {
        full := fmt.Sprintf("%s/%s/%s", sd.root, sd.dirn, name)
        file = &scrubfile{pool: pool, oper: oper, tbln: tbln, hdrs: []string{}}
        sd.files[name] = file
        if fd, err := os.Create(full); err == nil {
            file.file = fd
			file.bufw = bufio.NewWriter(fd)
            if oper == "insert" {
                file.hdrs = cols
            }
			return file, nil
        } else {
            return nil, err
        }
    } else {
		return file, nil
	}
}
func (sd *scrubdir) close(dest string) {
    for _, file := range sd.files {
        if file.bufw != nil {
            file.bufw.Flush()
        }
        file.file.Close()
    }
    os.Mkdir(dest, os.ModePerm)
    os.Rename(sd.root + "/" + sd.dirn, dest + "/" + sd.dirn)
}
func (sd *scrubdir) cancel() {
    for _, file := range sd.files {
        file.file.Close()
    }
    os.RemoveAll(sd.root + "/" + sd.dirn)
}
func (sf *scrubfile) insert(vals []string) {
	sf.cnt++
	if sf.cnt == 1 {
		sf.bufw.WriteString(strings.Join(sf.hdrs, ",") + "\n")
	} else {
		sf.bufw.WriteString(strings.Join(vals, ",") + "\n")
	}
}
func (sf *scrubfile) update(vals, where map[string]string) {
	var sb strings.Builder
	sb.WriteString("UPDATE " + sf.tbln)
	sb.WriteString(" SET ")
	colN := 0
	for attr, valu := range vals {
		colN++
		sb.WriteString(fmt.Sprintf("%s = '%s'", attr, valu))
		if colN < len(vals) {
			sb.WriteString(", ")
		} else {
			sb.WriteString(" ")
		}
	}
	if len(where) > 0 {
		sb.WriteString(" WHERE ")
		colN := 0
		for col, val := range where {
			if strings.Contains(val, "%") {
				sb.WriteString(col + " " + "like" + " '%" + val + "%'")
			} else {
				sb.WriteString(col + " = '" + val + "'")
			}
			colN++
			if colN < len(where) {
				sb.WriteString(" AND ")
			}
		}
		sf.bufw.WriteString(sb.String() + "\n")
	}
}
