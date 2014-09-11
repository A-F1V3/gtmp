package main

import (
	"log"
)

type Application struct {
	Name       string
	Streams    map[string]*Stream
	reqChannel chan *AppRequest
}

type AppRequest struct {
	streamName string
	f          func(string) *Stream
	resChan    chan *Stream
}

func (app *Application) Start() {
	log.Println("Starting app:", app.Name)
	app.reqChannel = make(chan *AppRequest)
	for req := range app.reqChannel {
		app.handleRequest(req)
	}
}

func (app *Application) handleRequest(req *AppRequest) {
	req.resChan <- req.f(req.streamName)
}

func (app *Application) Publish(streamName string) (*Stream, error) {
	req := AppRequest{
		streamName,
		func(sn string) *Stream {
			_, ok := app.Streams[streamName]
			if !ok {
				log.Println("New Stream:", streamName)
				return &Stream{name: streamName}
			} else {
				return nil
			}
		},
		make(chan *Stream),
	}
	app.reqChannel <- &req

	var err error
	var stream *Stream
	stream = <-req.resChan
	return stream, err
}
