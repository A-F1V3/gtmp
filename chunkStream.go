package main

import (
	"bytes"
	"io"
	"log"
	"sync"
)

type ChunkStream struct {
	sync.Mutex
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
		chunk, err := ReadChunk(input, c.readChunkSize)
		if err != nil {
			return
		}
		message, ok := chunkMap[chunk.csid]
		if !ok {
			message = &Message{payload: &bytes.Buffer{}}
		}

		message, more, err := message.addChunk(chunk)
		if err != nil {
			return
		}

		if !more {
			//log.Println("Full Message Parsed: ", message)
			messages <- message

			//copy message with with new payload buffer
			new_message := *message
			new_message.payload = &bytes.Buffer{}
			chunkMap[chunk.csid] = &new_message
		} else {
			chunkMap[chunk.csid] = message
		}
	}

}

func (cs *ChunkStream) WriteChunks(messages chan *Message, output io.Writer) error {
	for message := range messages {
		c := &Chunk{size: cs.writeChunkSize}
		c.fmt = 0
		c.csid = getChunkStreamId(message)
		c.ts = message.timestamp
		c.msid = message.streamid
		c.mlen = message.length
		c.mtypeid = message.typeid
		c.reader = message.payload

		for n := c.mlen; n > 0; {
			buf := &bytes.Buffer{}

			c.WriteChunkHeader(buf)

			written, _ := io.CopyN(buf, c.reader, int64(c.size))
			n -= int(written)

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

func ReadChunk(r io.Reader, chunkSize int) (c *Chunk, err error) {
	c = &Chunk{size: chunkSize, reader: r}
	err = c.ReadChunkHeader(r)

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
