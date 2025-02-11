package main

var Stats = struct {
	matched 		string
	nomatch			string
	invalid			string
}{
	"matched",
	"nomatch",
	"invalid",
}

var Results = struct {
	no_match_rx		string
	no_match_spi	string
	no_match_ndc	string
	clm_used 		string
	dos_rbt_aft_clm string
	dos_clm_aft_rbt string
	old_rbt_new_clm string
	new_rbt_old_clm string
	clm_not_cnfm    string
	phm_not_desg    string
	inv_desg_type   string
	wrong_network   string
	not_elig_at_sub string
}{
	"no_match_rx",
	"no_match_spi",
	"no_match_ndc",
	"clm_used",
	"dos_rbt_aft_clm",
	"dos_clm_aft_rbt",
	"old_rbt_new_clm",
	"new_rbt_old_clm",
	"clm_not_cnfm",
	"phm_not_desg",
	"inv_desg_type",
	"wrong_network",
	"not_elig_at_sub",
}