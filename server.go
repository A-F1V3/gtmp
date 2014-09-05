package main

import (
  "log"
  "net"
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
    }

    for {
      connection, err := lstnr.Accept()
      log.Println("New Connection from: ", connection.RemoteAddr().String())
      if err != nil {
        log.Println("New connection err: ", err)
        continue
      }
      fin := make(chan int)
      go NewSession(connection, server, fin)  //handleConnection(connection)
    }
    done <- 1
  }()

  return &server
}

// func handleConnection(conn io.ReadWriteCloser) {
//   defer func() {
//     log.Println("Closing Connections")
//     conn.Close()
//   }()

//   done := make(chan int)
//   handShake(conn)
//   NewSession(conn, done)
//   <-done
// }
