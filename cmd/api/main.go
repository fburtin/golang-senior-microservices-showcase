package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/repositories"
)

func main() {
	customerRepository := repositories.NewMemoryCustomerRepository()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "OK",
			"service": "Go Senior Microservices Showcase",
			"version": "1.0.0",
		})
	})

	http.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			customers := customerRepository.GetAll()
			writeJSON(w, http.StatusOK, customers)

		case http.MethodPost:
			var customer domain.Customer

			err := json.NewDecoder(r.Body).Decode(&customer)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": "invalid request body",
				})
				return
			}

			customer.ID = time.Now().Format("20060102150405")
			err = customerRepository.Create(customer)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "could not create customer",
				})
				return
			}

			writeJSON(w, http.StatusCreated, customer)

		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
				"error": "method not allowed",
			})
		}
	})

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(value)
}
