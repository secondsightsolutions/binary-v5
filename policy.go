package main

type policy struct {
    prepRebates	func(*scrub)
    prepClaims 	func(*scrub)
    scrubRebate func(*scrub, data)
    result      func(*scrub, data) string
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