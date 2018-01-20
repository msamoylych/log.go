package main

const (
	controls = "controls"
	event    = "event"
)

type Controls struct {
	Type    string              `json:"type"`
	Streams map[string][]string `json:"streams"`
	Nodes   map[string][]string `json:"nodes"`
}

type Event struct {
	Type    string `json:"type"`
	Node    string `json:"node"`
	Stream  string `json:"stream"`
	Message string `json:"message"`
}

func NewControls() (*Controls) {
	return &Controls{
		Type:    controls,
		Streams: make(map[string][]string),
		Nodes:   make(map[string][]string),
	}
}

func NewEvent(node string, stream string, msg string) (*Event) {
	return &Event{
		Type:    event,
		Node:    node,
		Stream:  stream,
		Message: msg,
	}
}
