package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println(1)
	go func() {
		fmt.Println(2)
		io.Copy(os.Stdout, conn)
	}()
	fmt.Println(3)
	io.Copy(conn, os.Stdin) // until you send ^Z
	fmt.Println(4)
	fmt.Printf("%s: exit", conn.LocalAddr())
}
