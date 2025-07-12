package customer

import (
	"errors"
	"fmt"
)

// Package-level errors for customer registration.
var (
	ErrValidation  = fmt.Errorf("validation error")
	ErrDuplication = fmt.Errorf("duplication error")
	ErrSystem      = fmt.Errorf("system error: contact support")
)

// Customer represents a customer in the CRM system.
type Customer struct {
	ID    string
	Name  string
	Email string
	Phone string
}

// Package-level variables to store customers in-memory.
var (
	customers  []*Customer
	idIndex    map[string]*Customer
	nameIndex  map[string]*Customer
	emailIndex map[string]*Customer
	phoneIndex map[string]*Customer
)

// Init initializes the in-memory customer repository, clearing all existing
// data.
func Init() {
	customers = make([]*Customer, 0)
	idIndex = make(map[string]*Customer)
	nameIndex = make(map[string]*Customer)
	emailIndex = make(map[string]*Customer)
	phoneIndex = make(map[string]*Customer)
}

// RegisterRequest carries the required data for registering a new customer.
type RegisterRequest struct {
	Name  string
	Email string
	Phone string
}

// Register validates the request, checks for duplicates, and adds a new
// customer to the repository.
func Register(request *RegisterRequest) (id string, err error) {
	if customers == nil || nameIndex == nil || emailIndex == nil || phoneIndex == nil {
		return "", ErrSystem
	}

	id = fmt.Sprintf("%d", len(customers)+1)

	if err = validateRegisterRequest(request); err != nil {
		return "", fmt.Errorf("%w: %w", ErrValidation, err)
	}

	if err = checkDuplication(request); err != nil {
		return "", fmt.Errorf("%w: %w", ErrDuplication, err)
	}

	customer := &Customer{
		ID:    id,
		Name:  request.Name,
		Email: request.Email,
		Phone: request.Phone,
	}

	customers = append(customers, customer)
	idIndex[customer.ID] = customer
	nameIndex[customer.Name] = customer
	emailIndex[customer.Email] = customer
	phoneIndex[customer.Phone] = customer

	return id, nil
}

// validateRegisterRequest checks that all required fields in the request are
// valid.
func validateRegisterRequest(request *RegisterRequest) error {
	var errs []error

	if request.Name == "" {
		errs = append(errs, fmt.Errorf("invalid name: '%s'", request.Name))
	}

	if request.Email == "" {
		errs = append(errs, fmt.Errorf("invalid email: '%s'", request.Email))
	}

	if request.Phone == "" {
		errs = append(errs, fmt.Errorf("invalid phone: '%s'", request.Phone))
	}

	return errors.Join(errs...)
}

// checkDuplication checks if the name, email, or phone in the request already
// exist in the repository.
func checkDuplication(request *RegisterRequest) error {
	var errs []error

	if _, found := nameIndex[request.Name]; found {
		errs = append(errs, fmt.Errorf("duplicated name: '%s'", request.Name))
	}

	if _, found := emailIndex[request.Email]; found {
		errs = append(errs, fmt.Errorf("duplicated email: '%s'", request.Email))
	}

	if _, found := phoneIndex[request.Phone]; found {
		errs = append(errs, fmt.Errorf("duplicated phone: '%s'", request.Phone))
	}

	return errors.Join(errs...)
}
