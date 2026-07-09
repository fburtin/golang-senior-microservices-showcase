package app

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/handlers"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/middleware"
)

func NewRouter(customerHandler *handlers.CustomerHandler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/customers", customerHandler.HandleCustomers)
	mux.HandleFunc("/customers/{id}", customerHandler.HandleCustomerByID)

	return middleware.RequestID(
		middleware.Recovery(
			middleware.RequestLogger(mux, logger),
			logger,
		),
	)
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

	_ = json.NewEncoder(w).Encode(value)
}
