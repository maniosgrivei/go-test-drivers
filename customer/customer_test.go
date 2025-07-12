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

		// Arrange: Set the desired initial state.
		customers = make([]*Customer, 0)
		idIndex = make(map[string]*Customer)
		nameIndex = make(map[string]*Customer)
		emailIndex = make(map[string]*Customer)
		phoneIndex = make(map[string]*Customer)

		// Act: Perform an action on the system.
		request := &RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		}

		id, err := Register(request)

		// Assert: Check the results.
		r.NoError(err)
		r.NotEmpty(id)

		// Assert: Check the side-effects.
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

				// Arrange: Set the desired initial state.
				customers = make([]*Customer, 0)
				idIndex = make(map[string]*Customer)
				nameIndex = make(map[string]*Customer)
				emailIndex = make(map[string]*Customer)
				phoneIndex = make(map[string]*Customer)

				// Act: Perform an action on the system.
				id, err := Register(c.Request)

				// Assert: Check the results.
				r.Empty(id)
				r.Error(err)
				r.ErrorIs(err, ErrValidation)

				for _, e := range c.FindOnError {
					r.Contains(err.Error(), e)
				}

				// Assert: Check the side-effects.
				r.Len(customers, 0)
				r.Len(idIndex, 0)
				r.Len(nameIndex, 0)
				r.Len(emailIndex, 0)
				r.Len(phoneIndex, 0)
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

		// Arrange: Set the desired initial state.
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

				// Act: Perform an action on the system.
				id, err := Register(c.Request)

				// Assert: Check the results.
				r.Empty(id)
				r.Error(err)
				r.ErrorIs(err, ErrDuplication)

				for _, e := range c.FindOnError {
					r.Contains(err.Error(), e)
				}

				// Assert: Check the side-effects.
				assertAreNotTheSame := func(a *Customer, b *RegisterRequest) {
					r.False(a.Name == b.Name && a.Email == b.Email && a.Phone == b.Phone)
				}

				for _, rc := range customers {
					assertAreNotTheSame(rc, c.Request)
				}

				if rc, found := nameIndex[c.Request.Name]; found {
					assertAreNotTheSame(rc, c.Request)
				}

				if rc, found := emailIndex[c.Request.Email]; found {
					assertAreNotTheSame(rc, c.Request)
				}

				if rc, found := phoneIndex[c.Request.Phone]; found {
					assertAreNotTheSame(rc, c.Request)
				}
			})
		}
	})

	t.Run("should not register the same user twice", func(t *testing.T) {
		r := require.New(t)

		// Arrange: Set the desired initial state.
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

		request := &RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		}

		id, err := Register(request)

		// Assert: Check the results.
		r.Empty(id)
		r.Error(err)
		r.ErrorIs(err, ErrDuplication)

		for _, e := range []string{"duplicated name", "duplicated email", "duplicated phone"} {
			r.Contains(err.Error(), e)
		}

		// Assert: Check the side-effects.
		r.Len(customers, 1)
		r.EqualValues(customers[0], &customer)

		r.Len(idIndex, 1)
		r.Contains(idIndex, customer.ID)
		r.EqualValues(idIndex[customer.ID], &customer)

		r.Len(nameIndex, 1)
		r.Contains(nameIndex, customer.Name)
		r.EqualValues(nameIndex[customer.Name], &customer)

		r.Len(emailIndex, 1)
		r.Contains(emailIndex, customer.Email)
		r.EqualValues(emailIndex[customer.Email], &customer)

		r.Len(phoneIndex, 1)
		r.Contains(phoneIndex, customer.Phone)
		r.EqualValues(phoneIndex[customer.Phone], &customer)
	})

	t.Run("should return a generic system error on failure", func(t *testing.T) {
		r := require.New(t)

		Init()

		// Arrange: Set the desired initial state.
		customers = nil
		idIndex = nil
		nameIndex = nil
		emailIndex = nil
		phoneIndex = nil

		// Act: Perform an action on the system.
		request := &RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		}

		id, err := Register(request)

		// Assert: Check the results.
		r.Empty(id)
		r.Error(err)
		r.ErrorIs(err, ErrSystem)
		r.Contains(err.Error(), "contact support")

		// Assert: Check the side-effects.
		customer := Customer{
			ID:    "1",
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		}

		r.NotContains(customers, &customer)
		r.NotContains(idIndex, id)
		r.NotContains(nameIndex, dfltName)
		r.NotContains(emailIndex, dfltEmail)
		r.NotContains(phoneIndex, dfltPhone)
	})
}
