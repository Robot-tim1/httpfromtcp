package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Robot-tim1/httpfromtcp/internal/request"
	"github.com/Robot-tim1/httpfromtcp/internal/response"
	"github.com/Robot-tim1/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, r *request.Request) {
	body := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
	var err error
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		body = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
		err = w.WriteStatusLine(response.StatusBadRequest)
	case "/myproblem":
		body = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
		err = w.WriteStatusLine(response.StatusServerError)
	default:
		err = w.WriteStatusLine(response.StatusOk)
	}
	if err != nil {
		log.Printf("error writing status line: %v", err)
		return
	}
	h := response.GetDefaultHeaders(len(body))
	h.AddHeader("Content-Type", "text/html")
	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("error writing headers: %v", err)
		return
	}
	_, err = w.WriteBody([]byte(body))
	if err != nil {
		log.Printf("error writing body: %v", err)
	}
}
