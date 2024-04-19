package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

var (
    apiToken string
    apiURL   string
    wsURL    string
)

func main() {

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiToken = os.Getenv("API_TOKEN")
    apiURL = os.Getenv("API_URL")
    wsURL = os.Getenv("WS_URL")

    if apiToken == "" || apiURL == "" || wsURL == "" {
        log.Fatal("Environment variables API_TOKEN, API_URL, or WS_URL are not set")
    }
	e := echo.New()

	go startWebSocketClient()

	e.Logger.Fatal(e.Start(":8080"))
}

func sendResult(data interface{}) {
	payload, err := json.Marshal(map[string]interface{}{
		"api_token": apiToken,
		"data":      data,
	})
	if err != nil {
		log.Fatalf("Error marshalling data: %v", err)
		return
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Fatalf("Error sending POST request: %v", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}
	body := string(bodyBytes)

	if resp.StatusCode == http.StatusOK {
		log.Println("Data sent successfully and received OK response from the server.")
	} else {
		log.Printf("Received non-OK response: %d, response body: %s", resp.StatusCode, body)
	}
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
		sendResult(message)
	}

}
