package main

import (
	"log"
  "bytes"
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

func (m *MessageStream) ReadMessages() {
	for message := range m.inChan {
		m.handleMessage(message)
	}
}

func (m *MessageStream) WriteMessages(messages chan *Message) {

}

func (m *MessageStream) handleMessage(message *Message) {
	log.Println("New Message Recieved: ", message)

	switch message.typeid {
	case MSG_AUDIO:
    m.handleAudio(message)
  case MSG_VIDEO:
		log.Println("GOT MEDIA")
	case MSG_AMF_CMD:
		log.Println("Got Command Message")
		m.handleAmfCommand(message)
  default:
    log.Println("unhandled msg type:",message)
	}

	m.lastMessage = message
}

func (m *MessageStream) handleAudio(message *Message) {

}

func (m *MessageStream) handleVideo(message *Message) {

}

func (m *MessageStream) handleAmfCommand(message *Message) {
  command_name := ReadAMF(message.payload)
  log.Println("Command Msg:", command_name.str)
  switch command_name.str {
  case "connect":
    m.handleConnect(message)
  }

}

func (m MessageStream) handleConnect(message *Message) {
  log.Println("Handle Connect")
  txnId := ReadAMF(message.payload)
  if txnId.f64 != 1 { log.Println("Bad TNX ID: ", txnId) }
  commandObject := ReadAMF(message.payload)
  if commandObject.atype != AMF_OBJECT { log.Println ("COMMAND OBJECT NOT OBJECT:", commandObject) }
  if _, ok := commandObject.obj["app"]; !ok || commandObject.obj["app"].str == "" {
    panic("Connect message has no app name");
  }
  appName := commandObject.obj["app"].str
  log.Println("Connect message for app:", appName)

  log.Println("Command Object: ", commandObject)

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

  result := []AMFObj {
    AMFObj { atype : AMF_STRING, str : "_result", },
    AMFObj { atype : AMF_NUMBER, f64 : txnId.f64, },
    AMFObj { atype : AMF_OBJECT,
      obj : map[string] AMFObj {
        "flashVer" : AMFObj { atype : AMF_STRING, str : "FMS/3,0,1,123", },
        "capabilities" : AMFObj { atype : AMF_NUMBER, f64 : 31, },
      },
    },
    AMFObj { atype : AMF_OBJECT,
      obj : map[string] AMFObj {
        "level" : AMFObj { atype : AMF_STRING, str : "status", },
        "code" : AMFObj { atype : AMF_STRING, str : "NetConnection.Connect.Success", },
        "description" : AMFObj { atype : AMF_STRING, str : "Connection Success.", },
        "objectEncoding" : AMFObj { atype : AMF_NUMBER, f64 : 0, },
      },
    },
  }

  var b bytes.Buffer
  for _, v := range result {
    WriteAMF(&b, v)
  }
  msg = NewControlMessage(MSG_AMF_CMD, b.Bytes())

  m.outChan <- msg

}

























