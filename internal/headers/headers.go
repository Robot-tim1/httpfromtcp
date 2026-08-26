package headers

import (
	"bytes"
	"errors"
	"log"
	"regexp"
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
		return 0, false, errors.New("error: invalid header format")
	}

	header = strings.ToLower(header)
	header = strings.ReplaceAll(header, " ", "")
	headerParts := strings.SplitN(header, ":", 2)

	match, err := regexp.MatchString("([^A-Za-z0-9!#-'+*"+regexp.QuoteMeta("-")+".^_`|~])", headerParts[0])
	if err != nil {
		log.Fatalf("error regex failure: %v", err)
	}
	if match {
		return 0, false, errors.New("error header has invalid character")
	}

	h[headerParts[0]] = headerParts[1]
	return len(parts[0]) + 2, false, nil
}
