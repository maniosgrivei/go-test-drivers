package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/maniosgrivei/go-test-drivers/customer"
)

// CustomerRESTAPIHandler handles the HTTP transport for customer-related
// operations.
type CustomerRESTAPIHandler struct {
	service *customer.CustomerService
	handler http.Handler
}

// NewCustomerRESTAPIHandler creates and initializes a new CustomerRESTAPI
// instance. It sets up the routing and returns a configured API handler.
func NewCustomerRESTAPIHandler(service *customer.CustomerService) *CustomerRESTAPIHandler {
	api := &CustomerRESTAPIHandler{
		service: service,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /customers", api.RegisterHandler)
	api.handler = mux

	return api
}

// ServeHTTP makes CustomerRESTAPIHandler compatible with the http.Handler
// interface.
func (api *CustomerRESTAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.handler.ServeHTTP(w, r)
}

// RegisterHandler handles the HTTP request for registering a new customer.
func (api *CustomerRESTAPIHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var request customer.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	id, err := api.service.Register(&request)
	if err != nil {
		switch {
		case errors.Is(err, customer.ErrValidation):
			writeError(w, err.Error(), http.StatusBadRequest)

		case errors.Is(err, customer.ErrDuplication):
			writeError(w, err.Error(), http.StatusConflict)

		default:
			writeError(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// writeJSON is a helper function to write a JSON response.
func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeError is a helper function to write a JSON error response.
func writeError(w http.ResponseWriter, message string, statusCode int) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}
