//go:build test

package sqlitepoc

import (
	"database/sql"
	"testing"

	"github.com/maniosgrivei/go-test-drivers/customer"
	"github.com/stretchr/testify/require"
)

// SQLiteCustomerRepositoryTestDriver is the test driver for SQLiteCustomerRepository.
type SQLiteCustomerRepositoryTestDriver struct {
	*SQLiteCustomerRepository
}

// Ensure SQLiteCustomerRepositoryTestDriver implements the CustomerRepositoryTestDriver interface.
var _ customer.CustomerRepositoryTestDriver = (*SQLiteCustomerRepositoryTestDriver)(nil)

// NewSQLiteCustomerRepositoryTestDriver creates a new test driver for the SQLiteCustomerRepository.
func NewSQLiteCustomerRepositoryTestDriver(repository *SQLiteCustomerRepository) *SQLiteCustomerRepositoryTestDriver {
	return &SQLiteCustomerRepositoryTestDriver{
		SQLiteCustomerRepository: repository,
	}
}

//
// Arrange

// ArrangeInternalsNoCustomerIsRegistered clears all customer data from the database.
func (td *SQLiteCustomerRepositoryTestDriver) ArrangeInternalsNoCustomerIsRegistered(t *testing.T) {
	t.Helper()
	td.db.MustExec("DELETE FROM customers")
}

// ArrangeInternalsSomeCustomersAreRegistered populates the database with a given list of customers.
func (td *SQLiteCustomerRepositoryTestDriver) ArrangeInternalsSomeCustomersAreRegistered(t *testing.T, cs []*customer.Customer) {
	t.Helper()
	td.ArrangeInternalsNoCustomerIsRegistered(t)

	for _, c := range cs {
		_, err := td.db.Exec("INSERT INTO customers (id, name, email, phone) VALUES (?, ?, ?, ?)", c.ID, c.Name, c.Email, c.Phone)
		require.NoError(t, err)
	}
}

// ArrangeInternalsSomethingCausingAProblem simulates a system error by closing the database connection.
func (td *SQLiteCustomerRepositoryTestDriver) ArrangeInternalsSomethingCausingAProblem(t *testing.T) {
	t.Helper()
	require.NoError(t, td.db.Close())
}

//
// Assert

// AssertInternalsCustomerShouldBeProperlyRegistered checks that the customer exists in the database.
func (td *SQLiteCustomerRepositoryTestDriver) AssertInternalsCustomerShouldBeProperlyRegistered(t *testing.T, c *customer.Customer) {
	t.Helper()
	r := require.New(t)

	var foundCustomer customer.Customer
	err := td.db.Get(&foundCustomer, "SELECT * FROM customers WHERE id = ?", c.ID)
	r.NoError(err, "customer with ID %s should be found", c.ID)
	r.Equal(c.Name, foundCustomer.Name)
	r.Equal(c.Email, foundCustomer.Email)
	r.Equal(c.Phone, foundCustomer.Phone)
}

// AssertInternalsCustomerShouldNotBeRegistered checks that the customer does not exist in the database.
func (td *SQLiteCustomerRepositoryTestDriver) AssertInternalsCustomerShouldNotBeRegistered(t *testing.T, c *customer.Customer) {
	t.Helper()
	r := require.New(t)

	if c.ID == "" {
		err := td.db.Get(&customer.Customer{}, "SELECT * FROM customers WHERE name = ? AND email = ? AND phone = ?", c.Name, c.Email, c.Phone)

		r.ErrorIs(err, sql.ErrNoRows, "customer should not be found in the database")

		return
	}

	err := td.db.Get(&customer.Customer{}, "SELECT * FROM customers WHERE id = ? AND name = ? AND email = ? AND phone = ?", c.ID, c.Name, c.Email, c.Phone)

	r.ErrorIs(err, sql.ErrNoRows, "customer should not be found in the database")
}

// AssertInternalsCustomerShouldNotBeDuplicated checks that no more than one record matches the customer's details.
func (td *SQLiteCustomerRepositoryTestDriver) AssertInternalsCustomerShouldNotBeDuplicated(t *testing.T, c *customer.Customer) {
	t.Helper()
	r := require.New(t)

	var count int
	query := "SELECT count(*) FROM customers WHERE name = ? OR email = ? OR phone = ?"
	err := td.db.Get(&count, query, c.Name, c.Email, c.Phone)
	r.NoError(err)
	r.LessOrEqual(count, 1, "customer should not be duplicated in the database")
}
