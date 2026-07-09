package app

import (
	"encoding/json"
	"log/slog"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/handlers"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/middleware"

	_ "github.com/fburtin/golang-senior-microservices-showcase/docs"
)

func NewRouter(customerHandler *handlers.CustomerHandler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("GET /customers", customerHandler.GetCustomers)
	mux.HandleFunc("POST /customers", customerHandler.CreateCustomer)
	mux.HandleFunc("GET /customers/{id}", customerHandler.GetCustomerByID)
	mux.HandleFunc("PUT /customers/{id}", customerHandler.UpdateCustomer)
	mux.HandleFunc("DELETE /customers/{id}", customerHandler.DeleteCustomer)

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

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
