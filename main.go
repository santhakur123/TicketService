package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-me"
		log.Println("WARNING: JWT_SECRET not set, using insecure default. Set JWT_SECRET in production.")
	}

	store := NewStore()
	app := NewApp(store, []byte(jwtSecret))

	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("GET /health", app.handleHealth)
	mux.HandleFunc("POST /auth/register", app.handleRegister)
	mux.HandleFunc("POST /auth/login", app.handleLogin)

	// Protected endpoints
	mux.HandleFunc("POST /tickets", app.authMiddleware(app.handleCreateTicket))
	mux.HandleFunc("GET /tickets", app.authMiddleware(app.handleListTickets))
	mux.HandleFunc("GET /tickets/{id}", app.authMiddleware(app.handleGetTicket))
	mux.HandleFunc("PATCH /tickets/{id}/status", app.authMiddleware(app.handleUpdateTicketStatus))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("ticket-system listening on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
