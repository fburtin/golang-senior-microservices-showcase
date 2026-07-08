package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/services"
)

type CustomerHandler struct {
	service *services.CustomerService
}

func NewCustomerHandler(service *services.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		service: service,
	}
}

func (h *CustomerHandler) HandleCustomers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getCustomers(w, r)

	case http.MethodPost:
		h.createCustomer(w, r)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

func (h *CustomerHandler) getCustomers(w http.ResponseWriter, r *http.Request) {
	customers := h.service.GetAll()
	writeJSON(w, http.StatusOK, customers)
}

func (h *CustomerHandler) createCustomer(w http.ResponseWriter, r *http.Request) {
	var customer domain.Customer

	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	customer.ID = time.Now().Format("20060102150405")

	err = h.service.Create(customer)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "could not create customer",
		})
		return
	}

	writeJSON(w, http.StatusCreated, customer)
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(value)
}

func (h *CustomerHandler) HandleCustomerByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	switch r.Method {
	case http.MethodGet:
		customer, err := h.service.GetByID(id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "customer not found",
			})
			return
		}

		writeJSON(w, http.StatusOK, customer)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}
