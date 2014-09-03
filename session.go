package main

import (
	"io"
	"log"
)

type Session struct {
	sessionType string
}

func startNewSession(rw io.ReadWriter, c chan int) (session *Session) {
	inMessageChannel := make(chan *Message, 50)
	inChunkStream := ChunkStream{chunkSize: 128}
	go inChunkStream.readChunks(rw, inMessageChannel)
	go messageReader(inMessageChannel)

	return
}

func messageReader(messages chan *Message) {
	for message := range messages {
		log.Println("I got a message!: ", message)
	}
}
