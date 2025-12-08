package request

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := Request{}
	buffer := make([]byte, bufferSize)
	readToIndex := 0
	for req.state != requestStateDone {
		if readToIndex == len(buffer) {
			newBuffer := make([]byte, 2*len(buffer))
			copy(newBuffer, buffer)
			buffer = newBuffer

		}

		numBytesRead, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.state = requestStateDone
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buffer[:readToIndex])
		if err != nil {
			return &req, err
		}

		copy(buffer, buffer[numBytesParsed:])
		readToIndex -= numBytesParsed
	}

	return &req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, idx, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, errors.New("Too many space separated parts in the request line.")
	}

	method := parts[0]

	if strings.ContainsFunc(method, func(r rune) bool { return r < 'A' || r > 'Z' }) {
		return nil, errors.New("Method contains non capital alphabetic characters.")
	}

	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 {
		return nil, errors.New("Invalid HTTP version format")
	}

	version := httpParts[1]
	if version != "1.1" {
		return nil, errors.New("Unsupported HTTP version")
	}

	return &RequestLine{
		HttpVersion:   version,
		RequestTarget: parts[1],
		Method:        method,
	}, nil
}

func (r *Request) parse(data []byte) (int, error) {
	requestLine, bytesRead, err := parseRequestLine(data)
	if err != nil {
		return bytesRead, err
	}

	if requestLine != nil {
		r.RequestLine = *requestLine
		r.state = requestStateDone
	} else {
		r.state = requestStateInitialized
	}

	return bytesRead, nil
}
