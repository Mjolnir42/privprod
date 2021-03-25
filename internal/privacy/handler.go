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

var dataPad, pseudoKey []byte

// initialize Handlers map
func init() {
	Handlers = make(map[int]erebos.Handler)

	// BUG: datapad should be read from Zookeeper
	dataPad, _ = hex.DecodeString(os.Getenv(`PRIVACY_DATAPAD`))
	// BUG: pseudokey does not rotate
	pseudoKey, _ = hex.DecodeString(os.Getenv(`PRIVACY_DAILY_KEY`))
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

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
