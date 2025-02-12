package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Titan struct {
	titan TitanServer
	pools map[string]*pgxpool.Pool
	opts  *Opts
	TLSCert  *tls.Certificate
    X509cert *x509.Certificate
}

var titan *Titan

func run_titan(done *sync.WaitGroup, opts *Opts, stop chan any) {
	titan = &Titan{titan: &titanServer{}, pools: map[string]*pgxpool.Pool{}, opts: opts}

	titan.X509cert, titan.TLSCert = crypt_init("titan", "run_titan", 33, titan_cert, cacr, "", titan_pkey)
	
	titan.pools["citus"] = db_pool("titan", citus_host, citus_port, citus_name, citus_user, citus_pass, true)
	titan.pools["titan"] = db_pool("titan", titan_host, titan_port, titan_name, titan_user, titan_pass, true)
	titan.pools["esp"]   = db_pool("titan", espdb_host, espdb_port, espdb_name, espdb_user, espdb_pass, true)

	run_datab_ping( done, stop, "titan", 60, titan.pools)
	run_titan_sync( done, stop, "titan", 60, titan)
	run_grpc_server(done, stop, "titan", titan_grpc_port, titan.TLSCert, RegisterTitanServer, titan.titan, titanUnaryInterceptor, titanStreamInterceptor)
}

func run_titan_sync(done *sync.WaitGroup, stop chan any, appl string, intv int, titan *Titan) {
	durn := time.Duration(0)
	done.Add(1)
	go func() {
		defer done.Done()
		for {
			select {
			case <-time.After(durn):
				titan.sync(stop)
				durn = time.Duration(intv) * time.Minute
			case <-stop:
				Log(appl, "run_titan_sync", "", "received stop signal, returning", 0, nil, nil)
				return
			}
		}
	}()
}

func (titan *Titan) sync(stop chan any) {
	titan.syncClaims(stop)
	titan.syncSPIs(stop)
	titan.syncNDCs(stop)
	titan.syncEntities(stop)
	titan.syncPharmacies(stop)
	titan.syncESP1(stop)
	titan.syncEligibility(stop)
}

func titan_db_sync[T any](fmPn, fmTn, whr string, fmM *dbmap, toPn, toTn string, multitx bool, stop chan any) {
	strt := time.Now()
	fmP  := titan.pools[fmPn]
	fmT  := fmTn
	fmM.table(fmP, fmT)

	toP := titan.pools[toPn]
	toT := toTn
	toM := new_dbmap[T]()
	toM.table(toP, toTn)
	
	if chn, err := db_select[T](fmP, "titan", fmT, fmM, whr, "", stop); err == nil {
		cnt, seq, err := db_insert(toP, "titan", toT, toM, chn, 5000, "", false, multitx)
		Log("titan", "titan_db_sync", toTn, "sync completed", time.Since(strt), map[string]any{"cnt": cnt, "seq": seq}, err)
	} else {
		Log("titan", "titan_db_sync", toTn, "sync completed", time.Since(strt), nil, err)
	}
}

func (titan *Titan) syncClaims(stop chan any) {
	seq, _ := db_max(titan.pools["titan"], "titan.claims", "seq")
	whr := fmt.Sprintf("manufacturer = 'astrazeneca' AND COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)*1000000, 0), 0) > %d", seq)
	fmM := new_dbmap[Claim]()
	fmM.column("chnm", "chain_name", 				"COALESCE(chain_name, '')")
	fmM.column("cnfm", "claim_conforms_flag", 		"COALESCE(claim_conforms_flag, true)")
	fmM.column("seq",  "created_at",  				"COALESCE(TRUNC(EXTRACT(EPOCH FROM created_at)   *1000000, 0), 0)")
	fmM.column("doc",  "created_at",  				"created_at")
	fmM.column("dop",  "formatted_dop",  			"formatted_dop")
	fmM.column("dos",  "formatted_dos",  			"formatted_dos")
	fmM.column("hdop", "date_prescribed", 			"COALESCE(date_prescribed, '')")
	fmM.column("hdos", "date_of_service", 			"COALESCE(date_of_service, '')")
	fmM.column("hfrx", "formatted_rx_number", 		"COALESCE(formatted_rx_number, '')")
	fmM.column("hrxn", "rx_number", 				"COALESCE(rx_number, '')")
	fmM.column("i340", "id_340b", 					"SPLIT_PART(COALESCE(id_340b, ''), '-', 1)")
	fmM.column("manu", "manufacturer", 				"COALESCE(manufacturer, '')")
	fmM.column("ndc",  "ndc",  						"REPLACE(COALESCE(ndc, ''), '-', '')")
	fmM.column("netw", "network", 					"COALESCE(network, '')")
	fmM.column("prnm", "product_name", 				"COALESCE(product_name, '')")
	fmM.column("qty",  "quantity",  				"COALESCE(quantity, 0)")
	fmM.column("clid", "short_id", 					"COALESCE(short_id, '')")
	fmM.column("spid", "service_provider_id", 		"COALESCE(service_provider_id, '')")
	fmM.column("prid", "prescriber_id", 			"COALESCE(prescriber_id, '')")
	fmM.column("elig", "eligible_at_submission",	"COALESCE(eligible_at_submission, true)")
	fmM.column("susp", "suspended_submission", 		"COALESCE(suspended_submission, false)")
	fmM.column("ihph", "in_house_pharmacy_ids", 	"array_to_string(in_house_pharmacy_ids, ',')")
	titan_db_sync[Claim]("citus", "public.submission_rows", whr, fmM, "titan", "titan.claims", true, stop)
}

