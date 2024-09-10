package main

import (
	"sync"

	"github.com/secondsightsolutions/binary-v4/api"
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

func (m *metrics) update_rbt(sc *scrub, rbt data) {
    m.Lock()
    m.rbt_total++
    switch rbt[api.Fields.Stat] {
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

func (m *metrics) update_rbt_clm(sc *scrub, rbt data, clm data) {
    // Get data from claim first. Don't like holding two locks at once.
    sc.cs["claims"].Lock()
    doc := clm[Fields.Doc]
    dof := clm[Fields.Dof]
    sc.cs["claims"].Unlock()

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

func (m *metrics) update_scrub(sc *scrub) {
    matched := int64(0)
    nomatch := int64(0)
    valid   := int64(0)
    invalid := int64(0)
    total   := int64(0)
	if c,ok := sc.cs["claims"]; ok {
        total = int64(len(c.rows))
        sc.cs["claims"].Lock()
		for _, clm := range sc.cs["claims"].rows {
            status := clm[Fields.Stat]
            if status == "matched" {
				matched++
				valid++
			} else if status == "invalid" {
				invalid++
			} else {
				nomatch++
				valid++
			}
		}
        sc.cs["claims"].Unlock()
	}
    m.Lock()
    m.clm_total   = total
    m.clm_matched = matched
    m.clm_invalid = invalid
    m.clm_nomatch = nomatch
    m.clm_valid   = valid
	m.rbt_valid   = m.rbt_total - m.rbt_invalid
	m.Unlock()
}

func (m *metrics) update_spi_counts(which string) {
    m.Lock()
    switch which {
    case "exact":
        m.spi_exact++
    case "cross":
        m.spi_cross++
    case "stack":
        m.spi_stack++
    case "chain":
        m.spi_chain++
    default:
    }
    m.Unlock()
}