/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package privacy // import "github.com/mjolnir42/privprod/internal/privacy"

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net"
	"os"
	"runtime"

	"github.com/mjolnir42/erebos"
	"github.com/mjolnir42/flowdata"
	"github.com/sirupsen/logrus"
)

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
	decoded := &flowdata.Message{}
	if err := json.Unmarshal(msg.Value, decoded); err != nil {
		logrus.Errorln(`privacy.Dispatch(): ` + err.Error())
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

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
