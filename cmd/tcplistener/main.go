package main

import (
	"fmt"
	//"os"
	"github.com/jmservic/httpfromtcp/internal/request"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	/*	file, err := os.Open("messages.txt")
		if err != nil {
			log.Fatalf("Error opening messages.txt: %v", err)
		}
		ch := getLinesChannel(file) */
	socket, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		log.Fatalf("Error opening a TCP socket on port 42069: %v", err)
	}
	defer socket.Close()

	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Fatalf("Error establishing a connection: %v", err)
		}
		fmt.Println("A connection has been established.")
		req, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Printf("Bad HTTP Request!: %s", err)
			conn.Close()
			continue
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, val := range req.Headers {
			fmt.Printf("- %s: %s\n", key, val)
		}
		fmt.Println("Body:")
		fmt.Printf("%s\n", req.Body)
		//	ch := getLinesChannel(conn)
		//	for line := range ch {
		//		fmt.Println(line)
		//	}
		conn.Close()
		fmt.Println("The connection has been closed.")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	go func() {
		defer f.Close()
		defer close(ch)
		line := ""
		buffer := make([]byte, 8)
		read, err := f.Read(buffer)
		for ; err != io.EOF; read, err = f.Read(buffer) {
			tempStr := string(buffer[:read])
			stringParts := strings.Split(tempStr, "\n")
			for i := 0; i < len(stringParts)-1; i++ {
				line += stringParts[i]
				ch <- line
				line = ""
			}
			line += stringParts[len(stringParts)-1]
		}
		if line != "" {
			ch <- line
		}
	}()
	return ch
}
