package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
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
		fullResponse := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					w.WriteChunkedBodyDone()
					trailers := headers.NewHeaders()
					hash := sha256.Sum256(fullResponse)
					trailers.AddHeader("X-Content-SHA256", hex.EncodeToString(hash[:]))
					trailers.AddHeader("X-Content-Length", fmt.Sprint(len(hash)))
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
