/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package privacy // import "github.com/mjolnir42/privprod/internal/privacy"

//
var (
	dataPad, pseudoKey   []byte
	employeePrivNetworks map[string]*net.IPNet
	employeePubNetworks  map[string]*net.IPNet
	companyPubNetworks   map[string]*net.IPNet
	reservedPrivNetworks map[string]*net.IPNet
	discardNetworks      map[string]*net.IPNet
)

//
const (
	keyLenBytes  = 32
	saltLenBytes = 16
)

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
