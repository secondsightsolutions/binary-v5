package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)


func TestMain(t *testing.T) {
	sc := new_scrub(1)
    test := "test1"
    sc.sr.manu = "amgen"
	sc.sr.file("rebates", test + "/rebates.csv", nil)
	sc.sr.file("claims",  test + "/claims.csv",  nil)
    sc.sr.file("ldns",    test + "/ldns.csv",    nil)
    sc.sr.file("desigs",  test + "/desigs.csv",  nil)
    sc.sr.file("spis",    test + "/spis.csv",    nil)
    sc.sr.file("ndcs",    test + "/ndcs.csv",    nil)
    sc.sr.file("pharms",  test + "/pharms.csv",  nil)
    sc.sr.file("entities",test + "/entities.csv",nil)
    sc.sr.file("ledger",  test + "/ledger.csv",  nil)
	w,_ := os.Create(test + "/out.csv")
    bw  := bufio.NewWriter(w)
	sc.run(bw)
    bw.Flush()
	fmt.Println("done")
}

var amgenHeaders = ""
var amgenRebates = 
`rbid,dos,rxn,ndc
001,01/01/2024,12345678,88888888888
002,01/02/2024,22345678,99999999999
003,01/01/2024,12399999,77777777777
`
var amgenClaims = 
`clid,doc,rxn,ndc,cnfm,elig
001,01/01/2024,12345678,88888888888,true,true
002,01/02/2024,22345678,99999999999,true,true
`
var amgenResults = 
`stat,rbid,dos,rxn,ndc
matched,001,01/01/2024,12345678,88888888888
matched,002,01/02/2024,22345678,99999999999
nomatch,003,01/01/2024,12399999,77777777777
`
func TestAmgen(t *testing.T) {
    var b bytes.Buffer
    w  := bufio.NewWriter(&b)
    sc := new_scrub(1)
    sc.sr.manu = "amgen"
    sc.sr.data("rebates", amgenRebates, amgenHeaders, ",", "rxn")
    sc.sr.data("claims",  amgenClaims,  "",           ",", "rxn")
    sc.run(w)
    w.Flush()
    testResults(t, amgenResults, b.String())
}


func (sr *scrub_req) file(name, path string, params map[string]string) error {
    sf := &scrub_file{name: name, path: path, csep: ","}
    if params != nil {
        sf.hdrs = params[fmt.Sprintf("%s_hdrs", name)]
        sf.csep = params[fmt.Sprintf("%s_csep", name)]
        sf.keys = params[fmt.Sprintf("%s_keys", name)]
    }
    sr.files[name] = sf
    if fd, err := os.Open(path); err == nil {
        sf.rdr = fd
    } else {
        return err
    }
    return nil
}

func (sr *scrub_req) data(name, data string, hdrs, csep, keys string) {
    sf := &scrub_file{name: name, csep: ","}
    sr.files[name] = sf
    sf.hdrs = hdrs
    sf.keys = keys
    if csep != "" {
        sf.csep = csep
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