package database

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//temporary connection string for testing
const connectionString = "mongodb+srv://admin:admin@cluster0.9ipq5.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
const dbName = "test"
const collName = "bookstore"

var Collection *mongo.Collection

func init() {
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
