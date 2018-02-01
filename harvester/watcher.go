package main

import (
	"log.go/api"
	"github.com/hpcloud/tail"
	"io"
	"log"
)

type watcher struct {
	config   *config
	eventsCh chan<- *api.LogEvent
	tails    []*tail.Tail
}

func newWatcher(config *config, eventsCh chan<- *api.LogEvent) *watcher {
	return &watcher{
		config:   config,
		eventsCh: eventsCh,
		tails:    make([]*tail.Tail, 0, 10),
	}
}

func (watcher *watcher) start() {
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
					watcher.eventsCh <- &api.LogEvent{stream, line.Text}
				}
			}()
		}
	}
}

func (watcher *watcher) stop() {
	for _, t := range watcher.tails {
		t.Stop()
	}
}
