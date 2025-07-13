//go:build test

package customer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

//
// Arrange

// ArrangeInternalsNoCustomerIsRegistered initializes the repository to a clean state.
func ArrangeInternalsNoCustomerIsRegistered(t *testing.T) {
	t.Helper()

	customers = make([]*Customer, 0)
	idIndex = make(map[string]*Customer)
	nameIndex = make(map[string]*Customer)
	emailIndex = make(map[string]*Customer)
	phoneIndex = make(map[string]*Customer)
}

// ArrangeInternalsSomeCustomersAreRegistered populates the repository with the
// given customers.
func ArrangeInternalsSomeCustomersAreRegistered(t *testing.T, customer ...*Customer) {
	t.Helper()

	ArrangeInternalsNoCustomerIsRegistered(t)

	for _, c := range customer {
		customers = append(customers, c)
		idIndex[c.ID] = c
		nameIndex[c.Name] = c
		emailIndex[c.Email] = c
		phoneIndex[c.Phone] = c
	}
}

// ArrangeInternalsSomethingCausingAProblem the internal state in a way that
// function calls will results in a system error.
func ArrangeInternalsSomethingCausingAProblem(t *testing.T) {
	t.Helper()

	customers = nil
	idIndex = nil
	nameIndex = nil
	emailIndex = nil
	phoneIndex = nil
}

//
// Act

// ActTryToRegisterACustomer registers a customer in the system by acting as a user.
func ActTryToRegisterACustomer(t *testing.T, request *RegisterRequest) (id string, err error) {
	t.Helper()

	return Register(request)
}

//
// Assert

// AssertRegistrationShouldSucceed asserts that the registration was successful.
func AssertRegistrationShouldSucceed(t *testing.T, id string, err error) {
	t.Helper()

	r := require.New(t)

	r.NoError(err)
	r.NotEmpty(id)
}

// AssertRegistrationShouldFail asserts that the registration failed.
func AssertRegistrationShouldFail(t *testing.T, id string, err error) {
	t.Helper()

	r := require.New(t)

	r.Empty(id)
	r.Error(err)
}

// AssertRegistrationShouldFailWithError asserts that the registration failed
// with the given error.
func AssertRegistrationShouldFailWithError(t *testing.T, id string, err error, targetError error) {
	t.Helper()

	AssertRegistrationShouldFail(t, id, err)

	require.ErrorIs(t, err, targetError)
}

// AssertRegistrationShouldFailWithMessage asserts that the registration failed
// with the given message(s).
func AssertRegistrationShouldFailWithMessage(t *testing.T, id string, err error, targetMessage ...string) {
	t.Helper()

	AssertRegistrationShouldFail(t, id, err)

	for _, msg := range targetMessage {
		require.Contains(t, err.Error(), msg)
	}
}

// AssertRegistrationShouldFailWithErrorAndMessage asserts that the
// registration failed with the given error and message(s).
func AssertRegistrationShouldFailWithErrorAndMessage(t *testing.T, id string, err error, targetError error, targetMessage ...string) {
	t.Helper()

	AssertRegistrationShouldFailWithError(t, id, err, targetError)

	AssertRegistrationShouldFailWithMessage(t, id, err, targetMessage...)
}

// AssertInternalsCustomerShouldBeProperlyRegistered asserts that the customer
// is properly registered in the internal data structures.
func AssertInternalsCustomerShouldBeProperlyRegistered(t *testing.T, customer *Customer) {
	t.Helper()

	r := require.New(t)

	r.True(sliceContainsCustomer(customers, customer))

	r.Contains(idIndex, customer.ID)
	r.True(customersAreSame(idIndex[customer.ID], customer))

	r.Contains(nameIndex, customer.Name)
	r.True(customersAreSame(nameIndex[customer.Name], customer))

	r.Contains(emailIndex, customer.Email)
	r.True(customersAreSame(emailIndex[customer.Email], customer))

	r.Contains(phoneIndex, customer.Phone)
	r.True(customersAreSame(phoneIndex[customer.Phone], customer))
}

// AssertInternalsCustomerShouldNotBeRegistered asserts that the customer is not
// properly registered in the internal data structures.
func AssertInternalsCustomerShouldNotBeRegistered(t *testing.T, customer *Customer) {
	t.Helper()

	r := require.New(t)

	r.False(sliceContainsCustomer(customers, customer))
	r.False(mapContainsCustomer(idIndex, customer))
	r.False(mapContainsCustomer(nameIndex, customer))
	r.False(mapContainsCustomer(emailIndex, customer))
	r.False(mapContainsCustomer(phoneIndex, customer))
}

// AssertInternalsCustomerShouldNotBeDuplicated asserts that the customer is not
// duplicated in the internal data structures. IDs are not compared.
func AssertInternalsCustomerShouldNotBeDuplicated(t *testing.T, customer *Customer) {
	t.Helper()

	r := require.New(t)

	r.LessOrEqual(sliceCountCustomeOccurrences(customers, customer), 1)
	r.LessOrEqual(mapCountCustomeOccurrences(idIndex, customer), 1)
	r.LessOrEqual(mapCountCustomeOccurrences(nameIndex, customer), 1)
	r.LessOrEqual(mapCountCustomeOccurrences(emailIndex, customer), 1)
	r.LessOrEqual(mapCountCustomeOccurrences(phoneIndex, customer), 1)
}

//
// Utility Functions

// customersAreSame compares two customers for equality. IDs are not compared.
func customersAreSame(c1, c2 *Customer) bool {
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
func sliceCountCustomeOccurrences(s []*Customer, customer *Customer) int {
	var count int

	for _, c := range s {
		if customersAreSame(c, customer) {
			count++
		}
	}

	return count
}

// sliceContainsCustomer checks if a slice of customers contains a specific
// customer. IDs are not compared.
func sliceContainsCustomer(s []*Customer, customer *Customer) bool {
	for _, c := range s {
		if customersAreSame(c, customer) {
			return true
		}
	}

	return false
}

// mapCountCustomeOccurrences counts the occurrences of a customer in a map.
// IDs are not compared.
func mapCountCustomeOccurrences(m map[string]*Customer, customer *Customer) int {
	var count int

	for _, c := range m {
		if customersAreSame(c, customer) {
			count++
		}
	}

	return count
}

// mapContainsCustomer checks if a map of customers contains a specific
// customer. IDs are not compared.
func mapContainsCustomer(m map[string]*Customer, customer *Customer) bool {
	for _, c := range m {
		if customersAreSame(c, customer) {
			return true
		}
	}

	return false
}
