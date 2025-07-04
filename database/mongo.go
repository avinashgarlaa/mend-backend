package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Client

// ConnectDB connects to MongoDB using URI from environment
func ConnectDB() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		panic("MONGO_URI is not set in environment")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to MongoDB")
	DB = client
}

// GetCollection returns a Mongo collection from the Mend DB
func GetCollection(collectionName string) *mongo.Collection {
	return DB.Database("mend").Collection(collectionName)
}
