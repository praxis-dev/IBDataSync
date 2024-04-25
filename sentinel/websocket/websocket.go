package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func StartWebSocketClient(wsURL string, messageChan chan<- []byte) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket server:", err)
		return
	}
	defer conn.Close()

	log.Println("Connected to WebSocket server")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading from WebSocket:", err)
			return
		}
		log.Printf("Received message: %s\n", message)
		messageChan <- message
	}
}