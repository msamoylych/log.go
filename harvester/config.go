package main

import (
	"strconv"
	"log"
)

type Server struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

type Config struct {
	NodeName   string              `json:"nodeName"`
	Server     Server              `json:"server"`
	LogStreams map[string][]string `json:"logStreams"`
}

func (s Server) Address() string {
	return s.Host + ":" + strconv.FormatUint(uint64(s.Port), 10)
}

func (c Config) Check() {
	if (c.NodeName == "") {
		log.Fatalln("Node name is not specified")
	}
	if (c.Server.Host == "") {
		log.Fatalln("Server host is not specified")
	}
	if (c.Server.Port == 0) {
		log.Fatalln("Server port is not specified")
	}
}
