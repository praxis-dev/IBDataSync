package orderprocessor

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"

	"sentinel/httpclient"
	"sentinel/ordercache"
)

func generateCustomOrderId() string {
    randomNumber := rand.Intn(900000) + 100000
    return fmt.Sprintf("prax%d", randomNumber)
}

type OrderProcessor struct {
    apiToken   string
    apiURL     string
    httpClient httpclient.Client
}

func NewOrderProcessor(apiToken, apiURL string, httpClient httpclient.Client) *OrderProcessor {
    return &OrderProcessor{
        apiToken:   apiToken,
        apiURL:     apiURL,
        httpClient: httpClient,
    }
}

func (op *OrderProcessor) ProcessOrder(data []byte) {
    receivedData, err := op.unmarshallData(data)
    if err != nil {
        log.Printf("Error unmarshalling received data: %v", err)
        return
    }

    permIDInt, orderDetails, ok := op.extractOrderDetails(receivedData)
    if ok {
        if ordercache.IsDuplicate(permIDInt, orderDetails) {
            log.Printf("Duplicate order update detected for permId: %d, skipping processing", permIDInt)
            return
        }
        ordercache.CacheOrder(permIDInt, orderDetails)
    }

    processedData := op.processOrderData(receivedData)

    modifiedData, err := op.marshallData(processedData)
    if err != nil {
        log.Printf("Error marshalling modified data: %v", err)
        return
    }

    op.sendAPIRequest(modifiedData)
}

func (op *OrderProcessor) unmarshallData(data []byte) (map[string]interface{}, error) {
    var receivedData map[string]interface{}
    err := json.Unmarshal(data, &receivedData)
    return receivedData, err
}

func (op *OrderProcessor) extractOrderDetails(receivedData map[string]interface{}) (int64, map[string]interface{}, bool) {
    if orderData, ok := receivedData["data"].(map[string]interface{}); ok {
        if permID, ok := orderData["permId"].(float64); ok {
            permIDInt := int64(permID)
            orderDetails := ordercache.GetOrderDetails(orderData)
            return permIDInt, orderDetails, true
        }
    }
    return 0, nil, false
}

func (op *OrderProcessor) processOrderData(receivedData map[string]interface{}) map[string]interface{} {
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

    return receivedData
}

func (op *OrderProcessor) marshallData(data map[string]interface{}) ([]byte, error) {
    return json.Marshal(data)
}

func (op *OrderProcessor) sendAPIRequest(data []byte) {
    jsonData := string(data)
    log.Printf("Sending data: %s\n", jsonData)

    form := url.Values{}
    form.Add("api_token", op.apiToken)
    form.Add("data", jsonData)

    resp, err := op.httpClient.PostForm(op.apiURL, form)
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