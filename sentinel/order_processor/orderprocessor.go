package orderprocessor

import (
	"log"

	customorderid "sentinel/custom_orderid"
	httpclient "sentinel/http_client"
	ordercache "sentinel/order_cache"
)

type OrderProcessor struct {
    apiToken   string
    apiURL     string
    httpClient httpclient.Client
    apiClient *APIClient
}

func NewOrderProcessor(apiToken, apiURL string, httpClient httpclient.Client) *OrderProcessor {
    apiClient := NewAPIClient(apiToken, apiURL, httpClient)

    return &OrderProcessor{
        apiToken:   apiToken,
        apiURL:     apiURL,
        httpClient: httpClient,
        apiClient: apiClient,

    }
}

func (op *OrderProcessor) ProcessOrder(data []byte) {
    receivedData, err := UnmarshallData(data)
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
    modifiedData, err := MarshallData(processedData)
    if err != nil {
        log.Printf("Error marshalling modified data: %v", err)
        return
    }

    op.apiClient.SendAPIRequest(modifiedData)
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
    newOrderId := customorderid.GenerateCustomOrderId()

    if orderData, ok := receivedData["data"].(map[string]interface{}); ok {
        orderData["orderId"] = newOrderId

        if order, ok := orderData["order"].(map[string]interface{}); ok {
            delete(order, "totalQuantity")
            order["percentageAllocation"] = 3.75
            order["testOrder"] = true
        }
    }

    receivedData["command"] = "orderStatusNew"

    return receivedData
}
