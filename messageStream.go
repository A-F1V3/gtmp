package main

import (
	"log"
)

const (
	UNDEF = 0
	PUB
	PLAY
)

type MessageStream struct {
	inChan      chan *Message
	outChan     chan *Message
	lastMessage *Message
	metadata    *Message
	audioHeader *Message
	videoHeader *Message
	app         string
	name        string
	role        int
}

func (m *MessageStream) ReadMessages() {
	for message := range m.inChan {
		handleMessage(message, m)
	}
}

func (m *MessageStream) WriteMessages(messages chan *Message) {

}

func handleMessage(message *Message, mStream *MessageStream) {
	log.Println("New Message Recieved: ", message)

	switch message.typeid {
	case MSG_AUDIO, MSG_VIDEO:
		log.Println("GOT MEDIA")
	case MSG_AMF_CMD:
		log.Println("Got Command Message")
		amf := ReadAMF(message.payload)
		log.Println("Command Msg:", amf.str)
	}

	mStream.lastMessage = message
}
