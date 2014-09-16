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
			// create a new message struct if no message already exists on this chunk stream
			message = &Message{}
		}

		message.CollectHeader(chunk)
		if message.payload == nil {
			message.payload = &bytes.Buffer{}
			message.payload.Grow(message.length)
		}

		//update chunk with any message info that may be missing, and align buffers
		chunk.CollectMessage(message)

		err = chunk.ReadPayload(input)
		if err != nil {
			return
		}

		//When the payload length matches the length, the full message has been recieved
		if message.payload.Len() < message.length {
			chunkMap[chunk.csid] = message
		} else {
			log.Println("Full Message Parsed: ", message)
			switch message.typeid {
			case MSG_CHUNK_SIZE, MSG_ABORT, MSG_ACK, MSG_ACK_SIZE, MSG_BANDWIDTH:
				// Protocol Control Messages: These messages operate on the chunk stream level
				c.handleProtocolControlMessage(message)
			default:
				messages <- message
			}
			//copy message header data and empty the buffer
			new_message := *message
			new_message.payload = nil
			//place the "empty" message back in the map, for type 1 & 2 chunk fmts
			chunkMap[chunk.csid] = &new_message
		}

	}

}

func (cs *ChunkStream) WriteChunks(messages chan *Message, output io.Writer) error {
	for message := range messages {
		c := NewChunk(cs.writeChunkSize)
		c.fmt = 0
		c.csid = getChunkStreamId(message)
		c.CollectMessage(message)

		switch message.typeid {
		case MSG_CHUNK_SIZE:
			//If we are setting a new chunk size, set it locally for consecutive messages to use
			body := bytes.NewBuffer(message.payload.Bytes())
			cs.writeChunkSize, _ = ReadInt(body, 4)
		}

		// Validate the outbound message
		if message.length != message.payload.Len() {
			log.Println("Outbound message length and buffer size do not match:", message)
		}

		for message.payload.Len() > 0 {
			err := c.WriteHeader(output)
			if err != nil {
				return err
			}

			err = c.WritePayload(output)
			if err != nil {
				return err
			}

			//Write the rest of the message with type 3 chunk
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
