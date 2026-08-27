package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Robot-tim1/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("could not start listener: %v\n", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error getting connection: %v\n", err)
			continue
		}
		fmt.Println("connection accepted")

		r, err := request.RequestFromReader(conn)
		conn.Close()
		if err != nil {
			log.Printf("error reading from connection: %v\n", err)
			continue
		}
		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		for k, v := range r.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Printf("Body:\n")
		fmt.Println(string(r.Body))
		fmt.Println("connection closed")
	}
}
