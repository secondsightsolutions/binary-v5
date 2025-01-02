package main

import (
	_ "embed"
)

// Values that can be injected at build time
var (
	//go:embed embed/name.txt
	name string

	//go:embed embed/type.txt
	Type string // manu or proc

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

	//go:embed embed/cacr.txt
	cacr string

	//go:embed embed/phrs.txt
	phrs string

	//go:embed embed/salt.txt
	salt string

	//go:embed embed/shell_cert.txt
	shell_cert string

	//go:embed embed/shell_pkey.txt
	shell_pkey string

	//go:embed embed/atlas_grpc.txt
	atlas_grpc string

	//go:embed embed/atlas_cert.txt
	atlas_cert string

	//go:embed embed/atlas_pkey.txt
	atlas_pkey string

	//go:embed embed/atlas_ogtm.txt
	atlas_ogtm string

	//go:embed embed/atlas_ogky.txt
	atlas_ogky string

	//go:embed embed/titan_grpc.txt
	titan_grpc string

	//go:embed embed/titan_cert.txt
	titan_cert string

	//go:embed embed/titan_pkey.txt
	titan_pkey string

	//go:embed embed/titan_ogtm.txt
	titan_ogtm string

	//go:embed embed/titan_ogky.txt
	titan_ogky string

	//go:embed embed/citus_host.txt
	citus_host string

	//go:embed embed/citus_port.txt
	citus_port string

	//go:embed embed/citus_name.txt
	citus_name string

	//go:embed embed/citus_user.txt
	citus_user string

	//go:embed embed/citus_pass.txt
	citus_pass string

	//go:embed embed/espdb_host.txt
	espdb_host string
	
	//go:embed embed/espdb_port.txt
	espdb_port string

	//go:embed embed/espdb_name.txt
	espdb_name string

	//go:embed embed/espdb_user.txt
	espdb_user string

	//go:embed embed/espdb_pass.txt
	espdb_pass string

	//go:embed embed/atlas_host.txt
	atlas_host string

	//go:embed embed/atlas_port.txt
	atlas_port string

	//go:embed embed/atlas_name.txt
	atlas_name string

	//go:embed embed/atlas_user.txt
	atlas_user string

	//go:embed embed/atlas_pass.txt
	atlas_pass string

	//go:embed embed/titan_host.txt
	titan_host string

	//go:embed embed/titan_port.txt
	titan_port string

	//go:embed embed/titan_name.txt
	titan_name string

	//go:embed embed/titan_user.txt
	titan_user string

	//go:embed embed/titan_pass.txt
	titan_pass string
)
