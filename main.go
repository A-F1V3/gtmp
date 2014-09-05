package main

import (
	"log"
)

func main() {
	log.Println("Starting RTMP Server")
	done := make(chan int)
	NewServer(":1935", done)
	<-done
}
