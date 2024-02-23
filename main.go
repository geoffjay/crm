package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func main() {
	initLogging()

	// app := app{}
	// app.init()

	fields := log.Fields{"service": "app", "context": "main"}

	ctx, cancelFunc := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	wg.Add(1)
	// go app.run(ctx, wg)

	log.WithFields(fields).Debug("starting")

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan

	log.WithFields(fields).Debug("terminated")

	cancelFunc()
	wg.Wait()

	log.WithFields(fields).Debug("exiting")
}
