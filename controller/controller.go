package controller

import (
	"Go-Learning/database"
	"Go-Learning/models"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var DatabaseConnection = database.Collection // Assuming GetDatabaseConnection() returns the database connection

// HomePage handles the home page request
func HomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "GET")

	json.NewEncoder(w).Encode("Welcome to the Bookstore Management System")
}

// GetAllBookDetails retrieves all book details from the database
func GetAllBookDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	books := readAllBooks()
	json.NewEncoder(w).Encode(books)
	log.Println("All Books: ", books)
}

func readAllBooks() []primitive.M {
	cur, err := DatabaseConnection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.Background())

	var bookstore []primitive.M

	for cur.Next(context.Background()) {
		var book bson.M
		err := cur.Decode(&book)
		if err != nil {
			log.Fatal(err)
		}
		bookstore = append(bookstore, book)
		log.Println("Record: ", bookstore)
	}
	return bookstore
}

// GetOneBookDetails retrieves a single book detail from the database using the provided ID
func GetOneBookDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "GET")

	params := mux.Vars(r)
	idParam := params["id"]
	log.Println("Received ID:", idParam)

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	bookDetails := readOneBook(id)
	if bookDetails == nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(bookDetails)
	log.Println("Single BookDetails: ", bookDetails)
}

func readOneBook(id primitive.ObjectID) primitive.M {
	var bookDetails bson.M
	err := DatabaseConnection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&bookDetails)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		log.Fatal(err)
	}
	return bookDetails
}

// InsertBookDetails inserts a new book detail into the database
func InsertBookDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")

	var bookDetail models.BookDetails
	_ = json.NewDecoder(r.Body).Decode(&bookDetail)

	insertOneBookDetails(bookDetail)
	json.NewEncoder(w).Encode(bookDetail)
}

func insertOneBookDetails(bookDetail models.BookDetails) {
	insertResult, err := DatabaseConnection.InsertOne(context.Background(), bookDetail)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Inserted a Single Record ", insertResult.InsertedID)
}

// UpdateBookDetails updates an existing book detail in the database
func UpdateBookDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "PUT")

	var bookDetail models.BookDetails
	_ = json.NewDecoder(r.Body).Decode(&bookDetail)
	params := mux.Vars(r)
	idParam := params["id"]
	log.Println("Received ID:", idParam)

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	updateOneBookDetails(id, bookDetail)
	json.NewEncoder(w).Encode(bookDetail)
}

func updateOneBookDetails(id primitive.ObjectID, bookDetail models.BookDetails) {
	update := bson.M{
		"$set": bookDetail,
	}
	updateResult, err := DatabaseConnection.UpdateOne(context.Background(), bson.M{"_id": id}, update)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Updated a Single Record ", updateResult.UpsertedID)
}

// DeleteBookDetails deletes a book detail from the database using the provided ID
func DeleteBookDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE")

	params := mux.Vars(r)
	idParam := params["id"]
	log.Println("Received ID:", idParam)

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	deleteOneBookDetails(id)
}

func deleteOneBookDetails(id primitive.ObjectID) {
	deleteResult, err := DatabaseConnection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Deleted a Single Record ", deleteResult.DeletedCount)
}
