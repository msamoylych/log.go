package main

import (
	"log.go/api"
	"log"
	"net"
	"encoding/gob"
)

type sender struct {
	config   *config
	eventsCh <-chan *api.LogEvent
	doneCh   chan struct{}
}

func newSender(config *config, eventsCh <-chan *api.LogEvent) *sender {
	return &sender{
		config:   config,
		eventsCh: eventsCh,
		doneCh:   make(chan struct{}),
	}
}

func (sender *sender) connect() {
	address := sender.config.Server.address()
	connection, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalln("Connection failed:", err)
	}
	log.Println("Connected to server:", address)

	go func() {
		encoder := gob.NewEncoder(connection)

		init := &api.Init{Node: sender.config.NodeName, Streams: sender.config.Streams()}
		err = encoder.Encode(init)
		if err != nil {
			log.Fatalln("Send init error:", err)
		}

		for event := range sender.eventsCh {
			err = encoder.Encode(event)
			if err != nil {
				log.Fatalln("Send event error:", err)
			}
		}

		connection.Close()
		sender.doneCh <- struct{}{}
	}()
}
