package main

import "sync"

type HarvesterMap struct {
	sync.RWMutex
	harvesters map[string][]string
}

func (h HarvesterMap) add(node string, streams []string) {
	h.Lock()
	defer h.Unlock()

	h.harvesters[node] = streams
}

func (h HarvesterMap) del(node string) {
	h.Lock()
	defer h.Unlock()

	delete(h.harvesters, node)
}

func (h HarvesterMap) Controls() *Controls {
	h.RLock()
	defer h.RUnlock()

	controls := &Controls{Streams: make(map[string][]string), Nodes: make(map[string][]string)}
	for node, streams := range h.harvesters {
		controls.Nodes[node] = streams
		for _, stream := range streams {
			if controls.Streams[stream] == nil {
				controls.Streams[stream] = make([]string, 0, 10)
			}
			controls.Streams[stream] = append(controls.Streams[stream], node)
		}
	}
	return controls
}

func NewHarvesterMap() *HarvesterMap {
	return &HarvesterMap{harvesters: make(map[string][]string)}
}
