package request

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	Initialized int = iota // 0
	Done                   // 1
)

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := Request{state: Initialized}
	buf := make([]byte, bufferSize)
	readToIndex := 0

	for request.state != Done {
		if len(buf) == cap(buf) {
			newBuf := make([]byte, cap(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}
		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.state = Done
				break
			}
			return nil, err
		}
		readToIndex += n
		m, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		if m == 0 {
			continue
		}
		newBuf := make([]byte, cap(buf))
		copy(newBuf, buf[m:])
		buf = newBuf
		readToIndex -= m
	}

	return &request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case Initialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = requestLine
		r.state = Done
		return n, nil
	case Done:
		return 0, errors.New("error: trying to read data in a done state")
	default:
		return 0, errors.New("error: unknown state")
	}
}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	requestParts := bytes.Split(data, []byte("\r\n"))
	if len(requestParts) == 1 {
		return RequestLine{}, 0, nil
	}

	rawRequestline := string(requestParts[0])
	parts := strings.Split(rawRequestline, " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, errors.New("invalid number of parts in request line")
	}

	if !isAllUpper(parts[0]) {
		return RequestLine{}, 0, errors.New("method is invalid")
	}

	httpVersion := strings.Split(parts[2], "/")[1]
	if !strings.Contains(httpVersion, "1.1") {
		return RequestLine{}, 0, errors.New("http version is not 1.1")
	}

	requestLine := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   httpVersion,
	}

	return requestLine, len(requestParts[0]), nil
}

func isAllUpper(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, letter := range s {
		if letter < 'A' || letter > 'Z' {
			return false
		}
	}
	return true
}
