package main

import (
	"bytes"
	"io"
	"log"
)

type ChunkStream struct {
	readChunkSize  int
	writeChunkSize int
}

const (
	DEFAULT_CHUNK_SIZE = 128
)

func NewChunkStream() *ChunkStream {
	return &ChunkStream{
		readChunkSize:  DEFAULT_CHUNK_SIZE,
		writeChunkSize: DEFAULT_CHUNK_SIZE,
	}
}

func (c *ChunkStream) ReadChunks(input io.Reader, messages chan *Message) {
	defer close(messages)

	chunkMap := make(map[int]*Message)

	for {
		chunk := NewChunk(c.readChunkSize)
		err := chunk.ReadHeader(input)
		if err != nil {
			return
		}

		message, ok := chunkMap[chunk.csid]
		if !ok {
			message = &Message{}
		}
		if message.payload == nil {
			message.payload = &bytes.Buffer{}
			message.payload.Grow(chunk.mlen)
		}
		message.CollectHeader(chunk)
		chunk.CollectMessage(message)

		err = chunk.ReadPayload(input)
		if err != nil {
			return
		}

		if message.payload.Len() < message.length {
			chunkMap[chunk.csid] = message
		} else {
			log.Println("Full Message Parsed: ", message)
			switch message.typeid {
			case MSG_CHUNK_SIZE, MSG_ABORT, MSG_ACK, MSG_ACK_SIZE, MSG_BANDWIDTH:
				// Protocol Control Messages
				// These messages operate on the chunk stream level
				c.handleProtocolControlMessage(message)
			default:
				messages <- message
			}
			//copy message with with new payload buffer
			new_message := *message
			new_message.payload = nil
			chunkMap[chunk.csid] = &new_message
		}

	}

}

func (cs *ChunkStream) WriteChunks(messages chan *Message, output io.Writer) error {
	for message := range messages {
		c := NewChunk(cs.writeChunkSize)
		c.fmt = 0
		c.csid = getChunkStreamId(message)
		c.ts = message.timestamp
		c.msid = message.streamid
		c.mlen = message.length
		c.mtypeid = message.typeid
		c.payload = message.payload

		switch message.typeid {
		case MSG_CHUNK_SIZE:
			body := bytes.NewBuffer(message.payload.Bytes())
			cs.writeChunkSize, _ = ReadInt(body, 4)
		}

		if message.length != message.payload.Len() {
			log.Println("uh oh")
		}

		for message.payload.Len() > 0 {
			buf := &bytes.Buffer{}

			c.WriteHeader(buf)

			c.WritePayload(buf)

			_, err := output.Write(buf.Bytes())
			if err != nil {
				return err
			}
			//log.Printf("Chunk written: %s, Bytes: %d",c,r)
			c.fmt = 3
		}
		log.Printf("Message Sent: %d, on cs %d", message, c.csid)

	}
	return nil
}

func (c *ChunkStream) handleProtocolControlMessage(message *Message) (err error) {
	switch message.typeid {
	case MSG_CHUNK_SIZE:
		var newChunkSize int
		newChunkSize, err = ReadInt(message.payload, 4)
		if newChunkSize > 0 && newChunkSize < 0x7FFFFFFF {
			c.readChunkSize = newChunkSize
		}
	case MSG_ABORT:
	case MSG_ACK:
	case MSG_ACK_SIZE:
	case MSG_BANDWIDTH:
	}
	return
}

func getChunkStreamId(message *Message) (csid int) {
	switch message.typeid {
	case MSG_AMF_CMD:
		return 3
	default:
		return 2
	}
}
