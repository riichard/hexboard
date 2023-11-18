package tcpserver

import (
	"log"
	"net"
	"os"
	"time"
)

const (
	HOST = "0.0.0.0"
	PORT = "8080"
	TYPE = "tcp"
)

func Start(ch chan<- byte) {

	defer close(ch)

	listen, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	// close listener
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		go handleRequest(conn, ch)
	}
}

func handleRequest(conn net.Conn, ch chan<- byte) {

	// incoming request
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range buffer {
		if v != 0 {
			// log.Default().Printf("(%v) ", v)
			ch <- v
			time.Sleep(30 * time.Millisecond)
		}
	}

	// close conn
	conn.Close()
}
