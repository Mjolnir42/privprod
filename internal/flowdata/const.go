/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package flowdata // import "github.com/mjolnir42/privprod/internal/flowdata"

const (
	flagFIN Bitmask = 1 << iota // No more data from sender
	flagSYN                     // Synchronize sequence numbers
	flagRST                     // Reset the connection
	flagPSH                     // Push Function
	flagACK                     // Acknowledgment field significant
	flagURG                     // Urgent Pointer field significant
	flagECE                     // ECN Echo
	flagCWR                     // Congestion Window Reduced
	flagNS                      // ECN Nonce Sum

	octetDeltaCount          = 1
	packetDeltaCount         = 2
	protocolIdentifier       = 4
	tcpControlBits           = 6
	sourceTransportPort      = 7
	sourceIPv4Address        = 8
	destinationTransportPort = 11
	destinationIpv4Address   = 12
	ingressInterface         = 10
	egressInterface          = 14
	sourceIPv6Address        = 27
	destinationIPv6Address   = 28
	ipVersion                = 60
	flowDirection            = 61 // 0x00: ingress, 0x01: egress
	exporterIPv4Address      = 130
	exporterIPv6Address      = 131
	exportingProcessID       = 144
	flowStartMilliseconds    = 152
	flowEndMilliseconds      = 153

	ProtocolUnknown = 0
	ProtocolICMP4   = 1
	ProtocolIGMP    = 2
	ProtocolIPv4    = 3
	ProtocolTCP     = 6
	ProtocolUDP     = 17
	ProtocolIPv6    = 41
	ProtocolGRE     = 47
	ProtocolESP     = 50
	ProtocolAH      = 51
	ProtocolICMP6   = 58
	ProtocolL2TP    = 115
	ProtocolSCTP    = 132
	ProtocolUDPLite = 136
	ProtocolMPLS    = 137

	ProtoNameUnknown = `unknown`
	ProtoNameICMP4   = `ICMP`
	ProtoNameIGMP    = `IGMP`
	ProtoNameIPv4    = `IPv4`
	ProtoNameTCP     = `TCP`
	ProtoNameUDP     = `UDP`
	ProtoNameIPv6    = `IPv6`
	ProtoNameGRE     = `GRE`
	ProtoNameESP     = `ESP`
	ProtoNameAH      = `AH`
	ProtoNameICMP6   = `IPv6-ICMP`
	ProtoNameL2TP    = `L2TP`
	ProtoNameSCTP    = `SCTP`
	ProtoNameUDPLite = `UDPLite`
	ProtoNameMPLS    = `MPLS-in-IP`
)

var ProtocolNameByID = map[uint8]string{
	ProtocolUnknown: ProtoNameUnknown,
	ProtocolICMP4:   ProtoNameICMP4,
	ProtocolIGMP:    ProtoNameIGMP,
	ProtocolIPv4:    ProtoNameIPv4,
	ProtocolTCP:     ProtoNameTCP,
	ProtocolUDP:     ProtoNameUDP,
	ProtocolIPv6:    ProtoNameIPv6,
	ProtocolGRE:     ProtoNameGRE,
	ProtocolESP:     ProtoNameESP,
	ProtocolAH:      ProtoNameAH,
	ProtocolICMP6:   ProtoNameICMP6,
	ProtocolL2TP:    ProtoNameL2TP,
	ProtocolSCTP:    ProtoNameSCTP,
	ProtocolUDPLite: ProtoNameUDPLite,
	ProtocolMPLS:    ProtoNameMPLS,
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
