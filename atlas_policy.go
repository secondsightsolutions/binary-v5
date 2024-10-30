package main

type Policy struct {
	prepRebates func(*scrub)
	prepClaims  func(*scrub)
	scrubRebate func(*scrub, *Rebate)
	result      func(*scrub, *Rebate) string
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
