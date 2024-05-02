package orderprocessor

import (
	"context"
	"encoding/json"
	"log"
	"time"

	customorderid "sentinel/custom_orderid"
	httpclient "sentinel/http_client"
	ordercache "sentinel/order_cache"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderProcessor struct {
    apiToken        string
    apiURL          string
    httpClient      httpclient.Client
    apiClient       *APIClient
    mongoClient     *mongo.Client
    mongoDatabase   string
    mongoCollection string
}

func NewOrderProcessor(apiToken, apiURL, mongoURI, mongoDatabase, mongoCollection string, httpClient httpclient.Client) *OrderProcessor {
    apiClient := NewAPIClient(apiToken, apiURL, httpClient)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
    if err != nil {
        log.Fatalf("Error connecting to MongoDB: %v", err)
    }

    return &OrderProcessor{
        apiToken:        apiToken,
        apiURL:          apiURL,
        httpClient:      httpClient,
        apiClient:       apiClient,
        mongoClient:     mongoClient,
        mongoDatabase:   mongoDatabase,
        mongoCollection: mongoCollection,
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
    op.saveOrderToMongoDB(modifiedData)

}

func (op *OrderProcessor) saveOrderToMongoDB(orderData []byte) {
    var data map[string]interface{}
    err := json.Unmarshal(orderData, &data)
    if err != nil {
        log.Printf("Error unmarshalling order data: %v", err)
        return
    }

    log.Printf("Received data for saving to MongoDB: %+v", data)

    nestedData, ok := data["data"].(map[string]interface{})
    if !ok {
        log.Printf("Error: 'data' key is missing or not the expected type")
        return
    }

    order, ok := nestedData["order"].(map[string]interface{})
    if !ok {
        log.Printf("Error: 'order' key is missing or not the expected type")
        return
    }

    testOrder, ok := order["testOrder"].(bool)
    if !ok {
        log.Printf("Error: 'testOrder' key is missing or not the expected type")
        return
    }

    if !testOrder {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        collection := op.mongoClient.Database(op.mongoDatabase).Collection(op.mongoCollection)
        _, err := collection.InsertOne(ctx, nestedData)  // Inserting nestedData which contains the 'order'
        if err != nil {
            log.Printf("Error saving order to MongoDB: %v", err)
        }
    }
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
            order["percentageAllocation"] = 3.75
            order["testOrder"] = true
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
