package main

import (
	_ "embed"
)

// Values that can be injected at build time
var (
	//go:embed embed/type.txt
	Type string

	//go:embed embed/manu.txt
	manu string

	//go:embed embed/vers.txt
	vers string

	//go:embed embed/desc.txt
	desc string

	//go:embed embed/envr.txt
	envr string

	//go:embed embed/hash.txt
	hash string

	//go:embed embed/pkey.txt
	pkey string

	//go:embed embed/cacr.txt
	cacr string

	//go:embed embed/mycr.txt
	cert string

	//go:embed embed/phrs.txt
	phrs string

	//go:embed embed/salt.txt
	salt string

	//go:embed embed/host.txt
	host string

	//go:embed embed/port.txt
	port string

)

var (
	test string	// test directory
	scid int64  // scrub id
	auth string
	name string
	fin  string
	fout string

	doPing bool
	doVers bool
	Http   bool
)
