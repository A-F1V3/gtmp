package main

import (
  "log"
  "net"
)

var (
  MAX_CONNECITONS = 2
)

type Server struct {
  apps map[string]*Application
  address string
}

func NewServer(address string, done chan int) *Server{
  server := Server{apps: make(map[string]*Application), address: address}

  go func() {
    lstnr, err := net.Listen("tcp", server.address)
    if err != nil {
      log.Println("Socket listen err: ", err)
      return
    }

    sessionLimit := make(chan int, MAX_CONNECITONS)
    for {
      sessionLimit <- 1
      connection, err := lstnr.Accept()
      log.Println("New Connection from: ", connection.RemoteAddr().String())
      if err != nil {
        log.Println("New connection err: ", err)
        continue
      }
      go NewSession(connection, server, sessionLimit)
    }
    done <- 1
  }()

  return &server
}
