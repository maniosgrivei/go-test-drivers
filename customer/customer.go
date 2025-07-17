package customer

import (
	"fmt"
	"time"
)

// Package-level errors for customer registration.
var (
	ErrValidation  = fmt.Errorf("validation error")
	ErrDuplication = fmt.Errorf("duplication error")
	ErrSystem      = fmt.Errorf("system error: contact support")
)

//
// Domain

// Customer represents a customer in the CRM system.
type Customer struct {
	ID    string
	Name  string
	Email string
	Phone string
}

//
// Dependencies

type CustomerRepository interface {
	// Save adds a new customer to the repository.
	Save(c *Customer) error
}

//
// Service

// CustomerService manages customer-related operations.
type CustomerService struct {
	repository CustomerRepository
}

// NewCustomerService creates a new instance of CustomerService.
func NewCustomerService(repository CustomerRepository) *CustomerService {
	return &CustomerService{repository: repository}
}

// RegisterRequest carries the required data for registering a new customer.
type RegisterRequest struct {
	Name  string
	Email string
	Phone string
}

// Register validates the request, checks for duplicates, and adds a new
// customer to the repository.
func (s *CustomerService) Register(request *RegisterRequest) (id string, err error) {
	if err = ValidateRegisterRequest(request); err != nil {
		return "", fmt.Errorf("%w: %w", ErrValidation, err)
	}

	id, err = GenerateID(request.Name, time.Now())
	if err != nil {
		return "", err
	}

	customer := &Customer{
		ID:    id,
		Name:  request.Name,
		Email: request.Email,
		Phone: request.Phone,
	}

	if err = s.repository.Save(customer); err != nil {
		return "", err
	}

	return id, nil
}
