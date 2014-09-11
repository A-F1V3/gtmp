package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

func main() {

	configFile := flag.String("config", "config/config.json", "JSON configuration file")
	conf, err := ioutil.ReadFile(*configFile)
	if err != nil {
		panic(err)
	}

	log.Println(string(conf))
	var config Config
	json.Unmarshal(conf, &config)

	done := make(chan int, len(config.Servers))

	for _, server := range config.Servers {
		log.Println("Starting RTMP Server on", server.Address)
		go server.Start(done)
	}

	for cnt := len(config.Servers); cnt > 0; cnt-- {
		<-done
	}
}
