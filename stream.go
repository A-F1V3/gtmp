package main

type Stream struct {
	name      string
	publisher *MessageStream
	players   []*MessageStream
}
