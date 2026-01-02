package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"secret-share/internal/api"
	"secret-share/internal/store"
)

func main() {
	// Initialize Store with 2GB Limit
	// 2 * 1024 * 1024 * 1024 = 2147483648 bytes
	memStore := store.NewMemoryStore(2 * 1024 * 1024 * 1024)

	// Initialize API Server
	server := api.NewServer(memStore)

	// Define Routes
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	// Serve OpenAPI Spec
	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "api/docs/openapi.yaml")
	})

	// Serve Swagger UI
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("static/swagger-ui"))))

	mux.HandleFunc("/secret", server.HandleCreateSecret)
	mux.HandleFunc("/retrieve", server.HandleRetrieveSecret)

	// Stats Endpoint
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		used, limit, created, retrieved := memStore.Stats()
		// Simple JSON
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"memory_used": %d, "memory_limit": %d, "percent_used": %.2f, "secrets_created": %d, "secrets_retrieved": %d}`,
			used, limit, float64(used)/float64(limit)*100, created, retrieved)
	})

	// Add simple health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := ":8080"
	s := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Printf("Secret Share Server starting on %s...\n", addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
