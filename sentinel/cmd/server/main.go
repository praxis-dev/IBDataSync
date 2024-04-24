package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
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

func init() {
    rand.Seed(time.Now().UnixNano())
}

func generateCustomOrderId() string {
    randomNumber := rand.Intn(900000) + 100000
    return fmt.Sprintf("prax%d", randomNumber)
}


func sendResult(data []byte) {
    var receivedData map[string]interface{}
    err := json.Unmarshal(data, &receivedData)
    if err != nil {
        log.Printf("Error unmarshalling received data: %v", err)
        return
    }

    newOrderId := generateCustomOrderId()

    if orderData, ok := receivedData["data"].(map[string]interface{}); ok {
        orderData["orderId"] = newOrderId

        if order, ok := orderData["order"].(map[string]interface{}); ok {
            delete(order, "totalQuantity")
            order["percentageAllocation"] = 7.5
        }
    }

    if _, ok := receivedData["command"]; !ok {
        receivedData["command"] = "orderStatus"
    }

    modifiedData, err := json.Marshal(receivedData)
    if err != nil {
        log.Printf("Error marshalling modified data: %v", err)
        return
    }

    jsonData := string(modifiedData)
    log.Printf("Sending data: %s\n", jsonData)

    form := url.Values{}
    form.Add("api_token", apiToken)
    form.Add("data", jsonData)
    formData := strings.NewReader(form.Encode())

    req, err := http.NewRequest("POST", apiURL, formData)
    if err != nil {
        log.Fatalf("Error creating POST request: %v", err)
        return
    }
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

    client := &http.Client{}
    resp, err := client.Do(req)
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

    if resp.StatusCode != http.StatusOK {
        log.Printf("Received non-OK response: %d, response body: %s", resp.StatusCode, body)
    } else {
        log.Println("Data sent successfully and received OK response from the server.")
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
