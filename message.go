package main

import (
	"io"
	//"log"
	"bytes"
	"fmt"
)

const (
	MSG_CHUNK_SIZE  = 1
	MSG_ABORT       = 2
	MSG_ACK         = 3
	MSG_USER        = 4
	MSG_ACK_SIZE    = 5
	MSG_BANDWIDTH   = 6
	MSG_EDGE        = 7
	MSG_AUDIO       = 8
	MSG_VIDEO       = 9
	MSG_AMF3_META   = 15
	MSG_AMF3_SHARED = 16
	MSG_AMF3_CMD    = 17
	MSG_AMF_META    = 18
	MSG_AMF_SHARED  = 19
	MSG_AMF_CMD     = 20
	MSG_AGGREGATE   = 22
	MSG_MAX         = 22
)

type Message struct {
	typeid    int
	length    int
	timestamp int
	streamid  int
	payload   *bytes.Buffer
}

func (msg *Message) String() string {
	return fmt.Sprintf("Message{ Type: %d, Timestamp: %d, Stream: %d }", msg.typeid, msg.timestamp, msg.streamid)
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


