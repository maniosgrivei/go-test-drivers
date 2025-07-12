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
		r := require.New(t)

		request := RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		}

		// Ensure that the repository is clean
		customers = make([]*Customer, 0)
		idIndex = make(map[string]*Customer)
		nameIndex = make(map[string]*Customer)
		emailIndex = make(map[string]*Customer)
		phoneIndex = make(map[string]*Customer)

		id, err := Register(request)
		r.NoError(err)
		r.NotEmpty(id)

		// Asserting side-effects from inside
		customer := Customer{
			ID:    id,
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		}

		r.Contains(customers, &customer)

		r.Contains(idIndex, id)
		r.EqualValues(idIndex[id], &customer)

		r.Contains(nameIndex, dfltName)
		r.EqualValues(nameIndex[dfltName], &customer)

		r.Contains(emailIndex, dfltEmail)
		r.EqualValues(emailIndex[dfltEmail], &customer)

		r.Contains(phoneIndex, dfltPhone)
		r.EqualValues(phoneIndex[dfltPhone], &customer)
	})

	t.Run("should reject a registration with invalid data", func(t *testing.T) {
		cases := []struct {
			Case        string
			Request     RegisterRequest
			FindOnError []string
		}{
			{"when missing name", RegisterRequest{"", dfltEmail, dfltPhone}, []string{"invalid name"}},
			{"when missing email", RegisterRequest{dfltName, "", dfltPhone}, []string{"invalid email"}},
			{"when missing phone", RegisterRequest{dfltName, dfltEmail, ""}, []string{"invalid phone"}},
			{"when missing name and email", RegisterRequest{"", "", dfltPhone}, []string{"invalid name", "invalid email"}},
			{"when missing name and phone", RegisterRequest{"", dfltEmail, ""}, []string{"invalid name", "invalid phone"}},
			{"when missing email and phone", RegisterRequest{dfltName, "", ""}, []string{"invalid email", "invalid phone"}},
			{"when missing all data", RegisterRequest{"", "", ""}, []string{"invalid name", "invalid email", "invalid phone"}},
		}

		for _, c := range cases {
			t.Run(c.Case, func(t *testing.T) {
				r := require.New(t)

				// Ensure that the repository is clean
				customers = make([]*Customer, 0)
				idIndex = make(map[string]*Customer)
				nameIndex = make(map[string]*Customer)
				emailIndex = make(map[string]*Customer)
				phoneIndex = make(map[string]*Customer)

				id, err := Register(c.Request)

				r.Empty(id)
				r.Error(err)
				r.ErrorIs(err, ErrValidation)

				for _, e := range c.FindOnError {
					r.Contains(err.Error(), e)
				}
			})
		}
	})

	t.Run("should reject a registration with duplicated data", func(t *testing.T) {
		cases := []struct {
			Case        string
			Request     RegisterRequest
			FindOnError []string
		}{
			{"having same name", RegisterRequest{dfltName, "didi@dada.com", "+1 098 765 432"}, []string{"duplicated name"}},
			{"having same email", RegisterRequest{"Didi Dada", dfltEmail, "+1 098 765 432"}, []string{"duplicated email"}},
			{"having same phone", RegisterRequest{"Didi Dada", "didi@dada.com", dfltPhone}, []string{"duplicated phone"}},
			{"having same name and email", RegisterRequest{dfltName, dfltEmail, "+1 098 765 432"}, []string{"duplicated name", "duplicated email"}},
			{"having same name and phone", RegisterRequest{dfltName, "didi@dada.com", dfltPhone}, []string{"duplicated name", "duplicated phone"}},
			{"having same email and phone", RegisterRequest{"Didi Dada", dfltEmail, dfltPhone}, []string{"duplicated email", "duplicated phone"}},
			{"having same all data", RegisterRequest{dfltName, dfltEmail, dfltPhone}, []string{"duplicated name", "duplicated email", "duplicated phone"}},
		}

		// Controlling internal state directly
		customer := Customer{
			ID:    "1",
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		}

		customers = []*Customer{&customer}
		idIndex = map[string]*Customer{"1": &customer}
		nameIndex = map[string]*Customer{dfltName: &customer}
		emailIndex = map[string]*Customer{dfltEmail: &customer}
		phoneIndex = map[string]*Customer{dfltPhone: &customer}

		for _, c := range cases {
			t.Run(c.Case, func(t *testing.T) {
				r := require.New(t)

				id, err := Register(c.Request)

				r.Empty(id)
				r.Error(err)
				r.ErrorIs(err, ErrDuplication)

				for _, e := range c.FindOnError {
					r.Contains(err.Error(), e)
				}
			})
		}
	})

	t.Run("should return a generic system error on failure", func(t *testing.T) {
		r := require.New(t)

		Init()

		// Causing a problem
		customers = nil
		idIndex = nil
		nameIndex = nil
		emailIndex = nil
		phoneIndex = nil

		request := RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		}

		id, err := Register(request)

		r.Empty(id)
		r.Error(err)
		r.ErrorIs(err, ErrSystem)
		r.Contains(err.Error(), "contact support")
	})
}
