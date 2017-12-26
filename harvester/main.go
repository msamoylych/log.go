package main

import (
	"log.go/api"
	"os"
	"os/signal"
	"syscall"
	"time"
	"log"
)

const bufSize = 1000

func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	signal.Notify(sig, syscall.SIGTERM)

	config := Parse()
	events := make(chan api.LogEvent, bufSize)
	sender := NewSender(config, events)
	watcher := NewWatcher(config, events)

	<-sig

	watcher.Stop()
	close(events)

	select {
	case <-sender.done:
		os.Exit(0)
	case <-time.After(5 * time.Second):
		log.Fatalln("Stop sender timeout")
	}
}
