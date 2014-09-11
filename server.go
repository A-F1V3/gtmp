package main

import (
	"log"
	"net"
)

const (
	MAX_CONNECITONS = 2
)

type Server struct {
	Applications map[string]*Application
	Address      string
}

func (server *Server) Start(done chan int) {
	defer func() { done <- 1 }()

	lstnr, err := net.Listen("tcp", server.Address)
	if err != nil {
		log.Println("Socket listen err: ", err)
		return
	}

	if len(server.Applications) < 1 {
		log.Println("Server has no applications configured")
		return
	}

	for _, app := range server.Applications {
		go app.Start()
	}

	sessionLimit := make(chan int, MAX_CONNECITONS)
	for {
		sessionLimit <- 1
		connection, err := lstnr.Accept()
		log.Println("New Connection from: ", connection.RemoteAddr().String())
		if err != nil {
			log.Println("New connection err: ", err)
			<-sessionLimit
			continue
		}

		go func() {
			done := make(chan int)
			session := Session{conn: connection, Server: server}
			go session.Start(done)
			<-done
			<-sessionLimit
		}()

	}
}
