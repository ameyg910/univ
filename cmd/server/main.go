package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ameyg910/unidate/internal/auth"
	"github.com/ameyg910/unidate/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Set up database connection
	db, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Set up routes
	r := chi.NewRouter()

	// Authentication routes
	r.Post("/register", auth.SendOTPHandler(db)) // Register via OTP
	r.Post("/login", auth.VerifyOTPHandler(db))  // Login via OTP and JWT

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Unidate server is up"))
	})

	// Start server on port 8081
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default to 8081 if not set in .env
	}
	fmt.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
