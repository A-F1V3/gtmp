package main

import (
  "io"
  "log"
  "bytes"
)

type ChunkStream struct {
  chunkSize int
}

func (c *ChunkStream) readChunks(input io.Reader, messages chan *Message) {
  chunkMap := make(map[int]*Message)

  for {
    chunk := readChunk(input, 128)
    log.Println("Chunk: ",chunk)
    message, ok := chunkMap[chunk.csid]
    if !ok {
      message = &Message{payload:&bytes.Buffer{}}
    }

    message, more := message.addChunk(chunk)

    if !more {
      log.Println("thats a message: ", message)
      messages <- message
      new_message := *message
      new_message.payload = &bytes.Buffer{}
      chunkMap[chunk.csid] = &new_message
    } else {
      chunkMap[chunk.csid] = message
    }
  }

}

func (c *ChunkStream) writeChunks(messages chan *Message, output io.Writer) {

}


func readChunk(r io.Reader, chunkSize int) (c *Chunk) {
  c = &Chunk{size: chunkSize, reader: r}

  i := ReadInt(r, 1)
  c.fmt = (i>>6)&3;
  c.csid = i&0x3f;

  if c.csid == 0 {
    j := ReadInt(r, 1)
    c.csid = j + 64
  } else if c.csid == 0x01 {
    j := ReadInt(r, 2)
    c.csid = j + 64
  }

  if c.fmt == 0 {
    c.ts = ReadInt(r, 3)
    c.mlen = ReadInt(r, 3)
    c.mtypeid = ReadInt(r, 1)
    c.msid = ReadIntLE(r, 4)
  }

  if c.fmt == 1 {
    c.tsdelta = ReadInt(r, 3)
    c.mlen = ReadInt(r, 3)
    c.mtypeid = ReadInt(r, 1)
  }

  if c.fmt == 2 {
    c.tsdelta = ReadInt(r, 3)
  }

  if c.ts == 0xffffff {
    c.ts = ReadInt(r, 4)
  }
  if c.tsdelta == 0xffffff {
    c.tsdelta = ReadInt(r, 4)
  }

  return
}
