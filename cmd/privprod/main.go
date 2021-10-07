/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package main

import (
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/mjolnir42/erebos"
	"github.com/mjolnir42/privprod/internal/privacy"
	"github.com/sirupsen/logrus"
)

var (
	// this version string is set by the build script
	privprodVersion string
)

func main() {
	logrus.Infof("Starting privprod version: %s\n", privprodVersion)

	handlerDeath := make(chan error)
	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, os.Interrupt, syscall.SIGTERM)

	// start application handlers
	handlerLock := sync.WaitGroup{}
	for i := 0; i < runtime.NumCPU(); i++ {
		handlerLock.Add(1)
		h := privacy.Protector{
			Num:      i,
			Input:    make(chan *erebos.Transport, 16),
			Shutdown: make(chan struct{}),
			Death:    handlerDeath,
		}
		privacy.Handlers[i] = &h
		go func(num int) {
			logrus.Infof("Main: running handler privacy.Protector %d\n", num)
			h.Start()
			handlerLock.Done()
			logrus.Infof("Main: handler finished: privacy.Protector %d\n", num)
		}(i)
	}

	addr := os.Getenv(`PRIVACY_LISTEN_ADDRESS`)
	switch addr {
	case ``:
		addr = `localhost:4150`
	default:
	}
	logrus.Infof("Main: configured tcpserver to listen on: %s\n", addr)

	server, err := NewTCPServer(addr)
	if err != nil {
		logrus.Errorln(err)
		goto shutdown
	}
	logrus.Infof("Main: started TCP server at %s", addr)

	// the main loop
	logrus.Infoln("Main: running main event loop")
runloop:
	for {
		select {
		case <-cancel:
			logrus.Infoln("Main: received interrupt request, exiting")
			break runloop
		case err := <-server.Err():
			if err != nil {
				logrus.Errorln(`TCPServer:`, err)
			}
		case err := <-handlerDeath:
			if err != nil {
				logrus.Errorln(`Privacy:`, err)
			}
			logrus.Infoln("Main: handler died, forced exiting")
			break runloop
		}
	}

shutdown:
	// stop tcp server, read the error channel until all connections
	// have finished
	logrus.Infoln("Main: shutting down TCP server, waiting for clients....")
	ch := server.Stop()
serverGrace:
	for {
		select {
		case err := <-ch:
			if err != nil {
				logrus.Errorln(`TCPServer:`, err)
				continue serverGrace
			}
			break serverGrace
		}
	}
	logrus.Infoln("Main: all TCP server connections closed")

	// close all handlers input channels, no new messages
	for i := range privacy.Handlers {
		close(privacy.Handlers[i].InputChannel())
	}

	// close all handlers shutdown channels, switch to drain mode
	for i := range privacy.Handlers {
		close(privacy.Handlers[i].ShutdownChannel())
	}

	// fetch final error messages
	logrus.Infoln("Main: draining final handler error messages")
drainloop:
	for {
		select {
		case err := <-handlerDeath:
			if err != nil {
				logrus.Errorln(err)
			}
		case <-time.After(time.Millisecond * 100):
			break drainloop
		}
	}

	// wait for handler shutdown
	logrus.Infoln("Main: waiting for handler shutdowns.")
	handlerLock.Wait()

}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
