package main

import (
	"log"
	"log.go/api"
	"io"
	"net"
	"encoding/gob"
	"sync/atomic"
)

type LogServer struct {
	config      *Config
	harvesters  *HarvesterMap
	events      chan<- *Event
	listener    net.Listener
	connections map[net.Conn]struct{}
	closed      int32
}

func NewLogServer(config *Config, harvesters *HarvesterMap, events chan<- *Event) *LogServer {
	return &LogServer{
		config:      config,
		harvesters:  harvesters,
		events:      events,
		connections: make(map[net.Conn]struct{}),
	}
}

func (server *LogServer) Listen() {
	addr := server.config.LogServer.Address()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Init LogServer error:", err)
	}
	log.Println("LogServer listen at", addr)

	server.listener = ln

	for {
		connection, err := ln.Accept()
		if err != nil {
			if server.IsClosed() {
				return
			}
			log.Println("Accept connection error:", err)
			continue
		}

		server.connections[connection] = struct{}{}

		go func() {
			defer connection.Close()

			decoder := gob.NewDecoder(connection)

			var init api.Init
			err := decoder.Decode(&init)
			if err != nil {
				log.Println("Init harvester error:", err)
				return
			}

			server.harvesters.add(init.Node, init.Streams)
			log.Println("Harvester '" + init.Node + "' connected")

			var event api.LogEvent
			for {
				err = decoder.Decode(&event)
				switch {
				case server.IsClosed():
					return
				case err != nil:
					server.harvesters.del(init.Node)
					delete(server.connections, connection)
					if err == io.EOF {
						log.Println("Harvester '" + init.Node + "' disconnected")
					} else {
						log.Println("Harvester '"+init.Node+"' disconnected:", err)
					}
					return
				}

				server.events <- &Event{Node: init.Node, Stream: event.Stream, Message: event.Msg}
			}
		}()
	}
}

func (server *LogServer) Close() {
	atomic.StoreInt32(&server.closed, 1)
	err := server.listener.Close()
	if err != nil {
		log.Println("Close listener error:", err)
	}
	for con := range server.connections {
		err = con.Close()
		if err != nil {
			log.Println("Close connection error:", err)
		}
	}
}

func (server *LogServer) IsClosed() bool {
	return atomic.LoadInt32(&server.closed) == 1
}
