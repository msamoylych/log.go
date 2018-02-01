package main

type harvester struct {
	add bool
	node string
	streams []string
}

type logEvent struct {
	node string
	stream string
	text string
}
