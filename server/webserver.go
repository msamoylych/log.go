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

const (
	controlsType = "controls"
	messageType  = "message"
)

type Type string

type controls struct {
	Type                        `json:"type"`
	Streams map[string][]string `json:"streams"`
	Nodes   map[string][]string `json:"nodes"`
}

type message struct {
	Type          `json:"type"`
	Stream string `json:"stream"`
	Node   string `json:"node"`
	Text   string `json:"text"`
}

func newControls() *controls {
	return &controls{controlsType, make(map[string][]string), make(map[string][]string)}
}

func newMessage(stream string, node string, text string) *message {
	return &message{messageType, stream, node, text}
}

type harvesters struct {
	sync.RWMutex
	harvesters map[string][]string
}

func (h *harvesters) add(hrv *harvester) {
	h.Lock()
	defer h.Unlock()

	h.harvesters[hrv.node] = hrv.streams
}

func (h *harvesters) del(hrv *harvester) {
	h.Lock()
	defer h.Unlock()

	delete(h.harvesters, hrv.node)
}

func (h *harvesters) controls() *controls {
	h.RLock()
	defer h.RUnlock()

	controls := newControls()
	for node, streams := range h.harvesters {
		for _, stream := range streams {
			controls.Streams[stream] = append(controls.Streams[stream], node)
			controls.Nodes[node] = append(controls.Nodes[node], stream)
		}
	}
	return controls
}

type clients struct {
	sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

func (c clients) add(connection *websocket.Conn) {
	c.Lock()
	defer c.Unlock()

	c.clients[connection] = struct{}{}
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

type webServer struct {
	*http.Server
	config       *config
	harvestersCh <-chan *harvester
	eventsCh     <-chan *logEvent
	harvesters   *harvesters
	clients      *clients
	closed       int32
	doneCh       chan struct{}
}

func newWebServer(config *config, harvestersCh <-chan *harvester, eventsCh <-chan *logEvent) (*webServer) {
	return &webServer{
		Server:       new(http.Server),
		config:       config,
		harvestersCh: harvestersCh,
		eventsCh:     eventsCh,
		harvesters:   &harvesters{harvesters: make(map[string][]string)},
		clients:      &clients{clients: make(map[*websocket.Conn]struct{})},
		doneCh:       make(chan struct{}),
	}
}

func (server *webServer) listen() {
	addr := server.config.WebServer.address()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Init webServer error:", err)
	}
	log.Println("webServer listen at", addr)

	go func() {
		for harvester := range server.harvestersCh {
			if harvester.add {
				server.harvesters.add(harvester)
				server.clients.send(server.harvesters.controls())
			} else {
				server.harvesters.del(harvester)
			}
		}
	}()

	go func() {
		for logEvent := range server.eventsCh {
			msg := newMessage(logEvent.stream, logEvent.node, logEvent.text)
			server.clients.send(msg)
		}
		server.doneCh <- struct{}{}
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
			w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(name)))
			w.Write(data)
		}
	})

	http.Handle("/ws", websocket.Handler(func(connection *websocket.Conn) {
		defer connection.Close()

		server.clients.add(connection)

		err := websocket.JSON.Send(connection, server.harvesters.controls())
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
	if err != nil && !server.isClosed() {
		log.Fatalln("webServer error:", err)
	}
}

func (server *webServer) close() {
	atomic.StoreInt32(&server.closed, 1)
	server.Server.Close()
}

func (server *webServer) isClosed() bool {
	return atomic.LoadInt32(&server.closed) == 1
}
