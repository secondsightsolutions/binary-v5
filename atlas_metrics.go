package main

import (
)


func (s *scrub) update_rbt(rbt *Rebate) {
    s.lckM.Lock()
    defer s.lckM.Unlock()
    s.metr.RbtTotal++
    switch rbt.Stat {
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
}

func (s *scrub) update_rbt_clm(rbt *Rebate, clm *Claim) {
    doc := clm.Doc
    dof := clm.Hdos
    s.lckM.Lock()
    defer s.lckM.Unlock()
    if clm != nil {
        if diff, err := dates.Compare(rbt.Dos, doc); err == nil {
            if diff == 0 {
                s.metr.DosEquDoc++
            } else if diff < 0 {
                s.metr.DosBefDoc++
            } else {
                s.metr.DosAftDoc++
            }
        }
        if diff, err := dates.Compare(rbt.Dos, dof); err == nil {
            if diff == 0 {
                s.metr.DosEquDof++
            } else if diff < 0 {
                s.metr.DosBefDof++
            } else {
                s.metr.DosAftDof++
            }
        }
    }
}

func (s *scrub) update_spi_counts(which string) {
    s.lckM.Lock()
    defer s.lckM.Unlock()
    switch which {
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