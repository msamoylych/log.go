package main

import (
	"log"
	"net/http"
	"golang.org/x/net/websocket"
	"sync"
	"net"
	"mime"
	"path/filepath"
	"sync/atomic"
	"io"
)

type client struct {
	harvesters *Harvesters
}

type clients struct {
	sync.RWMutex
	clients map[*websocket.Conn]*client
}

func (c clients) add(connection *websocket.Conn, clnt *client) {
	c.Lock()
	defer c.Unlock()

	c.clients[connection] = clnt
}

func (c clients) del(connection *websocket.Conn) {
	c.Lock()
	defer c.Unlock()

	delete(c.clients, connection)
}

func (c clients) send(message interface{}) {
	c.RLock()
	defer c.RUnlock()

	for connection := range c.clients {
		err := websocket.JSON.Send(connection, message)
		if err != nil {
			log.Println("WebSocket send error:", err)
		}
	}
}

func (c clients) close() {
	c.Lock()
	defer c.Unlock()

	for connection := range c.clients {
		connection.Close()
	}
}

type WebServer struct {
	*http.Server
	config     *Config
	harvesters *Harvesters
	events     <-chan *Event
	done       chan struct{}
	clients    clients
	closed     int32
}

func NewWebServer(config *Config, harvesters *Harvesters, events <-chan *Event) (*WebServer) {
	return &WebServer{
		Server:     new(http.Server),
		config:     config,
		harvesters: harvesters,
		events:     events,
		done:       make(chan struct{}),
		clients:    clients{clients: make(map[*websocket.Conn]*client)},
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

	go func() {
		for harvester := range server.harvesters.Added {
			for conn, clnt := range server.clients.clients {
				clnt.harvesters.Merge(harvester)
				websocket.JSON.Send(conn, clnt.harvesters.Controls())
			}
		}
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

	http.Handle("/ws", websocket.Handler(func(connection *websocket.Conn) {
		defer connection.Close()

		clnt := &client{harvesters: server.harvesters.Clone()}
		server.clients.add(connection, clnt)

		err := websocket.JSON.Send(connection, clnt.harvesters.Controls())
		if err != nil {
			log.Println("WebSocket send error:", err)
		}

		err = websocket.JSON.Receive(connection, nil)
		if err != nil && err != io.EOF {
			log.Println("WebSocket receive error:", err)
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
