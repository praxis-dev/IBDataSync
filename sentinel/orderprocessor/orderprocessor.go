package orderprocessor

import (
	"sentinel/httpclient"

	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type OrderDetails struct {
    Details   map[string]interface{}
    Timestamp time.Time
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


func generateCustomOrderId() string {
    randomNumber := rand.Intn(900000) + 100000
    return fmt.Sprintf("prax%d", randomNumber)
}

type OrderProcessor struct {
    apiToken string
    apiURL   string
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