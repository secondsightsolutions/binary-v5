package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)


func TestMain(t *testing.T) {
	sc := new_scrub()
	sc.sr.file("rebates", "rebates.csv", nil)
	sc.sr.file("claims",  "claims.csv",  nil)
	sc.sr.manu = "amgen"
	w,_ := os.Create("out.csv")
	sc.run(w)
	fmt.Println("done")
}

var amgenHeaders = ""
var amgenRebates = 
`rbid,dos,rxn,ndc
001,01/01/2024,12345678,88888888888
002,01/02/2024,22345678,99999999999
003,01/01/2024,12399999,77777777777
`
var amgenClaims = `
clid,dos,rxn,ndc
001,01/01/2024,12345678,88888888888
002,01/02/2024,22345678,99999999999
`
var amgenResults = 
`stat,rbid,dos,rxn,ndc
nomatch,001,01/01/2024,12345678,88888888888
nomatch,002,01/02/2024,22345678,99999999999
nomatch,003,01/01/2024,12399999,77777777777
`
func TestAmgen(t *testing.T) {
    var b bytes.Buffer
    w  := bufio.NewWriter(&b)
    sc := new_scrub()
    sc.sr.manu = "amgen"
    sc.sr.data("rebates", amgenRebates, amgenHeaders, ",", "rxn", "")
    sc.sr.data("claims",  amgenClaims,  "",           ",", "rxn", "")
    sc.run(w)
    w.Flush()
    testResults(t, amgenResults, b.String())
}


func (sr *scrub_req) file(name, file string, params map[string]string) error {
    sf := &scrub_file{name: name, file: file, keyl: 3, csep: ","}
    if params != nil {
        sf.hdrs = params[fmt.Sprintf("%s_hdrs", name)]
        sf.csep = params[fmt.Sprintf("%s_csep", name)]
        sf.keyn = params[fmt.Sprintf("%s_keyn", name)]
        keyln  := params[fmt.Sprintf("%s_keyl", name)]
        if keyln != "" {
            if keyl, err := strconv.ParseInt(keyln, 10, 64); err == nil {
                sf.keyl = int(keyl)
            }
        }
    }
    sr.files[name] = sf
    if fd, err := os.Open(file); err == nil {
        sf.rdr = fd
    } else {
        return err
    }
    return nil
}

func (sr *scrub_req) data(name, data string, hdrs, csep, keyn, keyl string) {
    sf := &scrub_file{name: name, keyl: 3, csep: ","}
    sr.files[name] = sf
    sf.hdrs = hdrs
    sf.keyn = keyn
    if csep != "" {
        sf.csep = csep
    }
    if keyl != "" {
        if keyln, err := strconv.ParseInt(keyl, 10, 64); err == nil {
            sf.keyl = int(keyln)
        }
    }
    sf.rdr = strings.NewReader(data)
}

func testResults(t *testing.T, expected, actual string) {
    t.Helper()
    parse := func(str string) [][]string {
        rows := [][]string{}
        scnr := bufio.NewScanner(strings.NewReader(str))
        for scnr.Scan() {
            line := scnr.Text()
            toks := strings.Split(line, ",")
            rows = append(rows, toks)
        }
        return rows
    }
    exp := parse(expected)
    act := parse(actual)

    if len(exp) != len(act) {
        t.Fatalf("expected has %d rows, actual has %d rows", len(exp), len(act))
        return
    }
    for i, expRow := range exp {
        actRow := act[i]
        if len(expRow) != len(actRow) {
            t.Fatalf("row %d... expected has %d cols, actual has %d cols", i, len(expRow), len(actRow))
            return
        }
        for j, expCol := range expRow {
            actCol := actRow[j]
            if expCol != actCol {
                t.Fatalf("row %d, col %d... expected has %s, actual has %s", i, j, expCol, actCol)
                return
            }
        }
    }
}