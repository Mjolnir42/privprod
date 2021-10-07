/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package main

import (
	"bufio"
	"errors"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/mjolnir42/erebos"
	"github.com/mjolnir42/privprod/internal/privacy"
	"github.com/sirupsen/logrus"
)

type TCPServer struct {
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
	err      chan error
}

func NewTCPServer(addr string) (*TCPServer, error) {
	var err error
	s := &TCPServer{
		quit: make(chan interface{}),
		err:  make(chan error),
	}
	if s.listener, err = net.Listen(`tcp`, addr); err != nil {
		return nil, err
	}
	s.wg.Add(1)
	go s.serve()
	return s, nil
}

func (s *TCPServer) Err() chan error {
	return s.err
}

func (s *TCPServer) serve() {
	defer s.wg.Done()
	logrus.Infoln(`TCPserver: start serving clients`)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				logrus.Infoln(`TCPserver: graceful stop of main serve loop`)
				return
			default:
				s.err <- err
			}
		} else {
			s.wg.Add(1)
			go func() {
				remote := conn.RemoteAddr().String()
				logrus.Infof("TCPserver: accepted connection from: %s\n",
					remote,
				)
				s.handleConnection(conn)
				logrus.Infof("TCPserver: finished connection from: %s\n",
					remote,
				)
				s.wg.Done()
			}()
		}
	}
}

func (s *TCPServer) Stop() chan error {
	go func(e chan error) {
		close(s.quit)
		s.listener.Close()
		s.wg.Wait()
		close(e)
	}(s.err)
	return s.err
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

ReadLoop:
	for {
		select {
		case <-s.quit:
			break ReadLoop
		default:
			conn.SetDeadline(time.Now().Add(400 * time.Millisecond))

			scanner := bufio.NewScanner(conn)
			scanner.Split(bufio.ScanLines)

			for scanner.Scan() {
				go func(data []byte) {
					// explicit copy to avoid this panic
					// panic: JSON decoder out of sync - data changing underfoot?
					datacopy := make([]byte, len(data))
					copy(datacopy, data)
					privacy.Dispatch(erebos.Transport{Value: datacopy})
				}(scanner.Bytes())

				// refresh deadline after a line has been read and s.quit has not
				// been closed yet
				select {
				case <-s.quit:
					logrus.Infof("TCPserver: forcing close on connection from: %s\n",
						conn.RemoteAddr().String(),
					)
				default:
					conn.SetDeadline(time.Now().Add(400 * time.Millisecond))
				}
			}

			if err := scanner.Err(); err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					// conn.Deadline triggered
					continue ReadLoop
				} else if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					// net package triggered timeout
					continue ReadLoop
				} else if err != io.EOF {
					s.err <- err
				}
			}
			// scanner finished without error or timeout -> received EOF and
			// connection is closed
			break ReadLoop
		}
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
