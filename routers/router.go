package routers

import (
	"Go-Learning/controller"

	"github.com/gorilla/mux"
)

// Router initializes the API routes
func Router() *mux.Router {
	router := mux.NewRouter()

	// Define the routes and their handlers
	router.HandleFunc("/", controller.HomePage).Methods("GET")
	router.HandleFunc("/books", controller.GetAllBookDetails).Methods("GET")
	router.HandleFunc("/book/{id}", controller.GetOneBookDetails).Methods("GET")
	router.HandleFunc("/book", controller.InsertBookDetails).Methods("POST")
	router.HandleFunc("/book/{id}", controller.UpdateBookDetails).Methods("PUT")
	router.HandleFunc("/book/{id}", controller.DeleteBookDetails).Methods("DELETE")

	return router
}
