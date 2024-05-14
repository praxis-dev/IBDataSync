package mongoservice

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoService struct {
	client     *mongo.Client
	database   string
	collection string
}

func NewMongoService(mongoURI, database, collection string) *MongoService {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	return &MongoService{
		client:     client,
		database:   database,
		collection: collection,
	}
}

func (ms *MongoService) SaveOrder(orderData []byte) {
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
		collection := ms.client.Database(ms.database).Collection(ms.collection)
		_, err := collection.InsertOne(ctx, nestedData)
		if err != nil {
			log.Printf("Error saving order to MongoDB: %v", err)
		}
	}
}
