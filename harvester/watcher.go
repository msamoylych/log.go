package main

import (
	"log.go/api"
	"github.com/hpcloud/tail"
	"io"
	"log"
)

type Watcher struct {
	config *Config
	events chan<- api.LogEvent
	tails  []*tail.Tail
}

func NewWatcher(config *Config, events chan<- api.LogEvent) *Watcher {
	return &Watcher{
		config: config,
		events: events,
		tails:  make([]*tail.Tail, 0, 10),
	}
}

func (watcher *Watcher) Start() {
	for stream, files := range watcher.config.LogStreams {
		for _, file := range files {
			seekInfo := tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}
			t, err := tail.TailFile(file, tail.Config{Follow: true, ReOpen: true, Location: &seekInfo, Logger: tail.DiscardingLogger})
			if err != nil {
				log.Fatalln("Init watcher error:", err)
			}
			watcher.tails = append(watcher.tails, t)
			log.Println("Start watching file:", file)

			go func() {
				for line := range t.Lines {
					watcher.events <- api.LogEvent{Stream: stream, Msg: line.Text}
				}
			}()
		}
	}
}

func (watcher *Watcher) Stop() {
	for _, t := range watcher.tails {
		t.Stop()
	}
}
