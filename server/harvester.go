package main

import (
	"sync"
)

type Harvester struct {
	Node    string
	Streams []string
}

type Harvesters struct {
	sync.RWMutex
	Harvesters map[string][]string
	Added      chan *Harvester
}

func (h *Harvesters) Add(node string, streams []string) {
	h.Lock()
	defer h.Unlock()

	h.Harvesters[node] = streams
	h.Added <- &Harvester{Node: node, Streams: streams}
}

func (h *Harvesters) Merge(harvester *Harvester) {
	h.Lock()
	defer h.Unlock()

	h.Harvesters[harvester.Node] = harvester.Streams
}

func (h *Harvesters) Del(node string) {
	h.Lock()
	defer h.Unlock()

	delete(h.Harvesters, node)
}

func (h *Harvesters) Controls() *Controls {
	h.RLock()
	defer h.RUnlock()

	controls := NewControls()
	for node, streams := range h.Harvesters {
		controls.Nodes[node] = streams
		for _, stream := range streams {
			controls.Streams[stream] = append(controls.Streams[stream], node)
		}
	}
	return controls
}

func NewHarvesters() *Harvesters {
	return &Harvesters{Harvesters: make(map[string][]string), Added: make(chan *Harvester)}
}

func (h *Harvesters) Clone() *Harvesters {
	harvesters := Harvesters{Harvesters: make(map[string][]string)}
	for node, streams := range h.Harvesters {
		for _, stream := range streams {
			harvesters.Harvesters[node] = append(harvesters.Harvesters[node], stream)
		}
	}
	return &harvesters
}
