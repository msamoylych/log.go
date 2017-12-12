package main

import (
	"net"
	"encoding/gob"
	"log.go/api"
	"log"
	"github.com/hpcloud/tail"
	"io/ioutil"
	"os/user"
	"io"
	"os"
	"os/signal"
	"syscall"
	"encoding/json"
)

var tails = make([]*tail.Tail, 0, 10)

func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	signal.Notify(sig, syscall.SIGTERM)

	conf := config()
	events := make(chan api.LogEvent)
	stop := make(chan bool)
	sender(conf, events, stop)
	watcher(conf, events)

	<-sig

	for _, t := range tails {
		t.Stop()
		log.Println("Stop watching file:", t.Filename)
	}
	close(events)
	log.Println("Events channel closed")

	<-stop

	os.Exit(0)
}

func config() Config {
	usr, err := user.Current()
	if err != nil {
		log.Fatalln("Get current user error:", err)
	}

	confFile, err := ioutil.ReadFile(usr.HomeDir + "/.log.go/harvester.conf")
	if err != nil {
		log.Fatalln("Read config error:", err)
	}

	var config Config
	err = json.Unmarshal(confFile, &config)
	if err != nil {
		log.Fatalln("Parse config error:", err)
	}
	return config.Check()
}

func sender(conf Config, events <-chan api.LogEvent, stop chan<- bool) {
	address := conf.Server.Address()
	log.Println("Connect to", address)
	connection, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalln("Connection failed:", err)
	}

	encoder := gob.NewEncoder(connection)

	init := api.Init{NodeName: conf.NodeName}
	streams := make([]string, len(conf.LogStreams))
	for k := range conf.LogStreams {
		streams = append(streams, k)
	}
	init.Streams = streams
	err = encoder.Encode(init)
	if err != nil {
		log.Fatalln("Send init error:", err)
	}
	log.Println("Harvester inited")

	go func() {
		defer connection.Close()

		for event := range events {
			err = encoder.Encode(event)
			if err != nil {
				log.Panicln("Send event error:", err)
			}
		}

		stop <- true
	}()
}

func watcher(conf Config, events chan<- api.LogEvent) {
	var code byte
	for _, v := range conf.LogStreams {
		for _, file := range v {
			seekInfo := tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}
			t, err := tail.TailFile(file, tail.Config{Follow: true, ReOpen: true, Location: &seekInfo, Logger: tail.DiscardingLogger})
			if err != nil {
				log.Fatalln("Init watcher error:", err)
			}

			tails = append(tails, t)
			log.Println("Start watching file:", file)

			go func() {
				for line := range t.Lines {
					events <- api.LogEvent{Code: code, Msg: line.Text}
				}
			}()
		}
		code++
	}
}
