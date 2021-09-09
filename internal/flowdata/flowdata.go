/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

// Package flowdata contains conversions for processing IPFIX flow
// messages as emitted by vflow.
package flowdata // import "github.com/mjolnir42/privprod/internal/flowdata"

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func FormatIP(addr string) string {
	raw := net.ParseIP(strings.Trim(addr, `"`)).To16()
	return fmt.Sprintf(
		"%x:%x:%x:%x:%x:%x:%x:%x",
		[]byte(raw)[0:2],
		[]byte(raw)[2:4],
		[]byte(raw)[4:6],
		[]byte(raw)[6:8],
		[]byte(raw)[8:10],
		[]byte(raw)[10:12],
		[]byte(raw)[12:14],
		[]byte(raw)[14:16],
	)
}

func unix2time(tstp int64) time.Time {
	return time.Unix(tstp/1000, (tstp%1000)*1000000)
}

func parseUint8(b []byte) uint8 {
	i64, _ := strconv.ParseUint(string(b), 10, 8)
	return uint8(i64)
}

func parseUint16(b []byte) uint16 {
	i64, _ := strconv.ParseUint(string(b), 10, 16)
	return uint16(i64)
}

func parseUint32(b []byte) uint32 {
	i64, _ := strconv.ParseUint(string(b), 10, 32)
	return uint32(i64)
}

func parseUint64(b []byte) uint64 {
	i64, _ := strconv.ParseUint(string(b), 10, 64)
	return uint64(i64)
}

func parseInt64(b []byte) int64 {
	i64, _ := strconv.ParseInt(string(b), 10, 64)
	return int64(i64)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
