package main

import (
	"io"
	"log"
	"bytes"
	"fmt"
	"errors"
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

const (
	USR_STREAM_BEGIN = 0
	USR_STREAM_EOF   = 1
	USR_STREAM_DRY   = 2
	USR_SET_BUF_LEN  = 3
	USR_STREAM_REC   = 4
	USR_PING_REQ     = 6
	USR_PING_RES     = 7
)

type Message struct {
	typeid    int
	length    int
	timestamp int
	streamid  int
	payload   *bytes.Buffer
}

func (msg *Message) String() string {
	return fmt.Sprintf("Message{ Type: %d, Timestamp: %d, Stream: %d, Length: %d }", msg.typeid, msg.timestamp, msg.streamid, msg.lengths)
}

func (msg *Message) addChunk(chunk *Chunk) (*Message, bool, error) {

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
		n, err := io.CopyN(msg.payload, chunk.reader, int64(size))
		if n != int64(size) {
			e := errors.New("Chunk data copy error")
			return msg, false ,e
		}
		if err != nil {
			return msg, false, err
		}
	}

	if size == left {
		return msg, false, nil
	}

	return msg, true, nil
}

func NewControlMessage(typeid int, payload []byte) (*Message){
	message := Message{typeid: typeid, timestamp: 0, streamid: 0, payload: &bytes.Buffer{}}
	message.payload.Write(payload)
	message.length = message.payload.Len()

	return &message
}

func NewSetChunkSizeMessage(chunksize int) (*Message, error) {
	if chunksize < 1 {
		log.Println("Error chunksize is less that one")
	} else if chunksize < 128 {
		log.Println("Warn: chunksize should be at least 128:", chunksize)
	}

	buf := IntToBuf(chunksize, 4)
	return NewControlMessage(MSG_CHUNK_SIZE, buf), nil
}

func NewAbortMessage(csid int) (*Message, error) {
	buf := IntToBuf(csid, 4)
	return NewControlMessage(MSG_ABORT, buf), nil
}

func NewAckMessage(byteCount int) (*Message, error) {
	buf := IntToBuf(byteCount, 4)
	return NewControlMessage(MSG_ACK, buf), nil
}

func NewSetWindowSizeMessage(byteCount int) (*Message, error) {
	buf := IntToBuf(byteCount, 4)
	return NewControlMessage(MSG_ACK_SIZE, buf), nil
}

func NewSetPeerBWMessage(byteCount int, limitType int) (*Message, error) {
	buf := IntToBuf(byteCount, 4)
	buf = append(buf, byte(limitType & 0xff))
	return NewControlMessage(MSG_BANDWIDTH, buf), nil
}

func NewUserCtrlMessage(typeid int, value []byte) (*Message, error) {
	var b bytes.Buffer
	WriteInt(&b, typeid, 2)
	b.Write(value)
	return NewControlMessage(MSG_USER, b.Bytes()), nil
}

func NewStreamBeginMessage(streamid int) (*Message, error) {
	buf := IntToBuf(streamid, 4)
  return NewUserCtrlMessage(USR_STREAM_BEGIN, buf)
}

