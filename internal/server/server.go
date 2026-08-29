package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/Robot-tim1/httpfromtcp/internal/request"
	"github.com/Robot-tim1/httpfromtcp/internal/response"
)

type Server struct {
	listener       net.Listener
	closed         atomic.Bool
	handler        Handler
	handlerTracker *sync.WaitGroup
	listenerDone   chan struct{}
}

type HandlerError struct {
	Code    response.StatusCode
	Message string
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, handler Handler) (*Server, error) {
	l, err := net.Listen("tcp", ":"+fmt.Sprint(port))
	if err != nil {
		return nil, err
	}
	server := Server{listener: l, closed: atomic.Bool{}, handler: handler, handlerTracker: &sync.WaitGroup{}, listenerDone: make(chan struct{})}
	go server.listen()
	return &server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	err := s.listener.Close()
	<-s.listenerDone
	s.handlerTracker.Wait()
	return err
}

func (s *Server) listen() {
	defer close(s.listenerDone)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("error getting connection: %v", err)
			continue
		}
		s.handlerTracker.Add(1)
		go s.handle(conn)
	}
}

func (h *HandlerError) Write(w *response.Writer) {
	err := w.WriteStatusLine(h.Code)
	if err != nil {
		log.Printf("error writing status line for error response: %v", err)
		return
	}
	header := response.GetDefaultHeaders(len(h.Message))
	err = w.WriteHeaders(header)
	if err != nil {
		log.Printf("error writing headers for error response: %v", err)
		return
	}
	w.WriteBody([]byte(h.Message))
}

func (s *Server) handle(conn net.Conn) {
	defer s.handlerTracker.Done()
	defer conn.Close()

	w := response.NewWriter(conn)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := HandlerError{Code: response.StatusBadRequest, Message: err.Error()}
		hErr.Write(w)
		return
	}

	s.handler(w, req)
}
