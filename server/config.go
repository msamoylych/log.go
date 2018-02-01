package main

import (
	"strconv"
	"log"
	"io/ioutil"
	"encoding/json"
	"os/user"
)

const confPath = "/.log.go/server.conf"

type server struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

func (s server) address() string {
	return s.Host + ":" + strconv.FormatUint(uint64(s.Port), 10)
}

type config struct {
	WebServer server `json:"webServer"`
	LogServer server `json:"logServer"`
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

	if config.LogServer.Port == 0 {
		log.Fatalln("LogServer port is not specified")
	}
	if config.WebServer.Port == 0 {
		log.Fatalln("WebServer port is not specified")
	}

	return config
}
