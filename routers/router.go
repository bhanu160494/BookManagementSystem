package routers

import (
	"Go-Learning/controller"
	"net/http"

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
	router.HandleFunc("/openapi.yaml", serveOpenAPISpec).Methods("GET")
	router.HandleFunc("/swagger", swaggerUIRedirect).Methods("GET")
	router.HandleFunc("/swagger/", swaggerUIPage).Methods("GET")

	return router
}

func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "openapi.yaml")
}

func swaggerUIRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
}

func swaggerUIPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<title>BookManagementSystem API Docs</title>
	<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
	<style>
		body {
			margin: 0;
			background: #f6f7fb;
		}
		.swagger-ui .topbar {
			display: none;
		}
		.swagger-ui {
			box-sizing: border-box;
		}
		.swagger-ui .info .title {
			color: #1f2937;
		}
	</style>
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
	<script>
		window.ui = SwaggerUIBundle({
			url: '/openapi.yaml',
			dom_id: '#swagger-ui',
			deepLinking: true,
			presets: [SwaggerUIBundle.presets.apis],
			layout: 'BaseLayout'
		});
	</script>
</body>
</html>`))
}
