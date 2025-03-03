package main

import (
	"Go-Learning/routers"
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Bookstore Management System")
	router := routers.Router()

	fmt.Print("Please enter the port number to run the server on: ")
	var port string
	fmt.Scanln(&port)
	fmt.Println("Server is running on port " + port)
	fmt.Println("To stop the server press ctrl+c")
	http.ListenAndServe(":"+port, router)
}
