package main

import (
	"net"
	"log"
	"bufio"
	"os"
	"fmt"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("Error resolving the UDP Addr: %v", err) 
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Error opening the UDP connection: %v", err)
	}
	defer udpConn.Close()
	
	consoleReader := bufio.NewReader(os.Stdin)
	if consoleReader == nil {
		log.Fatal("Error creating buffer for standard input")
	}

	for {
		fmt.Print("> ")
		line, err := consoleReader.ReadString('\n')
		if err != nil {
			log.Println(err)
			continue
		}
		udpConn.Write([]byte(line))
	}
}
