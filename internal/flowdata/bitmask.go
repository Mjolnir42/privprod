/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package flowdata // import "github.com/mjolnir42/privprod/internal/flowdata"

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

type Bitmask uint16

func ParseBitmask(s string) Bitmask {
	i, err := strconv.ParseUint(
		strings.Trim(s, `"`),
		0,
		16,
	)
	if err != nil {
		return Bitmask(0)
	}
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(i))
	return Bitmask(binary.BigEndian.Uint16(b[:]))
}

func (mask Bitmask) Set(flag Bitmask) {
	mask = mask | flag // OR
}

func (mask Bitmask) Clear(flag Bitmask) {
	mask = mask &^ flag // AND NOT
}

func (mask Bitmask) Toggle(flag Bitmask) {
	mask = mask ^ flag // XOR
}

func (mask Bitmask) Has(flag Bitmask) bool {
	return mask&flag != 0 // AND
}

func (mask Bitmask) Copy() Bitmask {
	var i Bitmask
	i.Set(mask)
	return i
}

func (mask Bitmask) String() string {
	return fmt.Sprintf("%#04x", string(mask))
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
