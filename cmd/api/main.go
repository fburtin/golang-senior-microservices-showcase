package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/handlers"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/repositories"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/services"
)

func main() {

	customerRepository := repositories.NewMemoryCustomerRepository()
	customerService := services.NewCustomerService(customerRepository)
	customerHandler := handlers.NewCustomerHandler(customerService)

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/customers", customerHandler.HandleCustomers)
	http.HandleFunc("/customers/{id}", customerHandler.HandleCustomerByID)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "OK",
		"service": "Go Senior Microservices Showcase",
		"version": "1.0.0",
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(value)
}
