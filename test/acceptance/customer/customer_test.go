package customer_test

import (
	"testing"

	"github.com/maniosgrivei/go-test-drivers/customer"
)

const (
	dfltName  = "John Due"
	dfltEmail = "john.due@somecompany.com"
	dfltPhone = "+1 234 567 890"
)

func TestRegisterCustomer(t *testing.T) {
	t.Run("should register a customer with valid data", func(t *testing.T) {
		// Given that
		customer.ArrangeInternalsNoCustomerIsRegistered(t)

		// When we
		id, err := customer.ActTryToRegisterACustomer(t, &customer.RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})
		// with valid data

		// Then the
		customer.AssertRegistrationShouldSucceed(t, id, err)

		// And the
		customer.AssertInternalsCustomerShouldBeProperlyRegistered(t, &customer.Customer{
			ID:    id,
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})
	})

	t.Run("should reject a registration with invalid data", func(t *testing.T) {
		cases := []struct {
			Case        string
			Request     *customer.RegisterRequest
			FindOnError []string
		}{
			{"when missing name", &customer.RegisterRequest{Name: "", Email: dfltEmail, Phone: dfltPhone}, []string{"invalid name"}},
			{"when missing email", &customer.RegisterRequest{Name: dfltName, Email: "", Phone: dfltPhone}, []string{"invalid email"}},
			{"when missing phone", &customer.RegisterRequest{Name: dfltName, Email: dfltEmail, Phone: ""}, []string{"invalid phone"}},
			{"when missing name and email", &customer.RegisterRequest{Name: "", Email: "", Phone: dfltPhone}, []string{"invalid name", "invalid email"}},
			{"when missing name and phone", &customer.RegisterRequest{Name: "", Email: dfltEmail, Phone: ""}, []string{"invalid name", "invalid phone"}},
			{"when missing email and phone", &customer.RegisterRequest{Name: dfltName, Email: "", Phone: ""}, []string{"invalid email", "invalid phone"}},
			{"when missing all data", &customer.RegisterRequest{Name: "", Email: "", Phone: ""}, []string{"invalid name", "invalid email", "invalid phone"}},
		}

		for _, c := range cases {
			t.Run(c.Case, func(t *testing.T) {
				// Given that
				customer.ArrangeInternalsNoCustomerIsRegistered(t)

				// When we
				id, err := customer.ActTryToRegisterACustomer(t, c.Request)
				// with invalid data

				// Then the
				customer.AssertRegistrationShouldFailWithErrorAndMessage(t, id, err, customer.ErrValidation, c.FindOnError...)

				// And the
				customer.AssertInternalsCustomerShouldNotBeRegistered(t, &customer.Customer{
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
			Request     *customer.RegisterRequest
			FindOnError []string
		}{
			{"having same name", &customer.RegisterRequest{Name: dfltName, Email: "didi@dada.com", Phone: "+1 098 765 432"}, []string{"duplicated name"}},
			{"having same email", &customer.RegisterRequest{Name: "Didi Dada", Email: dfltEmail, Phone: "+1 098 765 432"}, []string{"duplicated email"}},
			{"having same phone", &customer.RegisterRequest{Name: "Didi Dada", Email: "didi@dada.com", Phone: dfltPhone}, []string{"duplicated phone"}},
			{"having same name and email", &customer.RegisterRequest{Name: dfltName, Email: dfltEmail, Phone: "+1 098 765 432"}, []string{"duplicated name", "duplicated email"}},
			{"having same name and phone", &customer.RegisterRequest{Name: dfltName, Email: "didi@dada.com", Phone: dfltPhone}, []string{"duplicated name", "duplicated phone"}},
			{"having same email and phone", &customer.RegisterRequest{Name: "Didi Dada", Email: dfltEmail, Phone: dfltPhone}, []string{"duplicated email", "duplicated phone"}},
		}

		for _, c := range cases {
			t.Run(c.Case, func(t *testing.T) {
				// Given that
				customer.ArrangeInternalsSomeCustomersAreRegistered(t, &customer.Customer{
					ID:    "1",
					Name:  dfltName,
					Email: dfltEmail,
					Phone: dfltPhone,
				})

				// When we
				id, err := customer.ActTryToRegisterACustomer(t, c.Request)
				// with duplicated data

				// Then the
				customer.AssertRegistrationShouldFailWithErrorAndMessage(t, id, err, customer.ErrDuplication, c.FindOnError...)

				// And the
				customer.AssertInternalsCustomerShouldNotBeRegistered(t, &customer.Customer{
					ID:    "2",
					Name:  c.Request.Name,
					Email: c.Request.Email,
					Phone: c.Request.Phone,
				})
			})
		}
	})

	t.Run("should not register the same user twice", func(t *testing.T) {
		// Given that
		customer.ArrangeInternalsSomeCustomersAreRegistered(t, &customer.Customer{
			ID:    "1",
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})

		// When we
		id, err := customer.ActTryToRegisterACustomer(t, &customer.RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})
		// twice

		// Then the
		customer.AssertRegistrationShouldFailWithErrorAndMessage(t, id, err, customer.ErrDuplication, "duplicated name", "duplicated email", "duplicated phone")

		// And
		customer.AssertInternalsCustomerShouldNotBeDuplicated(t, &customer.Customer{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})
	})

	t.Run("should return a generic system error on failure", func(t *testing.T) {
		// Given that
		customer.ArrangeInternalsNoCustomerIsRegistered(t)

		// And
		customer.ArrangeInternalsSomethingCausingAProblem(t)

		// When we
		id, err := customer.ActTryToRegisterACustomer(t, &customer.RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})
		// with valid data

		// Them the
		customer.AssertRegistrationShouldFailWithErrorAndMessage(t, id, err, customer.ErrSystem, "contact support")

		// And the
		customer.AssertInternalsCustomerShouldNotBeRegistered(t, &customer.Customer{
			ID:    "1",
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		})
	})
}
