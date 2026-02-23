package main

import (
	"Go-Learning/config"
	"Go-Learning/routers"
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Bookstore Management System")
	router := routers.Router()

	port := config.GetEnvValue("PORT", "8080")
	fmt.Println("Server is running on port " + port)
	fmt.Println("To stop the server press ctrl+c")
	http.ListenAndServe(":"+port, router)
}
