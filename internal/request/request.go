package request

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/Robot-tim1/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	RequestLineParse int = iota // 0
	HeaderParse                 // 1
	BodyParse                   // 2
	Done                        // 3
)

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := Request{State: RequestLineParse, Headers: headers.NewHeaders()}
	buf := make([]byte, bufferSize)
	readToIndex := 0

	for request.State != Done {
		if readToIndex >= cap(buf) {
			newBuf := make([]byte, cap(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}
		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.State == HeaderParse {
					return nil, errors.New("error missing end of headers")
				}
				if request.State == BodyParse {
					return nil, errors.New("error body length shorter than stated content=length")
				}
				request.State = Done
				break
			}
			return nil, err
		}
		if request.State == HeaderParse && len(buf) > 8192 && !bytes.Contains(buf, []byte("\r\n\r\n")) {
			return nil, errors.New("error Header Fields Too Large")
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
	totalBytesParsed := 0
	for r.State != Done {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return totalBytesParsed, nil
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case RequestLineParse:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = requestLine
		r.State = HeaderParse
		return n, nil
	case HeaderParse:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if !done {
			return n, nil
		}
		r.State = BodyParse
		return n, nil
	case BodyParse:
		contentLength := r.Headers.Get("Content-length")
		if contentLength == "" {
			r.State = Done
			return 0, nil
		}
		if contentLength == "0" {
			r.State = Done
			return 0, nil
		}
		if len(data) == 0 {
			return 0, nil
		}
		length, err := strconv.Atoi(contentLength)
		if err != nil {
			return 0, err
		}
		r.Body = append(r.Body, data...)
		if len(r.Body) > length {
			return 0, errors.New("error body length longer than stated content-length")
		}
		if len(r.Body) == length {
			r.State = Done
		}
		return len(data), nil
	case Done:
		return 0, errors.New("error: trying to read data in a done State")
	default:
		return 0, errors.New("error: unknown State")
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

	_, httpVersion, _ := strings.Cut(parts[2], "/")
	if !strings.Contains(httpVersion, "1.1") {
		return RequestLine{}, 0, errors.New("http version is not 1.1")
	}

	requestLine := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   httpVersion,
	}

	return requestLine, len(requestParts[0]) + 2, nil
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
