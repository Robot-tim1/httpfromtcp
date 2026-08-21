package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("error resolving udp address: %v", err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("error dialling udp address: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		input, err := reader.ReadString(0x0A)
		if err != nil {
			log.Printf("error reading from stdin: %v", err)
			continue
		}
		_, err = conn.Write([]byte(input))
		if err != nil {
			log.Printf("error writing to connection: %v", err)
		}
	}
}
