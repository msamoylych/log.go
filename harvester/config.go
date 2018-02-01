package main

import (
	"strconv"
	"os/user"
	"log"
	"io/ioutil"
	"encoding/json"
)

const confPath = "/.log.go/harvester.conf"

type server struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

func (s server) address() string {
	return s.Host + ":" + strconv.FormatUint(uint64(s.Port), 10)
}

type config struct {
	NodeName   string              `json:"nodeName"`
	Server     server              `json:"server"`
	LogStreams map[string][]string `json:"logStreams"`
}

func (c *config) Streams() []string {
	streams := make([]string, 0, len(c.LogStreams))
	for stream := range c.LogStreams {
		streams = append(streams, stream)
	}
	return streams
}

func parse() *config {
	usr, err := user.Current()
	if err != nil {
		log.Fatalln("Get current user error:", err)
	}

	confFile, err := ioutil.ReadFile(usr.HomeDir + confPath)
	if err != nil {
		log.Fatalln("Read config error:", err)
	}

	config := new(config)
	err = json.Unmarshal(confFile, config)
	if err != nil {
		log.Fatalln("Parse config error:", err)
	}

	if config.NodeName == "" {
		log.Fatalln("Node name is not specified")
	}
	if config.Server.Host == "" {
		log.Fatalln("Server host is not specified")
	}
	if config.Server.Port == 0 {
		log.Fatalln("Server port is not specified")
	}

	return config
}
