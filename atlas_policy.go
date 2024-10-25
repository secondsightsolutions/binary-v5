package main

type Policy struct {
	prepRebates func(*Scrub)
	prepClaims  func(*Scrub)
	scrubRebate func(*Scrub, *Rebate)
	result      func(*Scrub, *Rebate) string
}

func GetPolicy(manu string) *Policy {
	switch manu {
	case "amgen":
		return &Policy{
			prepRebates: amgenPrepRebates,
			prepClaims:  amgenPrepClaims,
			scrubRebate: amgenScrubRebate,
			result:      amgenResult,
		}
	default:
		return nil
	}
}
