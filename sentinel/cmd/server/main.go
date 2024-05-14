package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	httpclient "sentinel/http_client"
	mongoservice "sentinel/mongo_service"
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
	mongoURI := os.Getenv("MONGO_URI")
	mongoDatabase := os.Getenv("MONGO_DATABASE")
	mongoCollection := os.Getenv("MONGO_COLLECTION")

	if apiToken == "" || apiURL == "" || wsURL == "" || mongoURI == "" || mongoDatabase == "" || mongoCollection == "" {
		log.Fatal("Environment variables API_TOKEN, API_URL, WS_URL, MONGO_URI, MONGO_DATABASE, or MONGO_COLLECTION are not set")
	}

	e := echo.New()
	httpClient := httpclient.NewClient()

	mongoService := mongoservice.NewMongoService(mongoURI, mongoDatabase, mongoCollection)

	orderProcessor := orderprocessor.NewOrderProcessor(apiToken, apiURL, httpClient, mongoService)

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
