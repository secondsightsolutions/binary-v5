package main

import (
    "strings"
    "sync"
)

// type SPI struct {
//     Ncp string
//     Npi string
//     Dea string
//     Sto string
//     Nam string
//     Lbn string
//     Chn string
//     Cde string // NCPDP 340b status code
// }
// type Pharmacy struct {
// 	I340  string
// 	Phid  string
// 	Ncp   string
// 	Npi   string
// 	Dea   string
// 	Chnm  string
// 	State string
// 	Ncps  []string
// 	Npis  []string
// 	Deas  []string
// }

type SPIs struct {
    sync.Mutex

    // The IDs for NPI, DEA, and NCPDP exist in different spaces. IOW they each have a different length.
    // NPI   - 10 digits
    // NCPDP - 7 or 8 digits
    // DEA   - 2 characters followed by 7 digits

    // Because of this, we can put all IDs into the same map. Makes lookups very easy.
    // Also, the NPI and NCPDP values are unique in the table, and more importantly, as a pair they are unique.
    // The DEA values are replicated across many SPIs, so this means we need to look at DEA ids differently.
    // While an NPI or NCPDP value will always take us to just one SPI, a DEA points to many. So if we get a
    // DEA value we'll have to match it to many SPIs. Not hard, it just means that we'll need a second map for
    // the DEA IDs, where each DEA ID maps to a list of SPIs.

    idMap  map[string]*SPI   // The unique NPI/NCPDP IDs point to SPIs
    deaMap map[string][]*SPI // One DEA can be across many NPI or NCPDP IDs.

    stacks map[string]map[string]interface{}

}

func newSPIs() *SPIs {
    spis := &SPIs{ 
        idMap:  make(map[string]*SPI), 
        deaMap: make(map[string][]*SPI), 
        stacks: make(map[string]map[string]interface{}),
    }
    return spis
}

func (spis *SPIs) load(c *cache) {
    if c == nil {
        return
    }
    spis.Lock()
    defer spis.Unlock()
    if len(spis.idMap) == 0 && len(spis.deaMap) == 0 {
		for _, row := range c.rows {
            spi := row.elem.(*SPI)
			spis.addSPI(spi)
        }
	}
}

func (spis *SPIs) addToStacks(lists... []string) {
    set := make(map[string]interface{})
    for _, list := range lists {
        for _, id := range list {
            if id != "" {
                set[id] = nil
            }
        }
    }
    for id1 := range set {
        for id2 := range set {
            if id2 == "" {
                continue
            }
            if id1 == id2 {
                continue
            }
            if _, ok := spis.stacks[id1]; !ok {
                spis.stacks[id1] = make(map[string]interface{})
            }
            spis.stacks[id1][id2] = nil
        }
    }
}

func (spis *SPIs) addSPI(spi *SPI) {
    spis.idMap[spi.Ncp] = spi
    spis.idMap[spi.Npi] = spi

    if spi.Dea != "" {
        if _, ok := spis.deaMap[spi.Dea]; !ok {
            spis.deaMap[spi.Dea] = make([]*SPI, 0, 2)
        }
        spis.deaMap[spi.Dea] = append(spis.deaMap[spi.Dea], spi)
    }
    
    // Let's make sure that the singular three values are also present in the stacks. Probably not necessary since we search
    // both the scalar and stacked values when running find.
    spis.addToStacks([]string{spi.Dea}, []string{spi.Dea, spi.Ncp, spi.Npi})
    spis.addToStacks([]string{spi.Ncp}, []string{spi.Dea, spi.Ncp, spi.Npi})
    spis.addToStacks([]string{spi.Npi}, []string{spi.Dea, spi.Ncp, spi.Npi})
}

func (spis *SPIs) match(spidA, spidB string, useChains, useStacks bool) (bool, string) {
    // The ID string values are identical. An exact match.
    if spidA == spidB {
        return true, "exact"
    }

    // The ID values are different but they point to the same SPI object (implying an NPI and NCPDP pair of values).
    // This only works for NCPDP and NPI - a DEA ID may have multpile SPIs. Grab low-hanging fruit first, since in most
    // cases the two IDs will be NPI and/or NCPDP.
    if spiA, ok := spis.idMap[spidA]; ok {
        if spiB, ok := spis.idMap[spidB]; ok {
            if spiA == spiB {
                return true, "cross"
            }
        }
    }

    // No luck. Either these two IDs don't map together, or one/both are DEA values, or we need to check stacks and chains.
    // Must now deal with lists of SPI objects. Let's first get our SPI object(s). 
    // If they are the same SPI ID, then they are an "exact" match (we handled this above).
    // If they are the same SPI, then they are "cross" between NPI and/or NCPDP (we handled this above).
    // If they are NPI/NCPDP SPI => [ DEA SPI, ...], then also a "cross" match.
    // If they are in the stacks map (anyID => {anyOtherIDa, anyOtherIDb}) then it's a "stack" match.
    // If they are different SPIs, check their chain names - may be a "chain" match.

    // Gets one or multple SPIs for the given ID, always returned as a list for consistent handling below.
    getSPIs := func(id string) []*SPI {
        spi := spis.idMap[id]
        if spi == nil {
            if list, ok := spis.deaMap[id]; ok {
                return list
            }
            return []*SPI{}
        } else {
            return []*SPI{spi}
        }
    }

    spisA := getSPIs(spidA) // NPI/NCPDP give singular SPI, DEA may provide multiple.
    spisB := getSPIs(spidB)
    
    // If we're here, we have spisA and spisB. These lists may have one or many SPI objects.
    // Only check for cross matches here (no chains or stacks). We want to prioritize a cross match over stacks/chains.
    for _, spiA := range spisA {        // This could be singular (NPI or NCPDP) or multiple (DEA)
        for _, spiB := range spisB {    // This could be singular (NPI or NCPDP) or multiple (DEA)
            if spiA == spiB {
                return true, "cross"
            }
        }
    }

    // Cross match check failed. Next is stacking.
    // Stacking is a little different. We don't use the SPIs. Instead we have a simple map: idX => {idY, idZ}.
    // We only have to search in one direction (spidA => spidB?) because we added them in both directions.
    if useStacks {
        if ids, ok := spis.stacks[spidA]; ok {
            if _, ok := ids[spidB]; ok {
                return true, "stack"
            }
        }
    }
    
    // Last check is pharmacy chains. A little more fuzzy, so we save it for last. 
    // To do chain matching we'll need the SPIs. We still have them from above.
    if useChains {
        for _, spiA := range spisA {
            for _, spiB := range spisB {
                if spiA.Chn != "" && strings.EqualFold(spiA.Chn, spiB.Chn) {
                    return true, "chain"
                }
            }
        }
    }
    return false, ""
}

