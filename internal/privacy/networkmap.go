/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package privacy // import "github.com/mjolnir42/privprod/internal/privacy"

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func buildNetworkMaps() {
	cfgPath := os.Getenv(`PRIVACY_NETWORKFILE_PATH`)

	employeePrivNetworks = map[string]*net.IPNet{}
	employeePubNetworks = map[string]*net.IPNet{}
	companyPubNetworks = map[string]*net.IPNet{}
	reservedPrivNetworks = map[string]*net.IPNet{}
	discardNetworks = map[string]*net.IPNet{}

	for _, fname := range []string{
		`company-public.txt`,
		`discard.txt`,
		`employee-private.txt`,
		`employee-public.txt`,
		`reserved.txt`,
	} {
		file, err := os.Open(filepath.Join(cfgPath, fname))
		if err != nil {
			logrus.Fatalln(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, `#`) {
				// ignore comment
				continue
			}
			line = strings.TrimSpace(line)

			var nmap *map[string]*net.IPNet
			var err error

			switch fname {
			case `company-public.txt`:
				nmap = &companyPubNetworks
			case `discard.txt`:
				nmap = &discardNetworks
			case `employee-private.txt`:
				nmap = &employeePrivNetworks
			case `employee-public.txt`:
				nmap = &employeePubNetworks
			case `reserved.txt`:
				nmap = &reservedPrivNetworks
			}
			if _, (*nmap)[line], err = net.ParseCIDR(line); err != nil {
				logrus.Fatalln(err)
			}
		}

		if err := scanner.Err(); err != nil {
			logrus.Fatalln(err)
		}
	}
}

func discard(ip net.IP) bool {
	for i := range discardNetworks {
		if discardNetworks[i].Contains(ip) {
			return true
		}
	}
	return false
}

func isPrivate(ip net.IP) bool {
	for i := range reservedPrivNetworks {
		if reservedPrivNetworks[i].Contains(ip) {
			return true
		}
	}
	return false
}

func isEmployeePriv(ip net.IP) bool {
	for i := range employeePrivNetworks {
		if employeePrivNetworks[i].Contains(ip) {
			return true
		}
	}
	return false
}

func isEmployeePub(ip net.IP) bool {
	for i := range employeePubNetworks {
		if employeePubNetworks[i].Contains(ip) {
			return true
		}
	}
	return false
}

func isPublic(ip net.IP) bool {
	if !isPrivate(ip) && !isCompany(ip) {
		return true
	}
	return false
}

func isCompany(ip net.IP) bool {
	for i := range companyPubNetworks {
		if companyPubNetworks[i].Contains(ip) {
			return true
		}
	}
	return false
}

func fmtEmployeePriv(b []byte) string {
	return fmt.Sprintf(
		"0100:a000:%x:%x:%x:%x:%x:%x",
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:12],
		b[12:14],
		b[14:16],
	)
}

func fmtEmployeePub(b []byte) string {
	return fmt.Sprintf(
		"0100:b000:%x:%x:%x:%x:%x:%x",
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:12],
		b[12:14],
		b[14:16],
	)
}

func fmtCustomer(b []byte) string {
	return fmt.Sprintf(
		"0100:c000:%x:%x:%x:%x:%x:%x",
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:12],
		b[12:14],
		b[14:16],
	)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
