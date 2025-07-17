package reference

import (
	"errors"
	"fmt"

	"github.com/maniosgrivei/go-test-drivers/customer"
)

type ReferenceCustomerRepository struct {
	customers  []*customer.Customer
	idIndex    map[string]*customer.Customer
	nameIndex  map[string]*customer.Customer
	emailIndex map[string]*customer.Customer
	phoneIndex map[string]*customer.Customer
}

var _ customer.CustomerRepository = (*ReferenceCustomerRepository)(nil)

func NewReferenceCustomerRepository() *ReferenceCustomerRepository {
	return &ReferenceCustomerRepository{
		customers:  make([]*customer.Customer, 0),
		idIndex:    make(map[string]*customer.Customer),
		nameIndex:  make(map[string]*customer.Customer),
		emailIndex: make(map[string]*customer.Customer),
		phoneIndex: make(map[string]*customer.Customer),
	}
}

// Save adds a new customer to the repository.
func (r *ReferenceCustomerRepository) Save(c *customer.Customer) error {
	if r.customers == nil || r.nameIndex == nil || r.emailIndex == nil || r.phoneIndex == nil {
		return customer.ErrSystem
	}

	if err := r.checkDuplication(c); err != nil {
		return fmt.Errorf("%w: %w", customer.ErrDuplication, err)
	}

	r.customers = append(r.customers, c)
	r.idIndex[c.ID] = c
	r.nameIndex[c.Name] = c
	r.emailIndex[c.Email] = c
	r.phoneIndex[c.Phone] = c

	return nil
}

// checkDuplication checks if the id name, email, or phone in the request
// already exist in the repository.
func (r *ReferenceCustomerRepository) checkDuplication(c *customer.Customer) error {
	var errs []error

	if _, found := r.idIndex[c.ID]; found {
		errs = append(errs, fmt.Errorf("duplicated id: '%s'", c.ID))
	}

	if _, found := r.nameIndex[c.Name]; found {
		errs = append(errs, fmt.Errorf("duplicated name: '%s'", c.Name))
	}

	if _, found := r.emailIndex[c.Email]; found {
		errs = append(errs, fmt.Errorf("duplicated email: '%s'", c.Email))
	}

	if _, found := r.phoneIndex[c.Phone]; found {
		errs = append(errs, fmt.Errorf("duplicated phone: '%s'", c.Phone))
	}

	return errors.Join(errs...)
}
