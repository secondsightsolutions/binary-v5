package main

type policy struct {
    prepRebates	func(*Scrub)
    prepClaims 	func(*Scrub)
    scrubRebate func(*Scrub, *Rebate)
    result      func(*Scrub, *Rebate) string
}

func getPolicy(manu string) *policy {
    switch manu {
    case "amgen": 
        return &policy{
        	prepRebates: amgenPrepRebates,
        	prepClaims:  amgenPrepClaims,
        	scrubRebate: amgenScrubRebate,
        	result:      amgenResult,
        }
    default:
        return nil
    }
}