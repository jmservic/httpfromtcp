package request

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jmservic/httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
)

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
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
	req := Request{
		state:   requestStateInitialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}
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
				switch req.state {
				case requestStateParsingHeaders:
					return nil, errors.New("missing end of headers")
				case requestStateParsingBody:
					return nil, errors.New("Received partial body content")
				}
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
	return requestLine, idx + 2, nil
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
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if n == 0 || err != nil {
			return totalBytesParsed + n, err
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, bytesRead, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if requestLine != nil {
			r.RequestLine = *requestLine
			r.state = requestStateParsingHeaders
		}
		return bytesRead, nil
	case requestStateParsingHeaders:
		bytesRead, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return bytesRead, nil
	case requestStateParsingBody:
		contentLengthStr, ok := r.Headers.Get("Content-Length")
		if !ok {
			// assume that if no content-length header is present, there is no body
			r.state = requestStateDone
			return 0, nil
		}
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("malformed Content-Length: %s", err)
		}

		r.Body = append(r.Body, data...)
		bodyLen := len(r.Body)
		if bodyLen > contentLength {
			return len(data), errors.New("error: body length is greater thancontent length")
		}
		if bodyLen == contentLength {
			r.state = requestStateDone
		}
		return len(data), nil
	case requestStateDone:
		return 0, errors.New("error: trying to read data in a done state")
	default:
		return 0, errors.New("unknown state")
	}
}
