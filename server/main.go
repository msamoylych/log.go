package main

import (
	"net"
	"encoding/gob"
	"log.go/api"
	"io"
	"log"
	"strconv"
)

func main() {
	server()
}

func server() {
	ln, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatalln("Init listener error:", err)
		return
	}
	defer ln.Close()

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println("Accept connection error:", err)
			continue
		}

		go handle(c)
	}
}

func handle(c net.Conn) {
	defer c.Close()

	var decoder = gob.NewDecoder(c)

	var init api.Init
	err := decoder.Decode(&init)

	if err != nil {
		log.Println("Init harvester error:", err)
		return
	}

	log.Println("Harvester '" + init.NodeName + "' connected")

	var event api.LogEvent
	for {
		err = decoder.Decode(&event)
		switch {
		case err == io.EOF:
			log.Println("Harvester '" + init.NodeName + "' disconnected")
			return
		case err != nil:
			log.Println("Read event error:", err)
			return
		}

		log.Println(strconv.FormatUint(uint64(event.Code), 10) + ": " + event.Msg)
	}
}
