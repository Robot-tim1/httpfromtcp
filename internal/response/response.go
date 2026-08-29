package response

import (
	"errors"
	"fmt"
	"io"

	"github.com/Robot-tim1/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOk          StatusCode = 200
	StatusBadRequest  StatusCode = 400
	StatusServerError StatusCode = 500
)

type WriterStatus int

const (
	writeStatusLine WriterStatus = iota
	writeHeaders
	writeBody
)

type Writer struct {
	writer      io.Writer
	writerState WriterStatus
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w, writerState: -1}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writeStatusLine {
		return errors.New("error write out of order")
	}
	var err error
	switch statusCode {
	case StatusOk:
		_, err = w.writer.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case StatusBadRequest:
		_, err = w.writer.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case StatusServerError:
		_, err = w.writer.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	}
	w.writerState = writeHeaders
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["content-length"] = fmt.Sprint(contentLen)
	h["connection"] = "close"
	h["content-type"] = "text/plain"
	return h
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writeHeaders {
		return errors.New("error write out of order")
	}
	for k, v := range headers {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	w.writer.Write([]byte("\r\n"))
	w.writerState = writeBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writeBody {
		return 0, errors.New("error write out of order")
	}
	return w.writer.Write(p)
}
