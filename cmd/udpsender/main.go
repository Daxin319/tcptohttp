package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	raddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("error resolving udp address: %v\n", err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatalf("error dialing UDP address: %v\n", err)
	}
	defer conn.Close()

	input := bufio.NewReader(os.Stdin)
	for true {
		fmt.Printf("> ")
		str, err := input.ReadString(10)
		if err != nil {
			log.Fatalf("error reading input string: %v\n", err)
		}
		_, err = conn.Write([]byte(str))
		if err != nil {
			log.Fatalf("error writing udp connection: %v\n", err)
		}
	}
}
