package gtmp

import (
	"io"
	"log"
)

type Session struct {
	sessionType string
}

func NewSession(rw io.ReadWriteCloser, server Server, c chan int) (session *Session) {
	defer func() {
		rw.Close()
		<-c
	}()

	handShake(rw)

	inMessageChannel := make(chan *Message, 50)
	outMessageChannel := make(chan *Message, 50)
	inChunkStream := ChunkStream{chunkSize: 128}
	messageStream := MessageStream{inChan: inMessageChannel, outChan: outMessageChannel}

	go messageStream.ReadMessages()
	go inChunkStream.WriteChunks(outMessageChannel, rw)

	inChunkStream.ReadChunks(rw, inMessageChannel)
	log.Println("Session done")
	return
}
