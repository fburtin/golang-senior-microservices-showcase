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

type CreateCustomerRequest struct {
	FirstName string `json:"firstName" example:"John"`
	LastName  string `json:"lastName" example:"Doe"`
	Email     string `json:"email" example:"john@example.com"`
}

// GetCustomers godoc
// @Summary List customers
// @Tags Customers
// @Produce json
// @Success 200 {array} domain.Customer
// @Router /customers [get]
func (h *CustomerHandler) GetCustomers(w http.ResponseWriter, r *http.Request) {
	customers := h.service.GetAll()
	writeJSON(w, http.StatusOK, customers)
}

// CreateCustomer godoc
// @Summary Create customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param customer body CreateCustomerRequest true "Customer"
// @Success 201 {object} domain.Customer
// @Failure 400 {object} map[string]string
// @Router /customers [post]
func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {

	var request CreateCustomerRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	customer := domain.Customer{
		ID:        time.Now().Format("20060102150405"),
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
		CreatedAt: time.Now(),
	}

	createdCustomer, err := h.service.Create(customer)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, createdCustomer)
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(value)
}

// GetCustomerByID godoc
// @Summary Get customer by ID
// @Tags Customers
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} domain.Customer
// @Failure 404 {object} map[string]string
// @Router /customers/{id} [get]
func (h *CustomerHandler) GetCustomerByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	customer, err := h.service.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "customer not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, customer)
}

// UpdateCustomer godoc
// @Summary Update customer by ID
// @Tags Customers
// @Produce json
// @Accept json
// @Param id path string true "Customer ID"
// @Param customer body CreateCustomerRequest true "Customer"
// @Success 200 {object} domain.Customer
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var request CreateCustomerRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {

		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	customer := domain.Customer{
		ID:        id,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
	}

	err = h.service.Update(id, customer)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
		return
	}

	updatedCustomer, err := h.service.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "customer not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, updatedCustomer)
}

// DeleteCustomer godoc
// @Summary Delete customer by ID
// @Tags Customers
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.service.Delete(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "customer not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "customer deleted successfully",
	})
}
