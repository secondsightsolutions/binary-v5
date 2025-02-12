package main

import (
)

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
        clm := sclm.gclm.clm
        doc := clm.Doc
        dof := clm.Hdos
        if diff, err := dates.Compare(rbt.rbt.Dos, doc); err == nil {
            if diff == 0 {
                s.metr.DosEquDoc++
            } else if diff < 0 {
                s.metr.DosBefDoc++
            } else {
                s.metr.DosAftDoc++
            }
        }
        if diff, err := dates.Compare(rbt.rbt.Dos, dof); err == nil {
            if diff == 0 {
                s.metr.DosEquDof++
            } else if diff < 0 {
                s.metr.DosBefDof++
            } else {
                s.metr.DosAftDof++
            }
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
