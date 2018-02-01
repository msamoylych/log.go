package main

import (
	"log"
	"log.go/api"
	"io"
	"net"
	"encoding/gob"
	"sync/atomic"
	"sync"
)

type logServer struct {
	sync.Mutex
	config       *config
	harvestersCh chan<- *harvester
	eventsCh     chan<- *logEvent
	listener     net.Listener
	connections  map[net.Conn]struct{}
	closed       int32
}

func (server *logServer) add(connection net.Conn) {
	server.Lock()
	defer server.Unlock()

	server.connections[connection] = struct{}{}
}

func (server *logServer) del(connection net.Conn) {
	server.Lock()
	defer server.Unlock()

	delete(server.connections, connection)
}

func (server *logServer) isClosed() bool {
	return atomic.LoadInt32(&server.closed) == 1
}

func newLogServer(config *config, harvestersCh chan<- *harvester, eventsCh chan<- *logEvent) *logServer {
	return &logServer{
		config:       config,
		harvestersCh: harvestersCh,
		eventsCh:     eventsCh,
	}
}

func (server *logServer) listen() {
	addr := server.config.LogServer.address()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Init logServer error:", err)
	}
	log.Println("logServer listen at", addr)

	server.listener = ln
	server.connections = make(map[net.Conn]struct{})

	for {
		connection, err := ln.Accept()
		if err != nil {
			if server.isClosed() {
				break
			}
			log.Println("Listener accept error:", err)
			continue
		}

		go func() {
			decoder := gob.NewDecoder(connection)

			init := new(api.Init)
			err = decoder.Decode(init)
			if err != nil {
				log.Println("Init harvester error:", err)
				connection.Close()
				return
			}
			log.Println("Harvester '" + init.Node + "' connected")
			server.add(connection)
			server.harvestersCh <- &harvester{true, init.Node, init.Streams}

			event := new(api.LogEvent)
			for {
				err = decoder.Decode(event)
				if err != nil {
					if server.isClosed() {
						break
					}
					server.del(connection)
					server.harvestersCh <- &harvester{false, init.Node, init.Streams}
					connection.Close()
					if err == io.EOF {
						log.Println("Harvester '" + init.Node + "' disconnected")
					} else {
						log.Println("Harvester '"+init.Node+"' disconnected:", err)
					}
					break
				}

				server.eventsCh <- &logEvent{init.Node, event.Stream, event.Msg}
			}
		}()
	}
}

func (server *logServer) close() {
	atomic.StoreInt32(&server.closed, 1)
	err := server.listener.Close()
	if err != nil {
		log.Println("Close listener error:", err)
	}
	for connection := range server.connections {
		err = connection.Close()
		if err != nil {
			log.Println("Close connection error:", err)
		}
	}
}
