package customer

import (
	"errors"
	"fmt"
	"time"
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

	if err = checkDuplication(customer); err != nil {
		return "", fmt.Errorf("%w: %w", ErrDuplication, err)
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

// checkDuplication checks if the id, name, email, or phone in the request
// already exist in the repository.
func checkDuplication(customer *Customer) error {
	var errs []error

	if _, found := idIndex[customer.ID]; found {
		errs = append(errs, fmt.Errorf("duplicated id: '%s'", customer.ID))
	}

	if _, found := nameIndex[customer.Name]; found {
		errs = append(errs, fmt.Errorf("duplicated name: '%s'", customer.Name))
	}

	if _, found := emailIndex[customer.Email]; found {
		errs = append(errs, fmt.Errorf("duplicated email: '%s'", customer.Email))
	}

	if _, found := phoneIndex[customer.Phone]; found {
		errs = append(errs, fmt.Errorf("duplicated phone: '%s'", customer.Phone))
	}

	return errors.Join(errs...)
}
