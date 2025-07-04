package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Create a new chi router
	r := chi.NewRouter()

	// Simple route that returns a string
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!"))
	})

	// Route that returns JSON
	r.Get("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"message": "Hello, World!",
			"status":  "success",
		}
		json.NewEncoder(w).Encode(response)
	})

	// Route with path parameter
	r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"user_id": id,
			"message": "User details retrieved",
		}
		json.NewEncoder(w).Encode(response)
	})

	// Start the server
	log.Println("Chi server starting on http://localhost:3000")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
