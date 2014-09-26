package main

import (
	"bytes"
	"fmt"
	"log"
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

const (
	VID_KEY_FRAME   = 1
	VID_INTER_FRAME = 2
	VID_DISP_FRAME  = 3
	VID_GEN_FRAME   = 4
	VID_CMD_FRAME   = 5
)

const (
	VID_CODEC_SOR     = 2
	VID_CODEC_SCREEN  = 3
	VID_CODEC_VP6     = 4
	VID_CODEC_VP6A    = 5
	VID_CODEC_SCREEN2 = 6
	VID_CODEC_AVC     = 7
)

const (
	AVC_SEQ_HDR = 0
	AVC_NALU    = 1
	AVN_SEQ_END = 2
)

const (
	AUDIO_CODEC_MP3 = 2
	AUDIO_CODEC_AAC = 10
)

const (
	AAC_SEQ_HDR = 0
	AAC_RAW     = 1
)

type Message struct {
	typeid    int
	length    int
	timestamp int
	tsdelta   int
	streamid  int
	payload   *bytes.Buffer
}

func (msg *Message) String() string {
	return fmt.Sprintf("Message{ Type: %d, Timestamp: %d, Stream: %d, Length: %d }", msg.typeid, msg.timestamp, msg.streamid, msg.length)
}

func (msg *Message) CollectHeader(chunk *Chunk) {
	switch chunk.fmt {
	case 0:
		msg.timestamp = chunk.ts
		msg.length = chunk.mlen
		msg.typeid = chunk.mtypeid
		msg.streamid = chunk.msid
	case 1:
		msg.tsdelta = chunk.tsdelta
		msg.timestamp += chunk.tsdelta // check for wraparound !!
		msg.length = chunk.mlen
		msg.typeid = chunk.mtypeid
	case 2:
		msg.tsdelta = chunk.tsdelta
		msg.timestamp += chunk.tsdelta
	}
}

func NewControlMessage(typeid int, payload []byte) *Message {
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
	log.Println("New Ack Message:", byteCount)
	buf := IntToBuf(byteCount, 4)
	return NewControlMessage(MSG_ACK, buf), nil
}

func NewSetWindowSizeMessage(byteCount int) (*Message, error) {
	buf := IntToBuf(byteCount, 4)
	return NewControlMessage(MSG_ACK_SIZE, buf), nil
}

func NewSetPeerBWMessage(byteCount int, limitType int) (*Message, error) {
	buf := IntToBuf(byteCount, 4)
	buf = append(buf, byte(limitType&0xff))
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

func NewAMFCmdMessage(body []AMFObj) *Message {
	var b bytes.Buffer
	for _, v := range body {
		WriteAMF(&b, v)
	}
	return NewControlMessage(MSG_AMF_CMD, b.Bytes())
}

func NewAMFStatusMessage(txnid float64, level string, code string, description string, more map[string]AMFObj) *Message {
	body := []AMFObj{
		AMFObj{atype: AMF_STRING, str: "onStatus"},
		AMFObj{atype: AMF_NUMBER, f64: txnid},
		AMFObj{atype: AMF_NULL},
		NewOnStatusAMFObj(level, code, description, more),
	}
	return NewAMFCmdMessage(body)
}
