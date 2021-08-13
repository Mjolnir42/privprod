/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package flowdata // import "github.com/mjolnir42/privprod/internal/flowdata"

import "encoding/json"

type Message struct {
	AgentID  string `json:"AgentID"`
	Header   Header `json:"Header"`
	DataSets []Data `json:"DataSets"`
}

type Header struct {
	Version    int `json:"Version"`
	Length     int `json:"Length"`
	ExportTime int `json:"ExportTime"`
	SequenceNo int `json:"SequenceNo"`
	DomainID   int `json:"DomainID"`
}

type Data []kvpair

type kvpair struct {
	Key   float64         `json:"I"`
	Value json.RawMessage `json:"V"`
}

func (m *Message) Convert() <-chan Record {
	ret := make(chan Record)
	go func() {
		for i := range m.DataSets {
			res := Record{
				AgentID:        m.AgentID,
				TcpControlBits: Bitmask(0),
				TcpFlags:       Flags{},
			}
			for _, pair := range m.DataSets[i] {
				switch pair.Key {
				case octetDeltaCount:
					res.OctetCount = parseUint64(pair.Value)
				case packetDeltaCount:
					res.PacketCount = parseUint64(pair.Value)
				case protocolIdentifier:
					res.ProtocolID = parseUint8(pair.Value)
					if _, ok := ProtocolNameByID[res.ProtocolID]; ok {
						res.Protocol = ProtocolNameByID[res.ProtocolID]
					} else {
						res.Protocol = ProtoNameUnknown
					}
				case tcpControlBits:
					res.TcpControlBits = ParseBitmask(string(pair.Value))
					res.TcpFlags.FIN = res.TcpControlBits.Has(flagFIN)
					res.TcpFlags.SYN = res.TcpControlBits.Has(flagSYN)
					res.TcpFlags.RST = res.TcpControlBits.Has(flagRST)
					res.TcpFlags.PSH = res.TcpControlBits.Has(flagPSH)
					res.TcpFlags.ACK = res.TcpControlBits.Has(flagACK)
					res.TcpFlags.URG = res.TcpControlBits.Has(flagURG)
					res.TcpFlags.ECE = res.TcpControlBits.Has(flagECE)
					res.TcpFlags.CWR = res.TcpControlBits.Has(flagCWR)
					res.TcpFlags.NS = res.TcpControlBits.Has(flagNS)
				case sourceTransportPort:
					res.SrcPort = parseUint16(pair.Value)
				case sourceIPv4Address:
					res.SrcAddress = FormatIP(string(pair.Value))
				case destinationTransportPort:
					res.DstPort = parseUint16(pair.Value)
				case destinationIpv4Address:
					res.DstAddress = FormatIP(string(pair.Value))
				case ingressInterface:
					res.IngressIf = parseUint32(pair.Value)
				case egressInterface:
					res.EgressIf = parseUint32(pair.Value)
				case sourceIPv6Address:
					res.SrcAddress = FormatIP(string(pair.Value))
				case destinationIPv6Address:
					res.DstAddress = FormatIP(string(pair.Value))
				case ipVersion:
					res.IPVersion = parseUint8(pair.Value)
				case flowDirection:
					res.FlowDirection = parseUint8(pair.Value)
				case exporterIPv4Address:
					res.ExpIPv4Addr = FormatIP(string(pair.Value))
				case exporterIPv6Address:
					res.ExpIPv6Addr = FormatIP(string(pair.Value))
				case exportingProcessID:
					res.ExpPID = parseUint32(pair.Value)
				case flowStartMilliseconds:
					res.StartMilli = unix2time(parseInt64(pair.Value))
				case flowEndMilliseconds:
					res.EndMilli = unix2time(parseInt64(pair.Value))
				default:
				}
			}
			ret <- res
		}
		close(ret)
	}()
	return ret
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
