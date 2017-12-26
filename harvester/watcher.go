package main

import (
	"log.go/api"
	"github.com/hpcloud/tail"
	"io"
	"log"
)

type Watcher struct {
	tails []*tail.Tail
}

func NewWatcher(config *Config, events chan<- api.LogEvent) *Watcher {
	watcher := Watcher{tails: make([]*tail.Tail, 0, 10)}

	for stream, files := range config.LogStreams {
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
					events <- api.LogEvent{Stream: stream, Msg: line.Text}
				}
			}()
		}
	}

	return &watcher
}

func (w *Watcher) Stop() {
	for _, t := range w.tails {
		t.Stop()
	}
}
