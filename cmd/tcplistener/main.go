package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	conn, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error listening for connection: %v\n", err)
	}
	defer conn.Close()
	t := true
	for t {
		data, err := conn.Accept()
		if err != nil {
			log.Fatalf("error accepting connection: %s\n", err)
		}
		if data != nil {
			fmt.Println("Connection Accepted")
			lines := getLinesChannel(data)
			for line := range lines {
				fmt.Printf("%s\n", line)
			}
			fmt.Println("Channel closed")
		}
	}
}
