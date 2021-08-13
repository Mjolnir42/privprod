/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package flowdata // import "github.com/mjolnir42/privprod/internal/flowdata"

import "time"

// IOC represents a stripped down version of the information contained
// inside a record, suitable for comparing against IOCs
type IOC struct {
	AgentID   string    `json:"AgentID"`
	Address   string    `json:"Address"`
	IPVersion uint8     `json:"IPVersion"`
	Start     time.Time `json:"DateTimeStart"`
	End       time.Time `json:"DateTimeEnd"`
}

// ToIOC exports the IOC relevant information from a record for
// a given address addr
func (r Record) ToIOC(addr string) IOC {
	return IOC{
		AgentID:   r.AgentID,
		Address:   addr,
		IPVersion: r.IPVersion,
		Start:     r.StartMilli.UTC(),
		End:       r.EndMilli.UTC(),
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
