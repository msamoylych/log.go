package main

import (
	"encoding/json"
	"log"
	"net/http"
	"golang.org/x/net/websocket"
	"sync"
	"io"
)

type client struct {
	con  *websocket.Conn
	done chan bool
}

type clients struct {
	sync.Mutex
	clnts map[int]*client
	n     int
}

func (c clients) add(clnt *client) int {
	c.Lock()
	defer c.Unlock()

	n := c.n
	c.clnts[n] = clnt
	c.n++
	return n
}

func (c clients) del(n int) {
	c.Lock()
	defer c.Unlock()

	delete(c.clnts, n)
}

func (c clients) send(event *Event) {
	c.Lock()
	defer c.Unlock()

	for _, clnt := range c.clnts {
		websocket.JSON.Send(clnt.con, event)
	}
}

var clnts = clients{clnts: make(map[int]*client)}

func WebServer(conf *Config, events <-chan *Event) {
	go func() {
		for event := range events {
			clnts.send(event)
		}
	}()

	addr := conf.WebServer.Address()
	log.Println("WebServer listen at", addr)
	http.Handle("/", http.FileServer(http.Dir("site")))
	http.Handle("/ws", websocket.Handler(ws))
	http.HandleFunc("/controls", controls)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalln("Init WebServer error:", err)
		return
	}
}

func controls(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(Harvesters.Controls())
	if err != nil {
		log.Println("Marshal controls error:", err)
		http.Error(w, http.StatusText(500), 500)
	} else {
		w.Write(resp)
	}
}

func ws(ws *websocket.Conn) {
	defer ws.Close()

	log.Println("WS client connected")

	done := make(chan bool)

	clnt := new(client)
	clnt.con = ws
	clnt.done = done
	n := clnts.add(clnt)

	var data []byte
	for {
		select {
		case <-done:
			log.Println("WS client disconnected")
			clnts.del(n)
			return
		default:
			err := websocket.Message.Receive(ws, &data)
			if err == io.EOF {
				done <- true
			} else if err != nil {
				log.Println("WS error:", err)
			}
		}
	}
}
