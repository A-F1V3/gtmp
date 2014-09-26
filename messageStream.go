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
	inChan          chan *Message
	outChan         chan *Message
	mediaChan       chan *Message
	lastSentMessage *Message
	metadata        *Message
	audioHeader     *Message
	videoHeader     *Message
	app             *Application
	stream          *Stream
	server          *Server
	role            int
}

func NewMessageStream(server *Server) *MessageStream {
	return &MessageStream{
		outChan:   make(chan *Message),
		mediaChan: make(chan *Message),
		server:    server,
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

func (m *MessageStream) QueueClientMessage(message *Message) {
	m.outChan <- message
}

func (m *MessageStream) publishMessage(message *Message) {
	if m.stream != nil {
		m.stream.PublishMessage(message)
	} else {
		log.Println("Stream to publish to")
	}
}

func (m *MessageStream) handleMessage(message *Message) {
	//log.Println("New Message Recieved: ", message)

	switch message.typeid {
	case MSG_AUDIO:
		log.Println("GOT AUDIO")
		m.handleAudio(message)
	case MSG_VIDEO:
		log.Println("GOT VIDEO")
		m.handleVideo(message)
	case MSG_AMF_META:
		log.Println("GOT METADATA")
		m.handleMetadata(message)
	case MSG_AMF_CMD:
		log.Println("Got Command Message")
		m.handleAmfCommand(message)
	case MSG_USER:
		log.Println("Got User Control Message")
		m.handleUserCtrlMessage(message)
	default:
		log.Println("unhandled msg type:", message)
	}

}

func (m *MessageStream) handleAudio(message *Message) {
	m.publishMessage(message)
}

func (m *MessageStream) handleVideo(message *Message) {
	m.publishMessage(message)
}

func (m *MessageStream) handleMetadata(message *Message) {
	m.publishMessage(message)
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
	case "play":
		m.handlePlay(amfs, message)
	}

}

func (m *MessageStream) handleUserCtrlMessage(message *Message) {
	log.Println("Handle User CTRL Message")
	ctrlType, err := ReadInt(message.payload, 2)
	if err != nil {
		log.Println("Busted ctrl message")
		return
	}
	log.Println("Got user Ctrl type:", ctrlType)
	switch ctrlType {
	case USR_STREAM_BEGIN:
	case USR_STREAM_EOF:
	case USR_STREAM_DRY:
	case USR_SET_BUF_LEN:
		streamID, _ := ReadInt(message.payload, 4)
		bufferMS, _ := ReadInt(message.payload, 4)
		log.Printf("Got set buffer length msg. Stream Id: %d, Buffer Millisec: %d", streamID, bufferMS)
	case USR_STREAM_REC:
	case USR_PING_REQ:
	case USR_PING_RES:
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
		result := NewAMFResult("_result",
			txnId.f64,
			AMFObj{atype: AMF_NULL},
			NewOnStatusAMFObj("error", "NetConnection.Connect.InvalidApp", "No app "+appName+" on server", nil),
		)
		msg := NewAMFCmdMessage(result)

		m.QueueClientMessage(msg)
		return
	}

	m.app = app
	log.Println("connected to:", app)

	//log.Println("Command Object: ", commandObject)

	var msg *Message
	var err error

	msg, err = NewSetWindowSizeMessage(DEFAULT_ACK_SIZE)
	m.QueueClientMessage(msg)

	msg, err = NewSetPeerBWMessage(DEFAULT_ACK_SIZE, 2)
	m.QueueClientMessage(msg)

	msg, err = NewStreamBeginMessage(message.streamid)
	m.QueueClientMessage(msg)

	msg, err = NewSetChunkSizeMessage(4096)
	if err != nil {
		log.Println("msg create error:", err)
	}
	m.QueueClientMessage(msg)

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

	m.QueueClientMessage(msg)

}

func (m *MessageStream) handleReleaseStream(amfs []AMFObj, message *Message) {
	log.Println("Handle ReleaseStream")
	txnId := amfs[0].f64
	streamName := amfs[2].str
	log.Println("Release stream:", streamName)
	result := NewAMFResult("_result", txnId)
	msg := NewAMFCmdMessage(result)
	m.QueueClientMessage(msg)
}

func (m *MessageStream) handleFCPublish(amfs []AMFObj, message *Message) {
	log.Println("Handle FCPublish")
	txnId := amfs[0].f64
	streamName := amfs[2].str
	log.Println("FCPublish:", streamName)
	result := NewAMFResult("_result", txnId)
	msg := NewAMFCmdMessage(result)
	m.QueueClientMessage(msg)
}

func (m *MessageStream) handleCreateStream(amfs []AMFObj, message *Message) {
	log.Println("Handle CreateStream")
	txnId := amfs[0].f64

	result := NewAMFResult("_result", txnId, AMFObj{atype: AMF_NULL}, AMFObj{atype: AMF_NUMBER, f64: 1})
	msg := NewAMFCmdMessage(result)
	m.QueueClientMessage(msg)
}

func (m *MessageStream) handlePublish(amfs []AMFObj, message *Message) {
	txnId := amfs[0].f64
	if txnId != 0 {
		log.Println("publish txnid should be 0, not:", txnId)
	}

	streamName := amfs[2].str
	pubType := "unknown"
	//pubType := amfs[3].str

	log.Printf("Publish: %s, Type: %s", streamName, pubType)

	stream, err := m.app.Publish(streamName)
	var res *Message
	if err != nil {
		log.Printf("unable to publish stream")
		res = NewAMFStatusMessage(0, "error", "NetStream.Publish.BadName", err.Error(), nil)
	} else {
		m.stream = stream
		res = NewAMFStatusMessage(0, "status", "NetStream.Publish.Start", streamName+" is now Published", map[string]AMFObj{"details": AMFObj{atype: AMF_STRING, str: streamName}, "clientid": AMFObj{atype: AMF_NUMBER, f64: 1.0}})
	}
	res.streamid = message.streamid

	m.QueueClientMessage(res)

}

func (m *MessageStream) handlePlay(amfs []AMFObj, message *Message) {
	txnId := amfs[0].f64
	if txnId != 0 {
		log.Println("play txnid should be 0, not:", txnId)
	}
	streamName := amfs[2].str
	log.Println("Play the stream:", streamName)
	for i := 3; i < len(amfs); i++ {
		log.Println("Play amf ", amfs[i])
	}

	//set a chunk size here?
	chunkMsg, _ := NewSetChunkSizeMessage(2048)
	m.QueueClientMessage(chunkMsg)

	stream, err := m.app.Play(streamName)
	var res *Message
	if err != nil {
		res = NewAMFStatusMessage(txnId, "error", "NetStream.Play.StreamNotFound", err.Error(), nil)
	} else {
		m.stream = stream
		stream.Subscribe(m.mediaChan)
		go func() {
			for mediaMessage := range m.mediaChan {
				m.QueueClientMessage(mediaMessage)
			}
		}()
		res = NewAMFStatusMessage(txnId, "status", "NetStream.Play.Start", "", nil)
	}

	m.QueueClientMessage(res)
}
