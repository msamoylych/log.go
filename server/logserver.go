package main

import (
	"log"
	"log.go/api"
	"io"
	"net"
	"encoding/gob"
	"sync"
)

type HarvesterMap struct {
	sync.Mutex
	Harvesters map[string][]string
}

func (h HarvesterMap) add(node string, streams []string) {
	h.Lock()
	defer h.Unlock()

	h.Harvesters[node] = streams
}

func (h HarvesterMap) del(node string) {
	h.Lock()
	defer h.Unlock()

	delete(h.Harvesters, node)
}

func (h HarvesterMap) Controls() *Controls {
	h.Lock()
	defer h.Unlock()

	controls := Controls{Streams: make(map[string][]string), Nodes: make(map[string][]string)}
	for node, streams := range h.Harvesters {
		controls.Nodes[node] = streams
		for _, stream := range streams {
			if _, ok := controls.Streams[stream]; !ok {
				controls.Streams[stream] = make([]string, 0, 10)
			}
			controls.Streams[stream] = append(controls.Streams[stream], node)
		}
	}
	return &controls
}

var Harvesters = HarvesterMap{Harvesters: make(map[string][]string)}

func LogServer(conf *Config, events chan<- *Event) {
	addr := conf.LogServer.Address()
	log.Println("LogServer listen at", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Init LogServer error:", err)
	}
	defer ln.Close()

	for {
		connection, err := ln.Accept()
		if err != nil {
			log.Println("Accept connection error:", err)
			continue
		}

		go func() {
			defer connection.Close()

			decoder := gob.NewDecoder(connection)

			var init api.Init
			err := decoder.Decode(&init)
			if err != nil {
				log.Println("Init harvester error:", err)
				return
			}

			Harvesters.add(init.Node, init.Streams)
			log.Println("Harvester '" + init.Node + "' connected")

			var event api.LogEvent
			for {
				err = decoder.Decode(&event)
				switch {
				case err == io.EOF:
					Harvesters.del(init.Node)
					log.Println("Harvester '" + init.Node + "' disconnected")
					return
				case err != nil:
					Harvesters.del(init.Node)
					log.Println("Read event error:", err)
					return
				}
				events <- &Event{Node: init.Node, Stream: event.Stream, Message: event.Msg}
			}
		}()
	}
}
