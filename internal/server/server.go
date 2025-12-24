package server

import (
	"fmt"
	"github.com/jmservic/httpfromtcp/internal/request"
	"github.com/jmservic/httpfromtcp/internal/response"
	"io"
	"log"
	"net"
	"sync/atomic"
)

// Server is an HTTP 1.1 server
type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (h HandlerError) Write(w io.Writer) error {
	_, err := fmt.Fprintf(w, "HTTP/1.1 %d \r\n\r\n%s", h.StatusCode, h.Message)
	return err
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	socket, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("Error creating listener: %s", err)
	}

	s := &Server{
		listener: socket,
		handler:  handler,
	}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		response.WriteStatusLine(conn, response.StatusBadRequest)
		return
	}

	writer := response.NewWriter(conn)
	s.handler(writer, req)
}
