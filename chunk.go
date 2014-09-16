package main

import (
	"bytes"
	"errors"
	"io"
)

type Chunk struct {
	fmt     int
	csid    int
	ts      int
	tsdelta int
	msid    int
	mlen    int
	mtypeid int
	size    int
	payload *bytes.Buffer
}

func NewChunk(maxSize int) *Chunk {
	return &Chunk{size: maxSize}
}

func (c *Chunk) ReadHeader(r io.Reader) (err error) {
	err = c.readFmtId(r)
	if err != nil {
		return
	}

	c.readMsgHeader(r)
	return
}

func (c *Chunk) ReadPayload(r io.Reader) (err error) {
	left := c.mlen - c.payload.Len()
	size := c.size
	if left < size {
		size = left
	}
	n, err := io.CopyN(c.payload, r, int64(size))
	if n != int64(size) {
		err = errors.New("Chunk data copy error")
	}
	return
}

func (c *Chunk) readFmtId(r io.Reader) (err error) {
	i, err := ReadInt(r, 1)
	if err != nil {
		return
	}

	c.fmt = (i >> 6) & 3
	c.csid = i & 0x3f

	var j int
	if c.csid == 0 {
		j, err = ReadInt(r, 1)
		c.csid = j + 64
	} else if c.csid == 0x01 {
		j, err = ReadInt(r, 2)
		c.csid = j + 64
	}

	return
}

func (c *Chunk) readMsgHeader(r io.Reader) (err error) {

	if c.fmt == 0 {
		c.ts, err = ReadInt(r, 3)
		if err != nil {
			return
		}
		c.mlen, err = ReadInt(r, 3)
		if err != nil {
			return
		}
		c.mtypeid, err = ReadInt(r, 1)
		if err != nil {
			return
		}
		c.msid, err = ReadIntLE(r, 4)
		if err != nil {
			return
		}
	}

	if c.fmt == 1 {
		c.tsdelta, err = ReadInt(r, 3)
		if err != nil {
			return
		}
		c.mlen, err = ReadInt(r, 3)
		if err != nil {
			return
		}
		c.mtypeid, err = ReadInt(r, 1)
		if err != nil {
			return
		}
	}

	if c.fmt == 2 {
		c.tsdelta, err = ReadInt(r, 3)
		if err != nil {
			return
		}
	}

	if c.ts == 0xffffff {
		c.ts, err = ReadInt(r, 4)
	}
	if c.tsdelta == 0xffffff {
		c.tsdelta, err = ReadInt(r, 4)
	}

	return
}

func (c *Chunk) CollectMessage(m *Message) {
	c.msid = m.streamid
	c.mlen = m.length
	c.mtypeid = m.typeid
	c.payload = m.payload
}

func (c *Chunk) WriteHeader(w io.Writer) (written int, err error) {
	var b bytes.Buffer
	f := c.fmt << 6
	if c.csid < 64 {
		WriteInt(&b, f+c.csid, 1)
	} else if c.csid < 320 {
		WriteInt(&b, f, 1)
		WriteInt(&b, c.csid-64, 1)
	} else {
		WriteInt(&b, f, 1)
		WriteInt(&b, c.csid-64, 2)
	}

	switch c.fmt {
	case 0:
		WriteInt(&b, c.ts, 3)
		WriteInt(&b, c.mlen, 3)
		WriteInt(&b, c.mtypeid, 1)
		WriteInt(&b, c.msid, 4)
	case 1:
		WriteInt(&b, c.tsdelta, 3)
		WriteInt(&b, c.mlen, 3)
		WriteInt(&b, c.mtypeid, 1)
	case 2:
		WriteInt(&b, c.tsdelta, 3)
	case 3:
		//no header for type three
	default:
		//TODO: warn of erroneous type
	}

	return w.Write(b.Bytes())

}

func (c *Chunk) WritePayload(w io.Writer) (err error) {
	_, err = io.CopyN(w, c.payload, int64(c.size))
	return
}
