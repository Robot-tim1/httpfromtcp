package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Robot-tim1/httpfromtcp/internal/headers"
	"github.com/Robot-tim1/httpfromtcp/internal/request"
	"github.com/Robot-tim1/httpfromtcp/internal/response"
	"github.com/Robot-tim1/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, videoHandler)
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

func videoHandler(w *response.Writer, r *request.Request) {
	if r.RequestLine.RequestTarget == "/video" {
		fileData, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			log.Fatalln(err)
		}
		err = w.WriteStatusLine(response.StatusOk)
		if err != nil {
			log.Fatalln(err)
		}
		h := response.GetDefaultHeaders(len(fileData))
		h.AddHeader("content-type", "video/mp4")
		err = w.WriteHeaders(h)
		if err != nil {
			log.Fatalln(err)
		}
		_, err = w.WriteBody(fileData)
		if err != nil {
			log.Fatalln(err)
		}
		return
	}
	w.WriteStatusLine(response.StatusOk)
	w.WriteHeaders(response.GetDefaultHeaders(0))
}

func testHandler(w *response.Writer, r *request.Request) {
	if strings.HasPrefix(r.RequestLine.RequestTarget, "/yourproblem") {
		w.WriteStatusLine(response.StatusBadRequest)
		body := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
		h := response.GetDefaultHeaders(len(body))
		h.AddHeader("Content-Type", "text/html")
		w.WriteHeaders(h)
		w.WriteBody([]byte(body))
		return
	}
	if strings.HasPrefix(r.RequestLine.RequestTarget, "/myproblem") {
		w.WriteStatusLine(response.StatusServerError)
		body := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
		h := response.GetDefaultHeaders(len(body))
		h.AddHeader("Content-Type", "text/html")
		w.WriteHeaders(h)
		w.WriteBody([]byte(body))
		return
	}
	w.WriteStatusLine(response.StatusOk)
	body := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
	h := response.GetDefaultHeaders(len(body))
	h.AddHeader("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody([]byte(body))
}

func handler(w *response.Writer, r *request.Request) {
	if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin/") {
		requestPath := strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin/")
		resp, err := http.Get("https://httpbingo.org/" + requestPath)
		if err != nil {
			HErr := server.HandlerError{Code: response.StatusBadRequest, Message: err.Error()}
			HErr.Write(w)
			return
		}
		err = w.WriteStatusLine(response.StatusOk)
		if err != nil {
			log.Fatalln(err)
		}
		h := response.GetDefaultHeaders(0)
		h.DeleteHeader("Content-length")
		h.AddHeader("Transfer-Encoding", "chunked")
		h.AddHeader("Trailer", "X-Content-SHA256, X-Content-Length")
		w.WriteHeaders(h)
		buf := make([]byte, 1024)
		fullResponse := make([]byte, 0, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if err != nil {
				if err == io.EOF {
					_, err = w.WriteChunkedBody(buf[:n])
					if err != nil {
						log.Fatalln(err)
					}
					fullResponse = append(fullResponse, buf[:n]...)

					w.WriteChunkedBodyDone()
					trailers := headers.NewHeaders()
					hash := sha256.Sum256(fullResponse)
					trailers.AddHeader("X-Content-SHA256", hex.EncodeToString(hash[:]))
					trailers.AddHeader("X-Content-Length", fmt.Sprint(len(fullResponse)))
					err = w.WriteTrailers(trailers)
					if err != nil {
						log.Fatalln(err)
					}
					break
				}
				log.Fatalln(err)
			}
			if n == 0 {
				continue
			}
			_, err = w.WriteChunkedBody(buf[:n])
			if err != nil {
				log.Fatalln(err)
			}
			fullResponse = append(fullResponse, buf[:n]...)
		}
		return

	}
	w.WriteStatusLine(response.StatusOk)
	w.WriteHeaders(response.GetDefaultHeaders(0))
}
