package gtmp

type Stream struct {
	name      string
	publisher *MessageStream
	players   []*MessageStream
}
