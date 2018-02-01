package main

import (
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
	harvestersCh := make(chan *harvester)
	eventsCh := make(chan *logEvent, bufSize)

	webServer := newWebServer(config, harvestersCh, eventsCh)
	go webServer.listen()

	logServer := newLogServer(config, harvestersCh, eventsCh)
	go logServer.listen()

	<-sig

	logServer.close()
	close(harvestersCh)
	close(eventsCh)

	select {
	case <-webServer.doneCh:
		webServer.close()
		os.Exit(0)
	case <-time.After(1 * time.Second):
		log.Fatalln("Stop server timeout")
	}
}
