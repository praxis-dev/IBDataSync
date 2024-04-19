package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	wsURL = "ws://localhost:5000/ws"
)

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World from Go")
	})

	go startWebSocketClient()

	e.Logger.Fatal(e.Start(":8080"))
}

func startWebSocketClient() {
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
	}
}
