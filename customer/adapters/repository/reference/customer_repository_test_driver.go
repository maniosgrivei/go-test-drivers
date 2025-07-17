//go:build test

package reference

import (
	"testing"

	"github.com/maniosgrivei/go-test-drivers/customer"
	"github.com/stretchr/testify/require"
)

type ReferenceCustomerRepositoryTestDriver struct {
	*ReferenceCustomerRepository
}

var _ customer.CustomerRepositoryTestDriver = (*ReferenceCustomerRepositoryTestDriver)(nil)

func NewReferenceCustomerRepositoryTestDriver(repository *ReferenceCustomerRepository) *ReferenceCustomerRepositoryTestDriver {
	return &ReferenceCustomerRepositoryTestDriver{
		ReferenceCustomerRepository: repository,
	}
}

//
// Arrange

// ArrangeInternalsNoCustomerIsRegistered initializes the repository to a clean state.
func (td *ReferenceCustomerRepositoryTestDriver) ArrangeInternalsNoCustomerIsRegistered(t *testing.T) {
	t.Helper()

	td.customers = make([]*customer.Customer, 0)
	td.idIndex = make(map[string]*customer.Customer)
	td.nameIndex = make(map[string]*customer.Customer)
	td.emailIndex = make(map[string]*customer.Customer)
	td.phoneIndex = make(map[string]*customer.Customer)
}

// ArrangeInternalsSomeCustomersAreRegistered populates the repository with the
// given customers.
func (td *ReferenceCustomerRepositoryTestDriver) ArrangeInternalsSomeCustomersAreRegistered(t *testing.T, cs []*customer.Customer) {
	t.Helper()
	td.ArrangeInternalsNoCustomerIsRegistered(t)

	for _, c := range cs {
		td.customers = append(td.customers, c)
		td.idIndex[c.ID] = c
		td.nameIndex[c.Name] = c
		td.emailIndex[c.Email] = c
		td.phoneIndex[c.Phone] = c
	}
}

// ArrangeInternalsSomethingCausingAProblem corrupts the internal state to
// ensure subsequent function calls will result in a system error.
func (td *ReferenceCustomerRepositoryTestDriver) ArrangeInternalsSomethingCausingAProblem(t *testing.T) {
	t.Helper()

	td.customers = nil
	td.idIndex = nil
	td.nameIndex = nil
	td.emailIndex = nil
	td.phoneIndex = nil
}

//
// Assert

// AssertInternalsCustomerShouldBeProperlyRegistered asserts that the customer
// is properly registered in the internal data structures.
func (td *ReferenceCustomerRepositoryTestDriver) AssertInternalsCustomerShouldBeProperlyRegistered(t *testing.T, c *customer.Customer) {
	t.Helper()

	r := require.New(t)

	r.True(sliceContainsCustomer(td.customers, c))

	r.Contains(td.idIndex, c.ID)
	r.True(customersAreSame(td.idIndex[c.ID], c))

	r.Contains(td.nameIndex, c.Name)
	r.True(customersAreSame(td.nameIndex[c.Name], c))

	r.Contains(td.emailIndex, c.Email)
	r.True(customersAreSame(td.emailIndex[c.Email], c))

	r.Contains(td.phoneIndex, c.Phone)
	r.True(customersAreSame(td.phoneIndex[c.Phone], c))
}

// AssertInternalsCustomerShouldNotBeRegistered asserts that the customer is not
// present in the internal data structures.
func (td *ReferenceCustomerRepositoryTestDriver) AssertInternalsCustomerShouldNotBeRegistered(t *testing.T, c *customer.Customer) {
	t.Helper()

	r := require.New(t)

	r.False(sliceContainsCustomer(td.customers, c))
	r.False(mapContainsCustomer(td.idIndex, c))
	r.False(mapContainsCustomer(td.nameIndex, c))
	r.False(mapContainsCustomer(td.emailIndex, c))
	r.False(mapContainsCustomer(td.phoneIndex, c))
}

// AssertInternalsCustomerShouldNotBeDuplicated asserts that the customer is not
// duplicated in the internal data structures. IDs are not compared.
func (td *ReferenceCustomerRepositoryTestDriver) AssertInternalsCustomerShouldNotBeDuplicated(t *testing.T, c *customer.Customer) {
	t.Helper()

	r := require.New(t)

	r.LessOrEqual(sliceCountCustomeOccurrences(td.customers, c), 1)
	r.LessOrEqual(mapCountCustomeOccurrences(td.idIndex, c), 1)
	r.LessOrEqual(mapCountCustomeOccurrences(td.nameIndex, c), 1)
	r.LessOrEqual(mapCountCustomeOccurrences(td.emailIndex, c), 1)
	r.LessOrEqual(mapCountCustomeOccurrences(td.phoneIndex, c), 1)
}

//
// Utility Functions

// customersAreSame compares two customers for equality. IDs are not compared.
func customersAreSame(c1, c2 *customer.Customer) bool {
	if c1 == c2 {
		return true
	}

	if c1.Name != c2.Name {
		return false
	}

	if c1.Email != c2.Email {
		return false
	}

	if c1.Phone != c2.Phone {
		return false
	}

	return true
}

// sliceCountCustomeOccurrences counts the occurrences of a customer in a slice.
// IDs are not compared.
func sliceCountCustomeOccurrences(s []*customer.Customer, c *customer.Customer) int {
	var count int

	for _, c1 := range s {
		if customersAreSame(c1, c) {
			count++
		}
	}

	return count
}

// sliceContainsCustomer checks if a slice of customers contains a specific
// customer. IDs are not compared.
func sliceContainsCustomer(s []*customer.Customer, c *customer.Customer) bool {
	for _, c1 := range s {
		if customersAreSame(c1, c) {
			return true
		}
	}

	return false
}

// mapCountCustomeOccurrences counts the occurrences of a customer in a map.
// IDs are not compared.
func mapCountCustomeOccurrences(m map[string]*customer.Customer, c *customer.Customer) int {
	var count int

	for _, c1 := range m {
		if customersAreSame(c1, c) {
			count++
		}
	}

	return count
}

// mapContainsCustomer checks if a map of customers contains a specific
// customer. IDs are not compared.
func mapContainsCustomer(m map[string]*customer.Customer, c *customer.Customer) bool {
	for _, c1 := range m {
		if customersAreSame(c1, c) {
			return true
		}
	}

	return false
}
