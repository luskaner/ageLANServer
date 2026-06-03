package internal

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error updating to WebSocket: %v", err)
		return
	}
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
	}(conn)
	log.Println("Client connected to WebSocket")
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Cliente disconnected: %v", err)
			break
		}
		log.Printf("Message received: %s", p)
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Printf("Error sending message: %v", err)
			break
		}
	}
}

func ListenAndServeWebsocket(port uint16, sslCert string, sslKey string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(int(port)),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("WSS starting on port: %d\n", port)
	go func() {
		var err error
		if sslCert != "" && sslKey != "" {
			err = server.ListenAndServeTLS(sslCert, sslKey)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()
}
