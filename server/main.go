package main

import (
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
	events := make(chan *Event, bufSize)
	harvesters := NewHarvesters()

	webServer := NewWebServer(config, harvesters, events)
	go webServer.Listen()

	logServer := NewLogServer(config, harvesters, events)
	go logServer.Listen()

	<-sig

	logServer.Close()
	close(events)

	select {
	case <-webServer.done:
		webServer.Close()
		os.Exit(0)
	case <-time.After(1 * time.Second):
		log.Fatalln("Stop server timeout")
	}
}
