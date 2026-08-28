package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type Server struct {
	Listener       net.Listener
	Closed         atomic.Bool
	handlerTracker *sync.WaitGroup
	listenerDone   chan struct{}
}

func Serve(port int) (*Server, error) {
	l, err := net.Listen("tcp", ":"+fmt.Sprint(port))
	if err != nil {
		return nil, err
	}
	server := Server{Listener: l, Closed: atomic.Bool{}, handlerTracker: &sync.WaitGroup{}, listenerDone: make(chan struct{})}
	go server.listen()
	return &server, nil
}

func (s *Server) Close() error {
	s.Closed.Store(true)
	err := s.Listener.Close()
	<-s.listenerDone
	s.handlerTracker.Wait()
	return err
}

func (s *Server) listen() {
	defer close(s.listenerDone)
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.Closed.Load() {
				return
			}
			log.Printf("error getting connection: %v\n", err)
			continue
		}
		s.handlerTracker.Add(1)
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer s.handlerTracker.Done()
	defer conn.Close()
	response := `HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 13

Hello World!
`
	_, err := conn.Write([]byte(response))
	if err != nil && s.Closed.Load() != true {
		log.Printf("error writing response to connection: %v\n", err)
	}

}
