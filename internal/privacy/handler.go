/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package privacy // import "github.com/mjolnir42/privprod/internal/privacy"

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mjolnir42/erebos"
	"github.com/mjolnir42/flowdata"
	"github.com/sirupsen/logrus"
)

// Handlers must be set before Protector.Start is called for
// the first time
var Handlers map[int]erebos.Handler

// initialize Handlers map
func init() {
	Handlers = make(map[int]erebos.Handler)

	// BUG: datapad should be read from Zookeeper
	dataPad, _ = hex.DecodeString(os.Getenv(`PRIVACY_DATAPAD`))
	// BUG: pseudokey does not rotate
	pseudoKey, _ = hex.DecodeString(os.Getenv(`PRIVACY_DAILY_KEY`))

	buildNetworkMaps()
}

// Dispatch implements erebos.Dispatcher
func Dispatch(msg erebos.Transport) error {
	// send all messages from the same Host to the same handler
	decoded := flowdata.Message{}
	if err := json.Unmarshal(msg.Value, decoded); err != nil {
		logrus.Errorln(err)
		return err
	}

	ip := net.ParseIP(decoded.AgentID)

	x := big.NewInt(23)
	x = x.SetBytes(ip)
	numCPU := big.NewInt(int64(runtime.NumCPU()))
	handler := big.NewInt(1)
	handler = handler.Mod(x, numCPU)

	Handlers[int(handler.Int64())].InputChannel() <- &msg
	return nil
}

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

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
