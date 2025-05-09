package main

import (
	"fmt"
	"log"
	"main/internal/request"
	"net"
)

func main() {
	conn, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error listening for connection: %v\n", err)
	}
	defer conn.Close()
	for {
		data, err := conn.Accept()
		if err != nil {
			log.Fatalf("error accepting connection: %s\n", err)
		}
		if data != nil {
			fmt.Printf("Connection Accepted\n\n")

			req, err := request.RequestFromReader(data)
			if err != nil {
				log.Fatalf("error reading data: %v\n", err)
			}

			fmt.Printf("Request line:\n")
			fmt.Printf("- Method: %s\n", req.RequestLine.Method)
			fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
			fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
			fmt.Println("Headers: ")
			for key := range req.Headers {
				fmt.Printf("- %s: %s\n", key, req.Headers[key])
			}
			fmt.Println("")
			fmt.Printf("Channel closed\n\n")
		}
	}
}
