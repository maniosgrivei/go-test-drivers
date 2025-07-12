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

		request := &RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		}

		Init()

		id, err := Register(request)

		r.NoError(err)
		r.NotEmpty(id)
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
				r := require.New(t)

				Init()

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
			Request     *RegisterRequest
			FindOnError []string
		}{
			{"having same name", &RegisterRequest{dfltName, "didi@dada.com", "+1 098 765 432"}, []string{"duplicated name"}},
			{"having same email", &RegisterRequest{"Didi Dada", dfltEmail, "+1 098 765 432"}, []string{"duplicated email"}},
			{"having same phone", &RegisterRequest{"Didi Dada", "didi@dada.com", dfltPhone}, []string{"duplicated phone"}},
			{"having same name and email", &RegisterRequest{dfltName, dfltEmail, "+1 098 765 432"}, []string{"duplicated name", "duplicated email"}},
			{"having same name and phone", &RegisterRequest{dfltName, "didi@dada.com", dfltPhone}, []string{"duplicated name", "duplicated phone"}},
			{"having same email and phone", &RegisterRequest{"Didi Dada", dfltEmail, dfltPhone}, []string{"duplicated email", "duplicated phone"}},
			{"having same all data", &RegisterRequest{dfltName, dfltEmail, dfltPhone}, []string{"duplicated name", "duplicated email", "duplicated phone"}},
		}

		request := &RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		}

		Init()

		id, err := Register(request)
		require.NoError(t, err)
		require.NotEmpty(t, id)

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

		// cause a problem
		customers = nil
		idIndex = nil
		nameIndex = nil
		emailIndex = nil
		phoneIndex = nil

		request := &RegisterRequest{
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
