package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type CreateRequest struct {
	Data string `json:"data"`
}

type DaprBindingRequest struct {
	Operation string `json:"operation"`
	Data      string `json:"data"`
}

func main() {
	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Endpoint to test secret retrieval
	http.HandleFunc("/test-secret", func(w http.ResponseWriter, r *http.Request) {
		// Call Dapr secrets endpoint
		resp, err := http.Get("http://localhost:3500/v1.0/secrets/aws-secretstore/test-secret")
		if err != nil {
			log.Printf("Error getting secret: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Read and return the response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Secret response: %s", string(body))
	})

	// Endpoint to create S3 binding
	http.HandleFunc("/create-s3", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		var createReq CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			log.Printf("Error decoding request: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Prepare the Dapr binding request
		daprReq := DaprBindingRequest{
			Operation: "create",
			Data:      createReq.Data,
		}

		// Convert to JSON
		payload, err := json.Marshal(daprReq)
		if err != nil {
			log.Printf("Error marshaling request: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Call Dapr bindings endpoint
		resp, err := http.Post(
			"http://localhost:3500/v1.0/bindings/aws-s3",
			"application/json",
			bytes.NewBuffer(payload),
		)
		if err != nil {
			log.Printf("Error calling binding: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Read and return the response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(resp.StatusCode)
		fmt.Fprintf(w, "Binding response: %s", string(body))
	})

	log.Println("Starting server on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
