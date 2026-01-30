package utils

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

// ConnectDB initializes the MongoDB connection and returns error if any
func ConnectDB(uri string) (*mongo.Client, error) {
	if Client != nil {
		return Client, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Ping to ensure connection is alive
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB")
	Client = client
	return Client, nil
}

// GetCollection returns a MongoDB collection
func GetCollection(dbName, collectionName string) *mongo.Collection {
	if Client == nil {
		log.Fatal("MongoDB client is not initialized")
	}
	return Client.Database(dbName).Collection(collectionName)
}
