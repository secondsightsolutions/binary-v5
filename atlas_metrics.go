package main

import "time"

type metrics struct {
    load_data       time.Duration
    load_claims     time.Duration
    load_esp1       time.Duration
    load_entities   time.Duration
    load_ledger     time.Duration
    load_ndcs       time.Duration
    load_pharms     time.Duration
    load_spis       time.Duration
    load_desg       time.Duration
    load_ldns       time.Duration
    init_spis       time.Duration
}

func (s *scrub) update_metrics(rbt *rebate) {
    s.lckM.Lock()
    defer s.lckM.Unlock()
    s.metr.RbtTotal++
    switch rbt.stat {
    case "matched":
        s.metr.RbtMatched++
    case "nomatch":
        s.metr.RbtNomatch++
    case "invalid":
        s.metr.RbtInvalid++
    case "pass":
        s.metr.RbtPassed++
    case "fail":
        s.metr.RbtFailed++
    default:
    }
    for _, sclm := range rbt.clms {
        clm  := sclm.gclm.clm
        doc  := clm.Doc
        dof  := atlas.dates.hashToDays[clm.Hdos]
        dos  := rbt.rbt.Dos
        diff := doc - dos
        if diff == 0 {
            s.metr.DosEquDoc++
        } else if diff > 0 {
            s.metr.DosBefDoc++
        } else {
            s.metr.DosAftDoc++
        }

        diff = dof - dos
        if diff == 0 {
            s.metr.DosEquDof++
        } else if diff > 0 {
            s.metr.DosBefDof++
        } else {
            s.metr.DosAftDof++
        }
        
        opts := s.plcy.options()
        if yes, how := CheckSPI(s, rbt.rbt.Spid, clm.Spid, opts.chains, opts.stacks); yes {
            switch how {
            case "exact":
                s.metr.SpiExact++
            case "cross":
                s.metr.SpiCross++
            case "stack":
                s.metr.SpiStack++
            case "chain":
                s.metr.SpiChain++
            default:
            }
        }
    }
}
