package main

import (
	"sync"

)

type metrics struct {
    sync.Mutex
	rbt_total   int64
    rbt_valid   int64
    rbt_matched int64
    rbt_nomatch int64
    rbt_invalid int64
    rbt_passed  int64
    rbt_failed  int64
    clm_total   int64
    clm_valid   int64
    clm_matched int64
    clm_nomatch int64
    clm_invalid int64
    spi_exact   int64
    spi_cross   int64
    spi_stack   int64
    spi_chain   int64
    dos_equ_doc int64
    dos_bef_doc int64
    dos_aft_doc int64
    dos_equ_dof int64
    dos_bef_dof int64
    dos_aft_dof int64
}

func (m *metrics) update_rbt(rbt data) {
    m.Lock()
    m.rbt_total++
    switch rbt[Fields.Stat] {
    case "matched":
        m.rbt_matched++
    case "nomatch":
        m.rbt_nomatch++
    case "invalid":
        m.rbt_invalid++
    case "pass":
        m.rbt_passed++
    case "fail":
        m.rbt_failed++
    default:
    }
    m.Unlock()
}

func (m *metrics) update_rbt_clm(rbt data, clm data) {
    doc := clm[Fields.Doc]  // Claims are copies. No need to lock.
    dof := clm[Fields.Dof]

    m.Lock()
    if clm != nil {
        if diff, err := dates.Compare(rbt[Fields.Dos], doc); err == nil {
            if diff == 0 {
                m.dos_equ_doc++
            } else if diff < 0 {
                m.dos_bef_doc++
            } else {
                m.dos_aft_doc++
            }
        }
        if diff, err := dates.Compare(rbt[Fields.Dos], dof); err == nil {
            if diff == 0 {
                m.dos_equ_dof++
            } else if diff < 0 {
                m.dos_bef_dof++
            } else {
                m.dos_aft_dof++
            }
        }
    }
    m.Unlock()
}
