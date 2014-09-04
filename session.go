package main

import (
	"io"
)

type Session struct {
	sessionType string
}

func startNewSession(rw io.ReadWriter, c chan int) (session *Session) {
	inMessageChannel := make(chan *Message, 50)
	inChunkStream := ChunkStream{chunkSize: 128}
	messageStream := MessageStream{inChan: inMessageChannel}
	go inChunkStream.ReadChunks(rw, inMessageChannel)

	go messageStream.ReadMessages()

	return
}
