package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type Headers map[string]string

var ctrl = []byte("\r\n")

func NewHeaders() Headers {
	headers := make(map[string]string)
	return headers
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	v, ok := h[key]
	return v, ok
}
func (h Headers) ParseHeaders(data []byte) (n int, done bool, err error) {
	readIndex := 0

	idx := bytes.Index(data[readIndex:], ctrl)
	if idx == -1 {
		return readIndex, false, nil
	}

	if idx == 0 {
		return readIndex + 2, true, nil
	}

	line := string(data[readIndex : readIndex+idx])
	splitByCol := strings.SplitN(line, ":", 2)
	if len(splitByCol) < 2 {
		return 0, false, errors.New("malformed header line: missing ':'")
	}

	key := strings.ToLower(splitByCol[0])
	if key != strings.TrimRight(key, " ") {
		return 0, false, errors.New("Invalid header")
	}

	value := strings.TrimSpace(splitByCol[1])

	key = strings.TrimSpace(key)
	if !validTokens([]byte(key)) {
		return 0, false, fmt.Errorf("invalid header key: %s", key)
	}

	if v, ok := h[key]; ok {
		value = v + ", " + value
	}
	h[key] = value

	readIndex += idx + 2
	return readIndex, false, nil

}
func validTokens(data []byte) bool {
	for _, c := range data {
		if !(c >= 'A' && c <= 'Z' ||
			c >= 'a' && c <= 'z' ||
			c >= '0' && c <= '9' ||
			c == '-') {
			return false
		}
	}
	return true
}
