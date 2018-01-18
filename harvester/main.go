package main

import (
	"log.go/api"
	"os"
	"os/signal"
	"time"
	"log"
)

const bufSize = 1000

func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)

	config := Parse()
	events := make(chan api.LogEvent, bufSize)

	sender := NewSender(config, events)
	sender.Connect()

	watcher := NewWatcher(config, events)
	watcher.Start()

	<-sig

	watcher.Stop()
	close(events)

	select {
	case <-sender.done:
		os.Exit(0)
	case <-time.After(1 * time.Second):
		log.Fatalln("Stop timeout")
	}
}
