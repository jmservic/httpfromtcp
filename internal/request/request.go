package request

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	requestLine, err := parseRequestLine(data)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *requestLine,
	}, nil
}

func parseRequestLine(data []byte) (*RequestLine, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, errors.New("could not find CRLF in request-line")
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, err
	}
	return requestLine, nil
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
