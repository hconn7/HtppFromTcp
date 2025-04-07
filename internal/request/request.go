package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/hconn7/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine    RequestLine
	Headers        headers.Headers
	Body           []byte
	State          requestState
	bodyLengthRead int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	StateInit = iota
	StateHeaders
	StateBody
	StateDone
)

type requestState int

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.State != StateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}

	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case StateInit:
		req, parsed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if parsed == 0 {
			return 0, nil
		}
		r.RequestLine = *req
		r.State = StateHeaders
		return parsed, nil
	case StateHeaders:
		n, done, err := r.Headers.ParseHeaders(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.State = StateBody
		}
		return n, nil
	case StateBody:
		contentLenStr, ok := r.Headers.Get("Content-Length")
		if !ok {
			r.State = StateDone
			return len(data), nil
		}
		contentLen, err := strconv.Atoi(contentLenStr)
		if err != nil {
			return 0, fmt.Errorf("malformed Content-Length: %s", err)
		}
		r.Body = append(r.Body, data...)
		r.bodyLengthRead += len(data)
		if r.bodyLengthRead > contentLen {
			return 0, fmt.Errorf("Content-Length too large")
		}
		if r.bodyLengthRead == contentLen {
			r.State = StateDone
		}
		return len(data), nil

	}
	return 0, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	buf := make([]byte, 8)
	readToIndex := 0

	req := Request{
		State:   StateInit,
		Headers: headers.NewHeaders(),
	}

	for req.State != StateDone {
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.State != StateDone {
					return nil, errors.New("state not done")
				}

			}
			return &Request{}, err
		}
		readToIndex += n
		parsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return &Request{}, err
		}
		if parsed > 0 {
			copy(buf, buf[parsed:readToIndex])
			readToIndex -= parsed
		}
	}
	return &req, nil
}
func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return &RequestLine{}, 0, nil
	}
	s := data[:idx]
	stringDat := string(s)

	requestSplit := strings.Split(stringDat, " ")
	if len(requestSplit) != 3 {
		return nil, 0, errors.New("Malformed format")
	}
	reqMethod := strings.TrimSpace(requestSplit[0])
	reqTarget := requestSplit[1]
	reqVersion := requestSplit[2]

	versionNum := strings.Split(reqVersion, "/")
	if versionNum[1] != "1.1" {
		return &RequestLine{}, 0, errors.New("Version is not 1.1")
	}
	validMethods := []string{"GET", "POST", "PUT", "DELETE"}
	valid := false
	for _, v := range validMethods {
		if reqMethod == v {
			valid = true
			break
		}
	}
	if !valid {
		return &RequestLine{}, 0, fmt.Errorf("Method is invalid: %s", reqMethod)

	}

	return &RequestLine{
		HttpVersion:   versionNum[1],
		RequestTarget: reqTarget,
		Method:        reqMethod}, idx + 2, nil

}
