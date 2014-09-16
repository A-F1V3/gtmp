package main

import (
	"errors"
	"log"
)

type Application struct {
	Name       string
	streams    map[string]*Stream
	reqChannel chan *AppRequest
	Callbacks  map[string]interface{}
}

type AppRequest struct {
	streamName string
	f          func(string) *Stream
	resChan    chan *Stream
}

func (app *Application) Start() {
	log.Println("Starting app:", app.Name)
	log.Println("Starting app:", app)

	for cb, v := range app.Callbacks {
		app.Callbacks[cb], _ = GetCallbacks(v)
	}
	log.Println(app.Callbacks)

	app.reqChannel = make(chan *AppRequest)
	app.streams = make(map[string]*Stream)
	for req := range app.reqChannel {
		app.handleRequest(req)
	}
}

func (app *Application) handleRequest(req *AppRequest) {
	req.resChan <- req.f(req.streamName)
}

func (app *Application) Publish(streamName string) (*Stream, error) {
	err := ExecuteCallbacks(app.Callbacks["publish"])
	if err != nil {
		log.Println("Publish CB failed", err)
	}

	req := AppRequest{
		streamName,
		func(sn string) *Stream {
			_, ok := app.streams[streamName]
			if !ok {
				log.Println("New Stream:", streamName)
				newStream := NewStream(streamName)
				go newStream.StartStream()
				app.streams[streamName] = newStream
				return newStream
			} else {
				return nil
			}
		},
		make(chan *Stream),
	}
	app.reqChannel <- &req

	var stream *Stream
	stream = <-req.resChan
	if stream == nil {
		err = errors.New("None for you")
	}
	return stream, err
}

func (app *Application) Play(streamName string) (*Stream, error) {
	err := ExecuteCallbacks(app.Callbacks["play"])
	if err != nil {
		log.Println("Publish CB failed", err)
	}
	req := AppRequest{
		streamName,
		func(sn string) *Stream {
			stream, _ := app.streams[streamName]
			return stream
		},
		make(chan *Stream),
	}
	app.reqChannel <- &req

	var stream *Stream
	stream = <-req.resChan
	if stream == nil {
		err = errors.New("Stream not found")
	}
	return stream, err

}
