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

// CustomerService manages customer-related operations.
type CustomerService struct {
	customers  []*Customer
	idIndex    map[string]*Customer
	nameIndex  map[string]*Customer
	emailIndex map[string]*Customer
	phoneIndex map[string]*Customer
}

// NewCustomerService creates a new instance of CustomerService.
func NewCustomerService() *CustomerService {
	return &CustomerService{
		customers:  make([]*Customer, 0),
		idIndex:    make(map[string]*Customer),
		nameIndex:  make(map[string]*Customer),
		emailIndex: make(map[string]*Customer),
		phoneIndex: make(map[string]*Customer),
	}
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
	if s.customers == nil || s.nameIndex == nil || s.emailIndex == nil || s.phoneIndex == nil {
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

	if err = s.checkDuplication(customer); err != nil {
		return "", fmt.Errorf("%w: %w", ErrDuplication, err)
	}

	s.customers = append(s.customers, customer)
	s.idIndex[customer.ID] = customer
	s.nameIndex[customer.Name] = customer
	s.emailIndex[customer.Email] = customer
	s.phoneIndex[customer.Phone] = customer

	return id, nil
}

// checkDuplication checks if the id, name, email, or phone in the request
// already exist in the repository.
func (s *CustomerService) checkDuplication(customer *Customer) error {
	var errs []error

	if _, found := s.idIndex[customer.ID]; found {
		errs = append(errs, fmt.Errorf("duplicated id: '%s'", customer.ID))
	}

	if _, found := s.nameIndex[customer.Name]; found {
		errs = append(errs, fmt.Errorf("duplicated name: '%s'", customer.Name))
	}

	if _, found := s.emailIndex[customer.Email]; found {
		errs = append(errs, fmt.Errorf("duplicated email: '%s'", customer.Email))
	}

	if _, found := s.phoneIndex[customer.Phone]; found {
		errs = append(errs, fmt.Errorf("duplicated phone: '%s'", customer.Phone))
	}

	return errors.Join(errs...)
}