/*
// addPharmacy allows us to add a chain name to the SPI we already loaded (in the database the chain name is stored in the separate contracted_pharmacies table).
// Note that we don't allow a Pharmacy to define an SPI - that must be done up above. Should we allow a CP in the database to create a new SPI here?
func (spis *SPIs) addPharmacy(phm *Pharmacy) {
    // Chains - a common chaim name between pharmacies allows us to cross-walk between pharmacies.
    if phm.Chnm != "" {
        // If a pharmacy has a dea value, set the chain name on the matching SPI(s).
        // Small issue here; these SPIs are probably also/already pointed to by NCPDP/NPI ids. Maybe less accurate? So do it first.
        if phm.Dea != "" {
            for _, spi := range spis.deaMap[phm.Dea] {
                spi.Chn = phm.Chnm
            }
        }
        // Chain name is in Pharmacy, not SPI, so we need to copy over the chain name to the SPI(s) (*should* be one/same SPI here).
        if spi, ok := spis.idMap[phm.Ncp]; ok {
            spi.Chn = phm.Chnm
        }
        if spi, ok := spis.idMap[phm.Npi]; ok {
            spi.Chn = phm.Chnm
        }
    }

    // Stacks - a pharmacy can have zero or more NPIs, DEAs, and NCPDPs. The Pharmacy struct gives us these arrays of ids.
    // Simply go through all the ids on the pharmacy (Pharmacy) and put each one in as a key, with the data portion being
    // all the other ids on the pharmacy.
    // If a pharmacy multiple ids, then this will allow a crosswalk to the other ids on the pharmacy. Note we are already
    // doing this in its initial form of the three singular values (npi <-> dea? <-> ncp) on this one SPI. We're now
    // just expanding this concept by putting *all* the ids (from the Pharmacy) into a relationship map.
    // So why not just add these additional ids into the original maps above? (idMap, deaMap). So that we can
    // turn on or off this "stacking" feature separately from the basic cross-walking alread provided.

    // A pharmacy might look like this:
    // CVS999 => DEAS['1111', '1112'], NPIS['2221', '2222', '2223', '2224'], NCP['3331'] (all these ids on this pharmacy)

    // We want to end up with this map:
    // Key      Values
    // '1111'   ['1111', '1112', '2221', '2222', '2223', '2224', '3331']
    // '1112'   ['1111', '1112', '2221', '2222', '2223', '2224', '3331']
    // '2221'   ['1111', '1112', '2221', '2222', '2223', '2224', '3331']
    // '2222'   ['1111', '1112', '2221', '2222', '2223', '2224', '3331']
    // '2223'   ['1111', '1112', '2221', '2222', '2223', '2224', '3331']
    // '2224'   ['1111', '1112', '2221', '2222', '2223', '2224', '3331']
    // '3331'   ['1111', '1112', '2221', '2222', '2223', '2224', '3331']

    spis.addToStacks([]string{phm.Dea}, []string{phm.Npi}, []string{phm.Ncp}, phm.Deas, phm.Npis, phm.Ncps)
}

func (spis *SPIs) getI340Status(spid string) string {
    spis.Lock()
    defer spis.Unlock()
    if spi := spis.find(spid); spi != nil {
        return spi.Cde
    }
    return ""
}

func (spis *SPIs) match2(spidA, spidB string, useChains bool) bool {
    res, _ := spis.match(spidA, spidB, useChains, false)
    return res
}
func (spis *SPIs) match3(spidA, spidB string, useChains, useStacks bool) bool {
    res, _ := spis.match(spidA, spidB, useChains, useStacks)
    return res
}

func (spis *SPIs) find(spid string) *SPI {
    if spi, ok := spis.idMap[spid]; ok {
        return spi
    }
    if list, ok := spis.deaMap[spid]; ok {
        for _, spi := range list {
            if spi.Dea == spid {
                return spi
            }
        }
    }
    return nil
}
*/
