package main

import (
	"testing"
)

func TestAmgen(t *testing.T) {
    dflts()
    manu = "amgen"
	scid = 1
	fin  = "rebates.csv"
	fout = "results.csv"
    main()
}
