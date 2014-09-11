package main

import (
	"log"
)

const (
	UNDEF = 0
	PUB
	PLAY
)

const (
	DEFAULT_ACK_SIZE = 5000000
)

type MessageStream struct {
	inChan      chan *Message
	outChan     chan *Message
	lastMessage *Message
	metadata    *Message
	audioHeader *Message
	videoHeader *Message
	app         *Application
	stream      *Stream
	server      *Server
	role        int
}

func NewMessageStream(server *Server) *MessageStream {
	return &MessageStream{
		outChan: make(chan *Message),
		server:  server,
	}
}

func (m *MessageStream) ReadMessages(messages chan *Message) {
	for message := range messages {
		m.handleMessage(message)
	}
}

func (m *MessageStream) WriteMessages(messages chan *Message) {
	for message := range m.outChan {
		messages <- message
	}
}

func (m *MessageStream) handleMessage(message *Message) {
	//log.Println("New Message Recieved: ", message)

	switch message.typeid {
	case MSG_AUDIO:
		m.handleAudio(message)
	case MSG_VIDEO:
		log.Println("GOT MEDIA")
	case MSG_AMF_CMD:
		log.Println("Got Command Message")
		m.handleAmfCommand(message)
	default:
		log.Println("unhandled msg type:", message)
	}

	m.lastMessage = message
}

func (m *MessageStream) handleAudio(message *Message) {

}

func (m *MessageStream) handleVideo(message *Message) {

}

func (m *MessageStream) handleAmfCommand(message *Message) {
	command_name, _ := ReadAMF(message.payload)
	log.Println("Command Msg:", command_name.str)
	amfs := make([]AMFObj, 0)

	for {
		obj, err := ReadAMF(message.payload)
		if err != nil {
			break
		}
		amfs = append(amfs, obj)
	}

	switch command_name.str {
	case "connect":
		m.handleConnect(amfs, message)
	case "releaseStream":
		m.handleReleaseStream(amfs, message)
	case "FCPublish":
		m.handleFCPublish(amfs, message)
	case "createStream":
		m.handleCreateStream(amfs, message)
	case "publish":
		m.handlePublish(amfs, message)
	}

}

func (m *MessageStream) handleConnect(amfs []AMFObj, message *Message) {
	log.Println("Handle Connect")
	txnId := amfs[0]
	if txnId.f64 != 1 {
		log.Println("Bad TNX ID: ", txnId)
	}
	commandObject := amfs[1]
	if commandObject.atype != AMF_OBJECT {
		log.Println("COMMAND OBJECT NOT OBJECT:", commandObject)
	}
	if _, ok := commandObject.obj["app"]; !ok || commandObject.obj["app"].str == "" {
		panic("Connect message has no app name")
	}
	appName := commandObject.obj["app"].str
	log.Println("Connect message for app:", appName)

	app, ok := m.server.Applications[appName]
	if !ok {
		log.Println("app doesnt exist, shut it down")
		result := NewAMFResult("_error",
			txnId.f64,
			AMFObj{atype: AMF_NULL},
			NewOnStatusAMFObj("error", "NetConnection.Connect.InvalidApp", "No app "+appName+" on server", nil),
		)
		msg := NewAMFCmdMessage(result)

		m.outChan <- msg
		return
	}

	m.app = app
	log.Println("connected to:", app)

	//log.Println("Command Object: ", commandObject)

	var msg *Message
	var err error

	msg, err = NewSetWindowSizeMessage(DEFAULT_ACK_SIZE)
	m.outChan <- msg

	msg, err = NewSetPeerBWMessage(DEFAULT_ACK_SIZE, 0)
	m.outChan <- msg

	msg, err = NewStreamBeginMessage(message.streamid)
	if err != nil {
		log.Println("msg create error:", err)
	}
	m.outChan <- msg

	result := NewAMFResult(
		"_result",
		txnId.f64,
		AMFObj{atype: AMF_OBJECT,
			obj: map[string]AMFObj{
				"flashVer":     AMFObj{atype: AMF_STRING, str: "FMS/3,0,1,123"},
				"capabilities": AMFObj{atype: AMF_NUMBER, f64: 31},
			},
		},
		NewOnStatusAMFObj(
			"status",
			"NetConnection.Connect.Success",
			"Connection Success.",
			map[string]AMFObj{
				"objectEncoding": AMFObj{
					atype: AMF_NUMBER,
					f64:   0,
				},
			},
		),
	)
	msg = NewAMFCmdMessage(result)

	m.outChan <- msg

}

func (m *MessageStream) handleReleaseStream(amfs []AMFObj, message *Message) {
	log.Println("Handle ReleaseStream")
	txnId := amfs[0].f64
	streamName := amfs[2].str
	log.Println("Release stream:", streamName)
	result := NewAMFResult("_result", txnId)
	msg := NewAMFCmdMessage(result)
	m.outChan <- msg
}

func (m *MessageStream) handleFCPublish(amfs []AMFObj, message *Message) {
	log.Println("Handle FCPublish")
	txnId := amfs[0].f64
	streamName := amfs[2].str
	log.Println("FCPublish:", streamName)
	result := NewAMFResult("_result", txnId)
	msg := NewAMFCmdMessage(result)
	m.outChan <- msg
}

func (m *MessageStream) handleCreateStream(amfs []AMFObj, message *Message) {
	log.Println("Handle CreateStream")
	txnId := amfs[0].f64

	result := NewAMFResult("_result", txnId, AMFObj{atype: AMF_NULL}, AMFObj{atype: AMF_NUMBER, f64: 1})
	msg := NewAMFCmdMessage(result)
	m.outChan <- msg
}

func (m *MessageStream) handlePublish(amfs []AMFObj, message *Message) {
	txnId := amfs[0].f64
	if txnId != 0 {
		log.Println("publish txnid should be 0, not:", txnId)
	}

	streamName := amfs[2].str
	pubType := amfs[3].str

	log.Printf("Publish: %s, Type: %s", streamName, pubType)

	m.app.Publish(streamName)

	m.outChan <- NewAMFStatusMessage("status", "NetStream.Publish.Start", "", nil)

}
