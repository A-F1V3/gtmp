package main

import (
	"io"
	"log"
	"net"
)

func handleConnection(conn io.ReadWriteCloser) {
	defer conn.Close()

	done := make(chan int)
	handShake(conn)
	startNewSession(conn, done)
	<-done
}

func main() {
	log.Println("Starting RTMP Server")
	lstnr, err := net.Listen("tcp", ":1935")
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
		go handleConnection(connection)
	}

}
