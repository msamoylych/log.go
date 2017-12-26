package main

type Controls struct {
	Streams map[string][]string `json:"streams"`
	Nodes   map[string][]string `json:"nodes"`
}

type Event struct {
	Node    string `json:"node"`
	Stream  string `json:"stream"`
	Message string `json:"message"`
}
