package gtmp

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

func main() {

	conf, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		panic(err)
	}

	log.Println(string(conf))
	var config Config
	json.Unmarshal(conf, &config)

	done := make(chan int, len(config.Servers))

	for _, server := range config.Servers {
		log.Println("Starting RTMP Server on", server.Address)
		NewServer(server.Address, done)
	}

	for cnt := len(config.Servers); cnt > 0; cnt-- {
		<-done
	}
}
