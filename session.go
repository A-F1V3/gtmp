package main

import (
	"io"
)

type Session struct {
	sessionType string
}

func NewSession(rw io.ReadWriteCloser, server Server, c chan int) (session *Session) {
  defer rw.Close()

  handShake(rw)

	inMessageChannel := make(chan *Message, 50)
	inChunkStream := ChunkStream{chunkSize: 128}
	messageStream := MessageStream{inChan: inMessageChannel}

	go inChunkStream.ReadChunks(rw, inMessageChannel)
	messageStream.ReadMessages()

	return
}
