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

func main() {
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
			logrus.Infof("Running handler privacy.Protector %d\n", num)
			h.Start()
			handlerLock.Done()
			logrus.Infof("Handler finished: privacy.Protector %d\n", num)
		}(i)
	}

	addr := os.Getenv(`PRIVACY_LISTEN_ADDRESS`)
	switch addr {
	case ``:
		addr = `localhost:4150`
	default:
	}

	server, err := NewTCPServer(addr)
	if err != nil {
		logrus.Errorln(err)
		goto shutdown
	}
	logrus.Infof("Started TCP server at %s", addr)

	// the main loop
	logrus.Infoln("Running main event loop")
runloop:
	for {
		select {
		case <-cancel:
			logrus.Infoln("Received interrupt request, exiting")
			break runloop
		case err := <-server.Err():
			if err != nil {
				logrus.Errorln(`TCPServer:`, err)
			}
		case err := <-handlerDeath:
			if err != nil {
				logrus.Errorln(`Privacy:`, err)
			}
			logrus.Infoln("Handler died, exiting")
			break runloop
		}
	}

shutdown:
	// stop tcp server, read the error channel until all connections
	// have finished
	logrus.Infoln("Shutting down TCP server, waiting for clients....")
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
	logrus.Infoln("All TCP server connections closed")

	// close all handlers input channels, no new messages
	for i := range privacy.Handlers {
		close(privacy.Handlers[i].InputChannel())
	}

	// close all handlers shutdown channels, switch to drain mode
	for i := range privacy.Handlers {
		close(privacy.Handlers[i].ShutdownChannel())
	}

	// fetch final error messages
	logrus.Infoln("Draining final handler error messages")
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
	logrus.Infoln("Waiting for handler shutdowns.")
	handlerLock.Wait()

}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
