package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

var connections = map[net.Conn]struct{}{}

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	go broadcast()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(c net.Conn) {
	defer c.Close()
	defer delete(connections, c)

	connections[c] = struct{}{}
	for {
		_, err := io.WriteString(c, time.Now().Format("15:04:05\n\r"))
		if err != nil {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func broadcast() {
	for {
		var broadCastMessage string
		_, err := fmt.Scanln(&broadCastMessage)
		if err != nil {
			log.Println(err)
			continue
		}

		sendBroadcastMessage(broadCastMessage)
	}
}

func sendBroadcastMessage(message string) {
	for c := range connections {
		_, err := io.WriteString(c, message)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
