package orderprocessor

import (
	"log"

	customorderid "sentinel/custom_orderid"
	httpclient "sentinel/http_client"
	mongoservice "sentinel/mongo_service"
	ordercache "sentinel/order_cache"
)

type OrderProcessor struct {
	apiToken     string
	apiURL       string
	httpClient   httpclient.Client
	apiClient    *APIClient
	mongoService *mongoservice.MongoService
}

func NewOrderProcessor(apiToken, apiURL string, httpClient httpclient.Client, mongoService *mongoservice.MongoService) *OrderProcessor {
	apiClient := NewAPIClient(apiToken, apiURL, httpClient)

	return &OrderProcessor{
		apiToken:     apiToken,
		apiURL:       apiURL,
		httpClient:   httpClient,
		apiClient:    apiClient,
		mongoService: mongoService,
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
	op.mongoService.SaveOrder(modifiedData)
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
		log.Printf("Original order data: %+v", orderData)
		orderData["orderId"] = newOrderId

		if order, ok := orderData["order"].(map[string]interface{}); ok {
			log.Printf("Before modification order data: %+v", order)
			delete(order, "totalQuantity")
			order["percentageAllocation"] = 7.5
			order["testOrder"] = false
			log.Printf("After modification order data: %+v", order)
		} else {
			log.Printf("Order key missing in received orderData")
		}
	} else {
		log.Printf("Data key missing in receivedData")
	}

	receivedData["command"] = "orderStatusNew"
	log.Printf("Processed order data to be sent: %+v", receivedData)
	return receivedData
}
