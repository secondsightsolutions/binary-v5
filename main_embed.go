package main

import (
	_ "embed"
)

// Values that can be injected at build time
var (
	//go:embed embed/appl.txt
	appl string

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

	//go:embed embed/srvh.txt
	srvh string

	//go:embed embed/svch.txt
	svch string

	// service

	//go:embed embed/azac.txt
	azac string

	//go:embed embed/azky.txt
	azky string
)
