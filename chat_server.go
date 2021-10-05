package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client struct {
	clientName string
	remoteAddr string
	clientCnan chan<- string
}

var (
	leaving  = make(chan client)
	entering = make(chan client)
	messages = make(chan string)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli.clientCnan <- msg
			}
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli.clientCnan)
		}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	client := client{
		clientCnan: ch,
		remoteAddr: conn.RemoteAddr().String(),
	}

	ch <- "Hello! Enter your name"
	clientNameScanner := bufio.NewScanner(conn)
	clientNameScanner.Scan()
	client.clientName = clientNameScanner.Text()

	entering <- client
	ch <- "You are " + client.clientName + " with IP address: " + client.remoteAddr
	messages <- client.clientName + " has arrived"

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- client.clientName + ": " + input.Text()
	}
	leaving <- client
	messages <- client.clientName + " has left"
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
