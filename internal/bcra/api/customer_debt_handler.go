package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/bcra"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/customerdebt"
)

type CustomerDebtHandler struct {
	service *customerdebt.Service
}

func NewCustomerDebtHandler(
	service *customerdebt.Service,
) *CustomerDebtHandler {
	return &CustomerDebtHandler{
		service: service,
	}
}

// GetCustomerDebt godoc
//
//	@Summary		Get customer debt information
//	@Description	Retrieves debt information from BCRA Central de Deudores and stores successful responses in MongoDB.
//	@Tags			BCRA
//	@Produce		json
//	@Param			cuit	path		string	true	"CUIT/CUIL/CDI without hyphens" example(20292456078)
//	@Success		200		{object}	bcra.DebtResponse
//	@Failure		400		{object}	bcra.ErrorResponse
//	@Failure		404		{object}	bcra.ErrorResponse
//	@Failure		502		{object}	bcra.ErrorResponse
//	@Failure		500		{object}	bcra.ErrorResponse
//	@Router			/customer-debts/{cuit} [get]
func (h *CustomerDebtHandler) GetCustomerDebt(
	writer http.ResponseWriter,
	request *http.Request,
) {
	cuit := request.PathValue("cuit")

	response, err := h.service.GetAndSave(
		request.Context(),
		cuit,
	)
	if err != nil {
		switch {
		case errors.Is(err, bcra.ErrInvalidCUIT):
			writeJSON(
				writer,
				http.StatusBadRequest,
				bcra.ErrorResponse{
					Status: http.StatusBadRequest,
					ErrorMessages: []string{
						"CUIT inválido.",
					},
				},
			)

		case errors.Is(err, bcra.ErrNotFound):
			writeJSON(
				writer,
				http.StatusNotFound,
				bcra.ErrorResponse{
					Status: http.StatusNotFound,
					ErrorMessages: []string{
						"No se encontró datos para la identificación ingresada.",
					},
				},
			)

		default:
			writeJSON(
				writer,
				http.StatusBadGateway,
				bcra.ErrorResponse{
					Status: http.StatusBadGateway,
					ErrorMessages: []string{
						"Unable to retrieve information from BCRA.",
					},
				},
			)
		}

		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		response,
	)
}

func writeJSON(
	writer http.ResponseWriter,
	statusCode int,
	value any,
) {
	writer.Header().Set(
		"Content-Type",
		"application/json",
	)
	writer.WriteHeader(statusCode)

	if err := json.NewEncoder(writer).Encode(value); err != nil {
		http.Error(
			writer,
			"failed to encode response",
			http.StatusInternalServerError,
		)
	}
}
