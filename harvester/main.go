package main

import (
	"log.go/api"
	"os"
	"os/signal"
	"time"
	"log"
)

const bufSize = 100

func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)

	config := parse()
	eventsCh := make(chan *api.LogEvent, bufSize)

	sender := newSender(config, eventsCh)
	sender.connect()

	watcher := newWatcher(config, eventsCh)
	watcher.start()

	<-sig

	watcher.stop()
	close(eventsCh)

	select {
	case <-sender.doneCh:
		os.Exit(0)
	case <-time.After(1 * time.Second):
		log.Fatalln("Stop timeout")
	}
}
