package main

import (
	"io"
	//"log"
	"bytes"
)

const (
	Tii = 1
)

type Message struct {
	typeid    int
	length    int
	timestamp int
	streamid  int
	payload   *bytes.Buffer
}

func (msg *Message) addChunk(chunk *Chunk) (*Message, bool) {

	switch chunk.fmt {
	case 0:
		msg.timestamp = chunk.ts
		msg.length = chunk.mlen
		msg.typeid = chunk.mtypeid
		msg.streamid = chunk.msid
	case 1:
		msg.timestamp += chunk.tsdelta // check for wraparound !!
		msg.length = chunk.mlen
		msg.typeid = chunk.mtypeid
	case 2:
		msg.timestamp += chunk.tsdelta
	}

	left := msg.length - msg.payload.Len()
	size := chunk.size
	if size > left {
		size = left
	}

	if size > 0 {
		io.CopyN(msg.payload, chunk.reader, int64(size))
	}

	if size == left {
		return msg, false
	}

	return msg, true
}
