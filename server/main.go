package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	signal.Notify(sig, syscall.SIGTERM)

	conf := Parse()
	events := make(chan *Event)
	go LogServer(conf, events)
	go WebServer(conf, events)

	<-sig

	os.Exit(0)
}
