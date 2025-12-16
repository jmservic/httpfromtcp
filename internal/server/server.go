package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

// Server is an HTTP 1.1 server
type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	socket, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("Error creating listener: %s", err)
	}

	s := &Server{
		listener: socket,
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
	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!\n"))
	if err != nil {
		fmt.Printf("Error serving http request: %s", err)
		return
	}
}
