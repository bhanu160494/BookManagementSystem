package database

import (
	"Go-Learning/config"
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Collection *mongo.Collection

func init() {
	connectionString := os.Getenv("MONGODB_URI")
	if connectionString == "" {
		log.Fatal("MONGODB_URI is not set")
	}

	dbName := config.GetEnvValue("MONGODB_DB_NAME", "test")
	collName := config.GetEnvValue("MONGODB_COLLECTION", "bookstore")

	//client option
	clientOptions := options.Client().ApplyURI(connectionString)

	//connect to mongodb
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	Collection = client.Database(dbName).Collection(collName)
	fmt.Println("Connected to MongoDB!", Collection)
}
