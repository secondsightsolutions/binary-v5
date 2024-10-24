package main

import (
	"testing"
)

func TestMain(t *testing.T) {
    dflts()
    main()
}
func TestAmgen(t *testing.T) {
    dflts()
    manu = "amgen"
	scid = 1
	fin  = "rebates.csv"
	fout = "results.csv"
	auth = "1234"
	srvh = "127.0.0.1"
	srvp = 23460
    main()
}