func (titan *Titan) syncSPIs(stop chan any) {
	seq, _ := db_max(titan.pools["titan"], "titan.spis", "seq")
	whr := fmt.Sprintf("id > %d", seq)
	fmM := new_dbmap[SPI]()
	fmM.column("ncp", "ncpdp_provider_id", 			"COALESCE(ncpdp_provider_id, '')")
	fmM.column("npi", "national_provider_id", 		"COALESCE(national_provider_id, '')")
	fmM.column("dea", "dea_registration_id", 		"COALESCE(dea_registration_id, '')")
	fmM.column("sto", "store_number", 				"COALESCE(store_number, '')")
	fmM.column("lbn", "legal_business_name", 		"COALESCE(legal_business_name, '')")
	fmM.column("cde", "status_code_340b", 			"COALESCE(status_code_340b, '')")
	fmM.column("chn", "chain_name", 				"COALESCE(chain_name, '')")
	fmM.column("nam", "name", 						"COALESCE(name, '')")
	fmM.column("seq", "id",                         "COALESCE(id, 0)")
	titan_db_sync[SPI]("esp", "public.ncpdp_providers", whr, fmM, "titan", "titan.spis", true, stop)
}

func (titan *Titan) syncNDCs(stop chan any) {
	seq, _ := db_max(titan.pools["titan"], "titan.ndcs", "seq")
	whr := fmt.Sprintf("id > %d", seq)
	fmM := new_dbmap[NDC]()
	fmM.column("ndc",  "item", 						"COALESCE(REPLACE(item, '-', ''), '')")
	fmM.column("name", "product_name", 				"COALESCE(product_name, '')")
	fmM.column("netw", "network", 					"COALESCE(network, '')")
	fmM.column("manu", "manufacturer_name", 		"COALESCE(manufacturer_name, '')")
	fmM.column("seq", "id",                         "COALESCE(id, 0)")
	titan_db_sync[NDC]("esp", "public.ndcs", whr, fmM, "titan", "titan.ndcs", true, stop)
}

func (titan *Titan) syncEntities(stop chan any) {
	seq, _ := db_max(titan.pools["titan"], "titan.entities", "seq")
	whr := fmt.Sprintf("id > %d", seq)
	fmM := new_dbmap[Entity]()
	fmM.column("i340", "id_340b",  					"COALESCE(id_340b, '')")
	fmM.column("state", "state", 					"COALESCE(state, '')")
	fmM.column("strt", "participating_start_date",	"to_date(participating_start_date, 'YYYY-MM-DD')")
	fmM.column("term", "term_date",  				"to_date(term_date, 'YYYY-MM-DD')")
	fmM.column("seq", "id",                         "COALESCE(id, 0)")
	titan_db_sync[Entity]("esp", "public.covered_entities", whr, fmM, "titan", "titan.entities", true, stop)
}

func (titan *Titan) syncPharmacies(stop chan any) {
	seq, _ := db_max(titan.pools["titan"], "titan.pharmacies", "seq")
	whr := fmt.Sprintf("id > %d", seq)
	fmM := new_dbmap[Pharmacy]()
	fmM.column("chnm", "chain_name",  				"COALESCE(chain_name, '')")
	fmM.column("i340", "id_340b",  					"COALESCE(id_340b, '')")
	fmM.column("phid", "pharmacy_id", 				"COALESCE(pharmacy_id, '')")
	fmM.column("dea",  "dea_id",   					"COALESCE(dea_id, '')")
	fmM.column("npi",  "national_provider_id",		"COALESCE(national_provider_id, '')")
	fmM.column("ncp",  "ncpdp_provider_id",   		"COALESCE(ncpdp_provider_id, '')")
	fmM.column("deas", "dea",  						"array_to_string(dea, ',')")
	fmM.column("npis", "npi",  						"array_to_string(npi, ',')")
	fmM.column("ncps", "ncpdp",  					"array_to_string(ncpdp, ',')")
	fmM.column("state", "pharmacy_state", 			"COALESCE(pharmacy_state, '')")
	fmM.column("seq", "id",                         "COALESCE(id, 0)")
	titan_db_sync[Pharmacy]("esp", "public.contracted_pharmacies", whr, fmM, "titan", "titan.pharmacies", true, stop)
}

func (titan *Titan) syncESP1(stop chan any) {
	whr := ""
	fmM := new_dbmap[ESP1PharmNDC]()
	fmM.column("manu", "manufacturer",              "COALESCE(manufacturer, '')")
	fmM.column("spid", "service_provider_id",		"COALESCE(service_provider_id, '')")
	fmM.column("ndc",  "ndc",  						"COALESCE(ndc, '')")
	fmM.column("strt", "start", 					"start")
	fmM.column("term", "term", 						"term")
	titan_db_sync[ESP1PharmNDC]("citus", "public.esp1_providers", whr, fmM, "titan", "titan.esp1", true, stop)
}

func (titan *Titan) syncEligibility(stop chan any) {
	seq, _ := db_max(titan.pools["titan"], "titan.eligibility", "seq")
	whr := fmt.Sprintf("id > %d", seq)
	fmM := new_dbmap[Eligibility]()
	fmM.column("i340", "id_340b", 					"COALESCE(id_340b, '')")
	fmM.column("phid", "pharmacy_id", 				"COALESCE(pharmacy_id, '')")
	fmM.column("manu", "manufacturer", 				"COALESCE(manufacturer, '')")
	fmM.column("netw", "network", 					"COALESCE(network, '')")
	fmM.column("strt", "start_at", 					"start_at")
	fmM.column("term", "end_at", 					"end_at")
	fmM.column("seq", "id",                         "COALESCE(id, 0)")
	titan_db_sync[Eligibility]("citus", "public.eligibility_ledger", whr, fmM, "titan", "titan.eligibility", true, stop)
}