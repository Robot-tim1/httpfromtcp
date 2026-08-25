package headers

import (
	"bytes"
	"errors"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if bytes.HasPrefix(data, []byte("\r\n")) {
		return 2, true, nil
	}

	parts := bytes.Split(data, []byte("\r\n"))
	if len(parts) == 1 {
		return 0, false, nil
	}

	header := string(parts[0])
	if !strings.Contains(strings.Split(header, " ")[0], ":") {
		return 0, false, errors.New("error: invalied header format")
	}

	header = strings.ReplaceAll(header, " ", "")
	headerParts := strings.SplitN(header, ":", 2)
	h[headerParts[0]] = headerParts[1]
	return len(parts[0]) + 2, false, nil
}
