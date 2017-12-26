package main

import (
	"strconv"
	"log"
	"io/ioutil"
	"encoding/json"
	"os/user"
)

const confPath = "/.log.go/server.conf"

type Server struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

func (s Server) Address() string {
	return s.Host + ":" + strconv.FormatUint(uint64(s.Port), 10)
}

type Config struct {
	WebServer Server `json:"webServer"`
	LogServer Server `json:"logServer"`
}

func Parse() *Config {
	usr, err := user.Current()
	if err != nil {
		log.Fatalln("Get current user error:", err)
	}

	confFile, err := ioutil.ReadFile(usr.HomeDir + confPath)
	if err != nil {
		log.Fatalln("Read config error:", err)
	}

	config := new(Config)
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
