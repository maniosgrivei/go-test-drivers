package customer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	dfltName  = "John Due"
	dfltEmail = "john.due@somecompany.com"
	dfltPhone = "+1 234 567 890"
)

func TestRegisterCustomer(t *testing.T) {
	t.Run("should register a customer with valid data", func(t *testing.T) {
		// Given that
		ArrangeInternalsNoCustomerIsRegistered(t)

		// When we
		id, err := ActTryToRegisterACustomer(t, &RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})
		// with valid data

		// Then the
		AssertRegistrationShouldSucceed(t, id, err)

		// And the
		AssertInternalsCustomerShouldBeProperlyRegistered(t, &Customer{
			ID:    id,
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})
	})

	t.Run("should reject a registration with invalid data", func(t *testing.T) {
		cases := []struct {
			Case        string
			Request     *RegisterRequest
			FindOnError []string
		}{
			{"when missing name", &RegisterRequest{"", dfltEmail, dfltPhone}, []string{"invalid name"}},
			{"when missing email", &RegisterRequest{dfltName, "", dfltPhone}, []string{"invalid email"}},
			{"when missing phone", &RegisterRequest{dfltName, dfltEmail, ""}, []string{"invalid phone"}},
			{"when missing name and email", &RegisterRequest{"", "", dfltPhone}, []string{"invalid name", "invalid email"}},
			{"when missing name and phone", &RegisterRequest{"", dfltEmail, ""}, []string{"invalid name", "invalid phone"}},
			{"when missing email and phone", &RegisterRequest{dfltName, "", ""}, []string{"invalid email", "invalid phone"}},
			{"when missing all data", &RegisterRequest{"", "", ""}, []string{"invalid name", "invalid email", "invalid phone"}},
		}

		for _, c := range cases {
			t.Run(c.Case, func(t *testing.T) {
				// Given the system has a
				ArrangeInternalsNoCustomerIsRegistered(t)

				// When we
				id, err := ActTryToRegisterACustomer(t, c.Request)
				// with invalid data

				// Then the
				AssertRegistrationShouldFailWithErrorAndMessage(t, id, err, ErrValidation, c.FindOnError...)

				// And the
				AssertInternalsCustomerShouldNotBeRegistered(t, &Customer{
					ID:    "1",
					Name:  c.Request.Name,
					Email: c.Request.Email,
					Phone: c.Request.Phone,
				})
			})
		}
	})

	t.Run("should reject a registration with duplicated data", func(t *testing.T) {
		cases := []struct {
			Case        string
			Request     *RegisterRequest
			FindOnError []string
		}{
			{"having same name", &RegisterRequest{dfltName, "didi@dada.com", "+1 098 765 432"}, []string{"duplicated name"}},
			{"having same email", &RegisterRequest{"Didi Dada", dfltEmail, "+1 098 765 432"}, []string{"duplicated email"}},
			{"having same phone", &RegisterRequest{"Didi Dada", "didi@dada.com", dfltPhone}, []string{"duplicated phone"}},
			{"having same name and email", &RegisterRequest{dfltName, dfltEmail, "+1 098 765 432"}, []string{"duplicated name", "duplicated email"}},
			{"having same name and phone", &RegisterRequest{dfltName, "didi@dada.com", dfltPhone}, []string{"duplicated name", "duplicated phone"}},
			{"having same email and phone", &RegisterRequest{"Didi Dada", dfltEmail, dfltPhone}, []string{"duplicated email", "duplicated phone"}},
		}

		for _, c := range cases {
			t.Run(c.Case, func(t *testing.T) {
				// Given that
				ArrangeInternalsSomeCustomersAreRegistered(t, &Customer{
					ID:    "1",
					Name:  dfltName,
					Email: dfltEmail,
					Phone: dfltPhone,
				})

				// When we
				id, err := ActTryToRegisterACustomer(t, c.Request)
				// with duplicated data

				// Then the
				AssertRegistrationShouldFailWithErrorAndMessage(t, id, err, ErrDuplication, c.FindOnError...)

				// And the
				AssertInternalsCustomerShouldNotBeRegistered(t, &Customer{
					ID:    "2",
					Name:  c.Request.Name,
					Email: c.Request.Email,
					Phone: c.Request.Phone,
				})
			})
		}
	})

	t.Run("should not register the same user twice", func(t *testing.T) {
		// Given the system
		ArrangeInternalsSomeCustomersAreRegistered(t, &Customer{
			ID:    "1",
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})

		// When we
		id, err := ActTryToRegisterACustomer(t, &RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})
		// twice

		// Then the
		AssertRegistrationShouldFailWithErrorAndMessage(t, id, err, ErrDuplication, "duplicated name", "duplicated email", "duplicated phone")

		// And
		AssertInternalsCustomerShouldNotBeDuplicated(t, &Customer{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})
	})

	t.Run("should return a generic system error on failure", func(t *testing.T) {
		// Given the system has
		ArrangeInternalsNoCustomerIsRegistered(t)

		// And
		ArrangeInternalsSomethingCausingAProblem(t)

		// When we
		id, err := ActTryToRegisterACustomer(t, &RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})
		// with valid data

		// Them the
		AssertRegistrationShouldFailWithErrorAndMessage(t, id, err, ErrSystem, "contact support")

		// And the
		AssertInternalsCustomerShouldNotBeRegistered(t, &Customer{
			ID:    "1",
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})
	})
}

//
// Test Helpers
//
//

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
