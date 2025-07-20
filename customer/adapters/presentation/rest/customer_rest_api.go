package rest

import "github.com/maniosgrivei/go-test-drivers/customer"

// CustomerRESTAPIHandler handles the HTTP transport for customer-related
// operations.
type CustomerRESTAPIHandler struct {
	service *customer.CustomerService
}

// NewCustomerRESTAPIHandler creates and initializes a new CustomerRESTAPI
// instance. It sets up the routing and returns a configured API handler.
func NewCustomerRESTAPIHandler(service *customer.CustomerService) *CustomerRESTAPIHandler {
	return &CustomerRESTAPIHandler{
		service: service,
	}
}
