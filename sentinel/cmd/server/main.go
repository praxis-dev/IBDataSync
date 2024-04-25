package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	httpclient "sentinel/http_client"
	orderprocessor "sentinel/order_processor"
	"sentinel/websocket"

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

    httpClient := httpclient.NewClient()

    orderProcessor := orderprocessor.NewOrderProcessor(apiToken, apiURL, httpClient)

	messageChan := make(chan []byte)
	go websocket.StartWebSocketClient(wsURL, messageChan)

	go func() {
		for message := range messageChan {
            orderProcessor.ProcessOrder(message)
		}
	}()

	e.Logger.Fatal(e.Start(":8080"))
}

func init() {
    rand.Seed(time.Now().UnixNano())
}
