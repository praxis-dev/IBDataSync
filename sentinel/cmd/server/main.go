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

type OrderDetails struct {
    Details    map[string]interface{}
    Timestamp  time.Time
}

var orderCache = make(map[int64]OrderDetails)



func isEqual(map1, map2 map[string]interface{}) bool {
    json1, err1 := json.Marshal(map1)
    json2, err2 := json.Marshal(map2)
    if err1 != nil || err2 != nil {
        return false
    }
    return string(json1) == string(json2)
}

func getOrderDetails(orderData map[string]interface{}) map[string]interface{} {
    orderDetails := make(map[string]interface{})
    orderDetails["permId"] = orderData["permId"]
    orderDetails["status"] = orderData["status"]
    orderDetails["filled"] = orderData["filled"]
    orderDetails["remaining"] = orderData["remaining"]
    orderDetails["avgFillPrice"] = orderData["avgFillPrice"]
    orderDetails["lastFillPrice"] = orderData["lastFillPrice"]
    return orderDetails
}


func sendResult(data []byte) {
    var receivedData map[string]interface{}
    err := json.Unmarshal(data, &receivedData)
    if err != nil {
        log.Printf("Error unmarshalling received data: %v", err)
        return
    }
    if orderData, ok := receivedData["data"].(map[string]interface{}); ok {
        if permID, ok := orderData["permId"].(float64); ok {
            permIDInt := int64(permID)
            orderDetails := getOrderDetails(orderData)

            isDuplicate := false
            for _, cachedOrder := range orderCache {
                if isEqual(cachedOrder.Details, orderDetails) {
                    isDuplicate = true
                    break
                }
            }

            if isDuplicate {
                log.Printf("Duplicate order update detected for permId: %d, skipping processing", permIDInt)
                return
            }

            orderCache[permIDInt] = OrderDetails{
                Details: orderDetails,
            }
        }
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
