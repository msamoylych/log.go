package main

import (
	"log.go/api"
	"log"
	"net"
	"encoding/gob"
)

type Sender struct {
	done chan bool
}

func NewSender(config *Config, events <-chan api.LogEvent) *Sender {
	sender := &Sender{done: make(chan bool)}
	address := config.Server.Address()
	connection, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalln("Connection failed:", err)
	}
	log.Println("Connected to server:", address)

	go func() {
		defer connection.Close()

		encoder := gob.NewEncoder(connection)

		init := api.Init{Node: config.NodeName, Streams: config.Streams()}
		err = encoder.Encode(&init)
		if err != nil {
			log.Panicln("Send init error:", err)
		}

		for event := range events {
			err = encoder.Encode(event)
			if err != nil {
				log.Panicln("Send event error:", err)
			}
		}

		sender.done <- true
	}()

	return sender
}
