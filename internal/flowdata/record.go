/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package flowdata // import "github.com/mjolnir42/privprod/internal/flowdata"

import "time"

type Record struct {
	OctetCount     uint64    `json:"OctetCount"`
	PacketCount    uint64    `json:"PacketCount"`
	ProtocolID     uint8     `json:"ProtocolID"`
	Protocol       string    `json:"Protocol,omitempty"`
	IPVersion      uint8     `json:"IPVersion"`
	SrcAddress     string    `json:"SrcAddress"`
	SrcPort        uint16    `json:"SrcPort"`
	DstAddress     string    `json:"DstAddress"`
	DstPort        uint16    `json:"DstPort"`
	TcpControlBits Bitmask   `json:"TcpControlBits"`
	TcpFlags       Flags     `json:"TcpFlags"`
	IngressIf      uint32    `json:"-"`
	EgressIf       uint32    `json:"-"`
	FlowDirection  uint8     `json:"-"`
	StartMilli     time.Time `json:"StartDateTimeMilli"`
	EndMilli       time.Time `json:"EndDateTimeMilli"`
	AgentID        string    `json:"AgentID"`
	RecordID       string    `json:"RecordID"`
	ExpIPv4Addr    string    `json:"-"`
	ExpIPv6Addr    string    `json:"-"`
	ExpPID         uint32    `json:"-"`
}

func (r Record) Copy() Record {
	return Record{
		OctetCount:     r.OctetCount,
		PacketCount:    r.PacketCount,
		ProtocolID:     r.ProtocolID,
		Protocol:       r.Protocol,
		IPVersion:      r.IPVersion,
		SrcAddress:     r.SrcAddress,
		SrcPort:        r.SrcPort,
		DstAddress:     r.DstAddress,
		DstPort:        r.DstPort,
		TcpControlBits: r.TcpControlBits.Copy(),
		TcpFlags:       r.TcpFlags.Copy(),
		IngressIf:      r.IngressIf,
		EgressIf:       r.EgressIf,
		FlowDirection:  r.FlowDirection,
		StartMilli:     r.StartMilli,
		EndMilli:       r.EndMilli,
		AgentID:        r.AgentID,
	}
}

type Flags struct {
	NS  bool `json:"ns,string"`
	CWR bool `json:"cwr,string"`
	ECE bool `json:"ece,string"`
	URG bool `json:"urg,string"`
	ACK bool `json:"ack,string"`
	PSH bool `json:"psh,string"`
	RST bool `json:"rst,string"`
	SYN bool `json:"syn,string"`
	FIN bool `json:"fin,string"`
}

func (f Flags) Copy() Flags {
	return Flags{
		NS:  f.NS,
		CWR: f.CWR,
		ECE: f.ECE,
		URG: f.ECE,
		ACK: f.ACK,
		PSH: f.PSH,
		RST: f.RST,
		SYN: f.SYN,
		FIN: f.FIN,
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
