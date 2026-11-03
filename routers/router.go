package routers

import (
	"net/http"

	"Go-Learning/auth"
	"Go-Learning/config"
	"Go-Learning/controller"

	"github.com/gorilla/mux"
)

// Router initializes the API routes
func Router() *mux.Router {
	keycloakCfg := config.LoadKeycloakConfig()
	keycloakClient := auth.NewKeycloakClient(keycloakCfg)
	authMiddleware := auth.NewMiddleware(auth.NewTokenVerifier(keycloakCfg))
	authController := controller.NewAuthController(keycloakClient)

	readerRole := config.GetEnvValue("BOOKSTORE_READER_ROLE", "book-user")
	adminRole := config.GetEnvValue("BOOKSTORE_ADMIN_ROLE", "book-admin")

	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/", controller.HomePage).Methods("GET")
	router.HandleFunc("/auth/signup", authController.Signup).Methods("POST")
	router.HandleFunc("/auth/signin", authController.Signin).Methods("POST")
	router.HandleFunc("/auth/refresh", authController.Refresh).Methods("POST")
	router.HandleFunc("/auth/logout", authController.Logout).Methods("POST")
	router.HandleFunc("/openapi.yaml", serveOpenAPISpec).Methods("GET")
	router.HandleFunc("/swagger", swaggerUIRedirect).Methods("GET")
	router.HandleFunc("/swagger/", swaggerUIPage).Methods("GET")

	// Any authenticated user may read their own profile and browse the catalogue.
	readRoutes := router.NewRoute().Subrouter()
	readRoutes.Use(authMiddleware.Authenticate, authMiddleware.RequireRole(readerRole, adminRole))
	readRoutes.HandleFunc("/auth/me", authController.Me).Methods("GET")
	readRoutes.HandleFunc("/books", controller.GetAllBookDetails).Methods("GET")
	readRoutes.HandleFunc("/book/{id}", controller.GetOneBookDetails).Methods("GET")

	// Mutating operations require the administrator role.
	writeRoutes := router.NewRoute().Subrouter()
	writeRoutes.Use(authMiddleware.Authenticate, authMiddleware.RequireRole(adminRole))
	writeRoutes.HandleFunc("/book", controller.InsertBookDetails).Methods("POST")
	writeRoutes.HandleFunc("/book/{id}", controller.UpdateBookDetails).Methods("PUT")
	writeRoutes.HandleFunc("/book/{id}", controller.DeleteBookDetails).Methods("DELETE")

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
			persistAuthorization: true,
			presets: [SwaggerUIBundle.presets.apis],
			layout: 'BaseLayout'
		});
	</script>
</body>
</html>`))
}
