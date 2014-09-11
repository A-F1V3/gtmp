package main

import (
	"io"
	"log"
)

const (
	MSG_CHAN_SIZE = 50
)

type Session struct {
	sessionType string
	Server      *Server
	conn        io.ReadWriteCloser
}

func (session *Session) Start(done chan int) {
	defer func() {
		log.Println("Session Close")
		session.conn.Close()
		done <- 1
	}()

	log.Println("Session Start")
	log.Println("Session Apps:", session.Server.Applications)
	handShake(session.conn)

	inMessageChannel := make(chan *Message, MSG_CHAN_SIZE)
	outMessageChannel := make(chan *Message, MSG_CHAN_SIZE)
	chunkStream := NewChunkStream()
	messageStream := NewMessageStream(session.Server)

	go messageStream.WriteMessages(outMessageChannel)
	go messageStream.ReadMessages(inMessageChannel)
	go chunkStream.WriteChunks(outMessageChannel, session.conn)
	chunkStream.ReadChunks(session.conn, inMessageChannel)

}
