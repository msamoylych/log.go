package main

import (
	"net"
	"encoding/gob"
	"log.go/api"
	"log"
	"github.com/hpcloud/tail"
	"fmt"
	"io/ioutil"
	"os/user"
	"github.com/flynn/json5"
	"io"
)

func main() {
	var conf = config()
	var channel = make(chan api.LogEvent)
	go client(conf, channel)
	go watcher(conf, channel)

	fmt.Scanln()
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
	err = json5.Unmarshal(confFile, &config)
	if err != nil {
		log.Fatalln("Parse config error:", err)
	}

	config.Check()

	return config
}

func client(conf Config, channel <-chan api.LogEvent) {
	var con, err = net.Dial("tcp", conf.Server.Address())
	if err != nil {
		log.Fatalln("Connection failed", err)
	}
	defer con.Close()

	var encoder = gob.NewEncoder(con)

	var init = api.Init{Code: "client"}
	var streams = make([]string, len(conf.LogStreams))
	var i byte = 0
	for k := range conf.LogStreams {
		streams[i] = k
		i++
	}
	init.Streams = streams
	err = encoder.Encode(&init)
	if err != nil {
		log.Panicln("Init error", err)
	}

	for {
		var event = <-channel
		err = encoder.Encode(&event)
		if err != nil {
			log.Println("Send error", err)
		}
	}
}

func watcher(conf Config, channel chan<- api.LogEvent) {
	var i byte = 0
	for _, v := range conf.LogStreams {
		for _, f := range v {
			go watch(i, f, channel)
		}
		i++
	}
}

func watch(code byte, filename string, channel chan<- api.LogEvent) {
	var seekInfo = tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}
	var logger = tail.DiscardingLogger
	t, err := tail.TailFile(filename, tail.Config{Follow: true, ReOpen: true, Location: &seekInfo, Logger: logger})
	if err != nil {
		log.Fatalln("Init watcher error:", err)
	}

	for line := range t.Lines {
		channel <- api.LogEvent{Code: code, Msg: line.Text}
	}
}
