package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("could not start listener: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error getting connection: %v", err)
		}
		fmt.Println("connection accepted")

		stringChannel := getLinesChannel(conn)

		for line := range stringChannel {
			fmt.Printf("%s\n", line)
		}
		conn.Close()
		fmt.Println("connection closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	stringChannel := make(chan string)

	go func() {
		currentLine := ""
		buf := make([]byte, 8)
		for {
			n, err := f.Read(buf)
			switch err {
			case io.EOF:
				if currentLine != "" {
					stringChannel <- currentLine
				}
				close(stringChannel)
				return
			case nil:
				tempString := string(buf[:n])
				if strings.Contains(tempString, "\n") {
					parts := strings.Split(tempString, "\n")
					for _, part := range parts[:len(parts)-1] {
						currentLine += part
					}
					stringChannel <- currentLine
					currentLine = parts[len(parts)-1]
				} else {
					currentLine += tempString
				}
			}
		}
	}()

	return stringChannel
}
