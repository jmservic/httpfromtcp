package response

import (
	"fmt"
	"github.com/jmservic/httpfromtcp/internal/headers"
	"io"
	"net"
	"strconv"
)

type StatusCode int

const (
	StatusOK                  = 200
	StatusBadRequest          = 400
	StatusInternalServerError = 500
)

type writerState int

const (
	writerStateStatusLine = iota
	writerStateHeaders
	writerStateBody
	writerStateTrailers
	writerStateComplete
)

type Writer struct {
	conn  net.Conn
	state writerState
}

func NewWriter(conn net.Conn) *Writer {
	return &Writer{
		conn:  conn,
		state: writerStateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerStateStatusLine {
		return fmt.Errorf("Attempted to write status line in the wrong state.")
	}
	err := WriteStatusLine(w.conn, statusCode)
	if err != nil {
		return fmt.Errorf("Error writing the HTTP Status line: %w", err)
	}
	w.state = writerStateHeaders
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writerStateHeaders {
		return fmt.Errorf("Attempted to write headers line in the wrong state.")
	}
	err := WriteHeaders(w.conn, headers)
	if err != nil {
		return fmt.Errorf("Error writing HTTP headers: %w", err)
	}
	w.state = writerStateBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.state)
	}
	w.state = writerStateComplete
	return w.conn.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.state)
	}
	writtenBytes, err := fmt.Fprintf(w.conn, "%X\r\n", len(p))
	if err != nil {
		return writtenBytes, err
	}
	n, err := w.conn.Write(p)
	writtenBytes += n
	if err != nil {
		return writtenBytes, err
	}
	n, err = w.conn.Write([]byte("\r\n"))
	if err != nil {
		return writtenBytes, err
	}
	writtenBytes += n
	return writtenBytes, nil
}

func (w *Writer) WriteChunkedBodyDone(hasTrailers bool) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.state)
	}

	line := []byte("0\r\n\r\n")
	var nextState writerState = writerStateComplete

	if hasTrailers {
		line = []byte("0\r\n")
		nextState = writerStateTrailers
	}
	n, err := w.conn.Write(line)
	if err != nil {
		return n, err
	}

	w.state = nextState
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != writerStateTrailers {
		return fmt.Errorf("cannot write trailers in state %d", w.state)
	}
	err := WriteHeaders(w.conn, h)
	if err != nil {
		return fmt.Errorf("Error writing HTTP trailers: %w", err)
	}

	w.state = writerStateComplete
	return nil
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var err error

	switch statusCode {
	case StatusOK:
		_, err = w.Write([]byte("HTTP/1.1 200 OK \r\n"))
	case StatusBadRequest:
		_, err = w.Write([]byte("HTTP/1.1 400 Bad Request \r\n"))
	case StatusInternalServerError:
		_, err = w.Write([]byte("HTTP/1.1 500 Internal Server Error \r\n"))
	default:
		_, err = fmt.Fprintf(w, "HTTP/1.1 %v \r\n", statusCode)
	}

	return err
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, val := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, val)
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	header := headers.NewHeaders()
	header.Set("Content-Length", strconv.Itoa(contentLen))
	header.Set("Connection", "close")
	header.Set("Content-Type", "text/plain")
	return header
}
