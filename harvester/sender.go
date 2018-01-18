package main

import (
	"log.go/api"
	"log"
	"net"
	"encoding/gob"
)

type Sender struct {
	config *Config
	events <-chan api.LogEvent
	done   chan struct{}
}

func NewSender(config *Config, events <-chan api.LogEvent) *Sender {
	return &Sender{
		config: config,
		events: events,
		done:   make(chan struct{}),
	}
}

func (sender *Sender) Connect() {
	address := sender.config.Server.Address()
	connection, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalln("Connection failed:", err)
	}
	log.Println("Connected to server:", address)

	go func() {
		defer connection.Close()

		encoder := gob.NewEncoder(connection)

		init := api.Init{Node: sender.config.NodeName, Streams: sender.config.Streams()}
		err = encoder.Encode(&init)
		if err != nil {
			log.Fatalln("Send init error:", err)
		}

		for event := range sender.events {
			err = encoder.Encode(event)
			if err != nil {
				log.Fatalln("Send event error:", err)
			}
		}

		sender.done <- struct{}{}
	}()
}
