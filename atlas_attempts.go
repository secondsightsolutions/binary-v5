package main

import (
	"bytes"
	"fmt"
)

type Attempts struct {
	list []*attempt
}
type attempt struct {
	rbt  string
	clm  string
	oper string
	code string
	desc string
}

func NewAttempts() *Attempts {
	return &Attempts{list: []*attempt{}}
}
func (a *Attempts) Add(rbt *Rebate, clm *Claim, oper, code, desc string) {
	// att := &attempt{oper: oper, code: code, desc: desc}
	// if rbt != nil {
	// 	att.rbt = rbt["indx"]
	// }
	// if clm != nil {
	// 	att.clm = clm["indx"]
	// }
	// a.list = append(a.list, att)
}
func (a *Attempts) String() string {
	var sb bytes.Buffer
	for i, att := range a.list {
		sb.WriteString(fmt.Sprintf("%s:%s:%s:%s:%s", att.rbt, att.clm, att.oper, att.code, att.desc))
		if i < len(a.list)-1 {
			sb.WriteString(";")
		}
	}
	return sb.String()
}
