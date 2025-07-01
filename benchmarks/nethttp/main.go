package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func main() {
	// Simple route that returns a string
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!"))
	})

	// Route that returns JSON
	http.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"message": "Hello, World!",
			"status":  "success",
		}
		json.NewEncoder(w).Encode(response)
	})

	// Route with path parameter
	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/users/") {
			http.NotFound(w, r)
			return
		}
		
		// Extract the ID from the path
		id := strings.TrimPrefix(r.URL.Path, "/users/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"user_id": id,
			"message": "User details retrieved",
		}
		json.NewEncoder(w).Encode(response)
	})

	// Start the server
	log.Println("Net/HTTP server starting on http://localhost:3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}