package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/mjolnir42/erebos"
	"github.com/mjolnir42/privprod/internal/privacy"
	nsq "github.com/nsqio/go-nsq"
	"github.com/nsqio/nsq/nsqd"
	"github.com/sirupsen/logrus"
)

func main() {
	stopNSQD := make(chan struct{})
	stopNSQConsumer := make(chan struct{})
	handlerDeath := make(chan error)
	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, os.Interrupt, syscall.SIGTERM)

	// start application handlers
	for i := 0; i < runtime.NumCPU(); i++ {
		h := privacy.Protector{
			Num:      i,
			Input:    make(chan *erebos.Transport, 4),
			Shutdown: make(chan struct{}),
			Death:    handlerDeath,
		}
		privacy.Handlers[i] = &h
		go func() {
			h.Start()
		}()
	}

	// start queue daemon
	go runNSQD(stopNSQD)
	go runNSQConsumer(stopNSQConsumer)

	// the main loop
runloop:
	for {
		select {
		case <-cancel:
			break runloop
		case err := <-handlerDeath:
			if err != nil {
				logrus.Errorln(err)
			}
			break runloop
		}
	}

	// close all handlers
	close(stopNSQConsumer)
	close(stopNSQD)
	for i := range privacy.Handlers {
		close(privacy.Handlers[i].ShutdownChannel())
		close(privacy.Handlers[i].InputChannel())
	}

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

func runNSQConsumer(done chan struct{}) {
	cfg := nsq.NewConfig()
	cfg.Snappy = false
	// topic: data, channel: main
	c, err := nsq.NewConsumer(`data`, `main`, cfg)
	if err != nil {
		logrus.Fatalln(err)
	}

	// receive messages and hand off to privacy library
	c.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		logrus.Infoln(string(m.Body))
		go privacy.Dispatch(erebos.Transport{
			Value: m.Body,
		})
		//
		return nil
	}))

	// connect consumer to queue
	c.ConnectToNSQD(`localhost:` + os.Getenv(`NSQD_TCP_PORT`))
	<-done

	c.Stop()
	select {
	case <-c.StopChan:
		// unblock once clean consumer stop is complete
		logrus.Infoln(`Clean consumer shutdown complete`)
	case <-time.After(time.Second * 8):
		logrus.Errorln(`Forced dirty consumer shutdown after 8s timeout`)
	}
}

func runNSQD(done chan struct{}) {
	logger := logrus.New()
	logger.Out = ioutil.Discard
	opts := nsqd.NewOptions()
	opts.TCPAddress = `localhost:` + os.Getenv(`NSQD_TCP_PORT`)
	opts.HTTPAddress = `localhost:` + os.Getenv(`NSQD_HTTP_PORT`)
	opts.HTTPSAddress = `localhost:` + os.Getenv(`NSQD_HTTPS_PORT`)
	opts.DataPath = os.Getenv(`NSQD_DATA_PATH`)
	opts.MemQueueSize = 8192 // number of messages
	opts.MaxMsgSize = 131072 // bytes per message

	nsqd, err := nsqd.New(opts)
	if err != nil {
		logrus.Fatalln(err)
	}
	nsqd.Main()

	<-done
	nsqd.Exit()
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
