package internal

import (
	"io"
	"log"
	"net"
	"strconv"
)

func handle(conn net.Conn) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)
	log.Printf("Client connected from: %s", conn.RemoteAddr().String())
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Printf("Cliente %s closed the connection.", conn.RemoteAddr().String())
			} else {
				log.Printf("Error reading from %s: %v", conn.RemoteAddr().String(), err)
			}
			return
		}
		log.Printf("[%s]: %s", conn.RemoteAddr().String(), string(buffer[:n]))
		_, err = conn.Write(buffer[:n])
		if err != nil {
			log.Printf("Error writing to %s: %v", conn.RemoteAddr().String(), err)
			return
		}
	}
}

func ListenTCP(port uint16) {
	go func() {
		listener, err := net.Listen("tcp4", ":"+strconv.Itoa(int(port)))
		if err != nil {
			log.Fatalf("Error starting TCP server: %v", err)
		}
		defer func(listener net.Listener) {
			_ = listener.Close()
		}(listener)
		log.Printf("Echo TCP Server started on port: %d\n", port)
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				continue
			}
			go handle(conn)
		}
	}()
}
