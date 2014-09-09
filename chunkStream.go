package main

import (
	"bytes"
	"io"
	"log"
	"sync"
)

type ChunkStream struct {
	sync.Mutex
	chunkSize int
}

const (
  DEFAULT_CHUNK_SIZE = 128
)

func (c *ChunkStream) ReadChunks(input io.Reader, messages chan *Message) {
	defer close(messages)

  chunkMap := make(map[int]*Message)
  chunkSize := DEFAULT_CHUNK_SIZE

	for {
		chunk, err := ReadChunk(input, chunkSize)
    if err != nil {
      return
    }
		message, ok := chunkMap[chunk.csid]
		if !ok {
			message  = &Message{payload: &bytes.Buffer{}}
		}

		message, more, err := message.addChunk(chunk)
		if err != nil {
			return
		}

		if !more {
			log.Println("Full Message Parsed: ", message)
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

func (cs *ChunkStream) WriteChunks(messages chan *Message, output io.Writer) {
	for message := range messages {
		log.Println("CHUNK IT:", message)
		c := &Chunk{size: cs.chunkSize}
		c.fmt = 0
		c.csid = 2
		c.ts = message.timestamp
		c.msid = message.streamid
		c.mlen = message.length
		c.mtypeid = message.typeid
		c.reader = message.payload

		for n := c.mlen; n > 0; {
			c.WriteChunkHeader(output)

			written, _ :=	io.CopyN(output, c.reader, int64(c.size))
			n -= int(written)

			if n > 0 && n < c.size {
				log.Println("Writing moar")
				tt, _ :=output.Write(make([]byte,n))
				n -= int(tt)
			}
		}
		log.Println("done with:", message)

	}
}

func ReadChunk(r io.Reader, chunkSize int) (c *Chunk, err error) {
	c = &Chunk{size: chunkSize, reader: r}
	err = c.ReadChunkHeader(r)

	return
}


