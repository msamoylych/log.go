package main

import (
	"log"
	"net/http"
	"golang.org/x/net/websocket"
	"sync"
	"net"
	"encoding/json"
	"mime"
	"path/filepath"
	"sync/atomic"
	"io"
)

type clients struct {
	sync.RWMutex
	connections map[*websocket.Conn]struct{}
}

func (c clients) add(connection *websocket.Conn) {
	c.Lock()
	defer c.Unlock()

	c.connections[connection] = struct{}{}
}

func (c clients) del(connection *websocket.Conn) {
	c.Lock()
	defer c.Unlock()

	delete(c.connections, connection)
}

func (c clients) send(event *Event) {
	c.RLock()
	defer c.RUnlock()

	for connection := range c.connections {
		websocket.JSON.Send(connection, event)
	}
}

func (c clients) close() {
	c.Lock()
	defer c.Unlock()

	for connection := range c.connections {
		connection.Close()
	}
}

type WebServer struct {
	*http.Server
	config     *Config
	harvesters *HarvesterMap
	events     <-chan *Event
	done       chan struct{}
	clients    *clients
	closed     int32
}

func NewWebServer(config *Config, harvesters *HarvesterMap, events <-chan *Event) (*WebServer) {
	return &WebServer{
		Server:     new(http.Server),
		config:     config,
		harvesters: harvesters,
		events:     events,
		done:       make(chan struct{}),
		clients:    &clients{connections: make(map[*websocket.Conn]struct{})},
	}
}

func (server *WebServer) Listen() {
	addr := server.config.WebServer.Address()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Init WebServer error:", err)
	}
	log.Println("WebServer listen at", addr)

	go func() {
		for event := range server.events {
			server.clients.send(event)
		}
		server.done <- struct{}{}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.RequestURI[1:]
		if name == "" {
			name = "index.html"
		}
		data, err := Asset(name)
		if err != nil {
			http.NotFound(w, r)
		} else {
			ctype := mime.TypeByExtension(filepath.Ext(name))
			w.Header().Set("Content-Type", ctype)
			w.Write(data)
		}
	})

	http.HandleFunc("/controls", func(w http.ResponseWriter, r *http.Request) {
		resp, err := json.Marshal(server.harvesters.Controls())
		if err != nil {
			log.Println("Marshal controls error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)
		}
	})

	http.Handle("/ws", websocket.Handler(func(connection *websocket.Conn) {
		defer connection.Close()

		server.clients.add(connection)
		err := websocket.JSON.Receive(connection, nil)
		if err != nil && err != io.EOF {
			log.Println("WebSocket connection error:", err)
		}
		server.clients.del(connection)
	}))

	err = server.Serve(ln)
	if err != nil && !server.IsClosed() {
		log.Fatalln("WebServer error:", err)
	}
}

func (server *WebServer) Close() {
	atomic.StoreInt32(&server.closed, 1)
	server.Server.Close()
	server.clients.close()
}

func (server *WebServer) IsClosed() bool {
	return atomic.LoadInt32(&server.closed) == 1
}
