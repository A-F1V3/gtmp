package main

import (
	"bytes"
	"log"
)

type Stream struct {
	name      string
	players   []chan *Message
	inChannel chan *Message
	subChan   chan chan *Message
	vHdr      *Message
	vKeyframe *Message
	aHdr      *Message
	metadata  *Message
}

func NewStream(name string) *Stream {
	return &Stream{
		name:      name,
		inChannel: make(chan *Message),
		players:   make([]chan *Message, 0),
		subChan:   make(chan chan *Message),
	}
}

func (s *Stream) StartStream() {

	for {
		select {
		case newSub, _ := <-s.subChan:
			if s.metadata != nil {
				newSub <- s.metadata
			}
			if s.vHdr != nil {
				newSub <- s.vHdr
			}
			if s.aHdr != nil {
				newSub <- s.aHdr
			}
			if s.vKeyframe != nil {
				newSub <- s.vKeyframe
			}
			s.players = append(s.players, newSub)
		case message, message_ok := <-s.inChannel:
			if !message_ok {
				break
			}
			s.cacheMeta(message)
			for _, subscriber := range s.players {
				subscriber <- message
			}
		}
	}

}

func (s *Stream) PublishMessage(message *Message) {
	s.inChannel <- message
}

func (s *Stream) Subscribe(sub chan *Message) {
	s.subChan <- sub
}

func (s *Stream) cacheMeta(message *Message) {
	switch message.typeid {
	case MSG_AUDIO:
		audioData := message.payload.Bytes()
		audioCodec := audioData[0] >> 4
		if audioCodec == AUDIO_CODEC_AAC {
			if audioData[1] == AAC_SEQ_HDR {
				log.Println("Got AAC Header")
				s.aHdr = message
			}
		}
	case MSG_VIDEO:
		videoData := message.payload.Bytes()
		frameType := videoData[0] >> 4
		vidCodec := videoData[0] & 0x0F
		if frameType == VID_KEY_FRAME && vidCodec == VID_CODEC_AVC {
			avcPacketType := videoData[1]
			switch avcPacketType {
			case AVC_SEQ_HDR:
				log.Println("Got AVC Header")
				s.vHdr = message
			case AVC_NALU:
				log.Println("Got AVC Keyframe")
				s.vKeyframe = message
			}
		}
	case MSG_AMF_META:
		metaData := message.payload.Bytes()
		mReader := bytes.NewBuffer(metaData)
		amf, _ := ReadAMF(mReader)
		if amf.str == "onMetaData" {
			log.Println("Got onMetaData Header")
			s.metadata = message
		}
	}
}
