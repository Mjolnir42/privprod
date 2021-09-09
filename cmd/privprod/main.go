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
		go func() {
			h.Start()
			handlerLock.Done()
		}()
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

	// the main loop
runloop:
	for {
		select {
		case <-cancel:
			break runloop
		case err := <-server.Err():
			if err != nil {
				logrus.Errorln(`TCPServer:`, err)
			}
		case err := <-handlerDeath:
			if err != nil {
				logrus.Errorln(`Privacy:`, err)
			}
			break runloop
		}
	}

shutdown:
	// stop tcp server, read the error channel until all connections
	// have finished
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

	// close all handlers input channels, no new messages
	for i := range privacy.Handlers {
		close(privacy.Handlers[i].InputChannel())
	}

	// close all handlers shutdown channels, switch to drain mode
	for i := range privacy.Handlers {
		close(privacy.Handlers[i].ShutdownChannel())
	}

	// wait for handler shutdown
	handlerLock.Wait()

	// fetch final error messages
drainloop:
	for {
		select {
		case err := <-handlerDeath:
			if err != nil {
				logrus.Errorln(err)
			}
		case <-time.After(time.Millisecond * 10):
			break drainloop
		}
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
