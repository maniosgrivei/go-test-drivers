//go:build test

package badgerpoc

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/maniosgrivei/go-test-drivers/customer"
	"github.com/stretchr/testify/require"
)

// BadgerCustomerRepositoryTestDriver is the test driver for BadgerCustomerRepository.
type BadgerCustomerRepositoryTestDriver struct {
	*BadgerCustomerRepository
}

// Ensure BadgerCustomerRepositoryTestDriver implements the CustomerRepositoryTestDriver interface.
var _ customer.CustomerRepositoryTestDriver = (*BadgerCustomerRepositoryTestDriver)(nil)

// NewBadgerCustomerRepositoryTestDriver creates a new test driver for the BadgerCustomerRepository.
func NewBadgerCustomerRepositoryTestDriver(repository *BadgerCustomerRepository) *BadgerCustomerRepositoryTestDriver {
	return &BadgerCustomerRepositoryTestDriver{
		BadgerCustomerRepository: repository,
	}
}

//
// Arrange

// ArrangeInternalsNoCustomerIsRegistered clears all data from the Badger database.
func (td *BadgerCustomerRepositoryTestDriver) ArrangeInternalsNoCustomerIsRegistered(t *testing.T) {
	t.Helper()
	require.NoError(t, td.db.DropAll())
}

// ArrangeInternalsSomeCustomersAreRegistered populates the database with a given list of customers.
func (td *BadgerCustomerRepositoryTestDriver) ArrangeInternalsSomeCustomersAreRegistered(t *testing.T, cs []*customer.Customer) {
	t.Helper()
	td.ArrangeInternalsNoCustomerIsRegistered(t)
	for _, c := range cs {
		err := td.Save(c)
		require.NoError(t, err)
	}
}

// ArrangeInternalsSomethingCausingAProblem simulates a system error by closing the database connection.
func (td *BadgerCustomerRepositoryTestDriver) ArrangeInternalsSomethingCausingAProblem(t *testing.T) {
	t.Helper()
	require.NoError(t, td.Close())
}

//
// Assert

// AssertInternalsCustomerShouldBeProperlyRegistered checks that the customer exists and is stored correctly.
func (td *BadgerCustomerRepositoryTestDriver) AssertInternalsCustomerShouldBeProperlyRegistered(t *testing.T, c *customer.Customer) {
	t.Helper()
	r := require.New(t)

	err := td.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getIDKey(c.ID))
		if err != nil {
			return fmt.Errorf("could not find customer with ID %s: %w", c.ID, err)
		}

		var foundCustomer customer.Customer
		err = item.Value(func(val []byte) error {
			return gob.NewDecoder(bytes.NewReader(val)).Decode(&foundCustomer)
		})
		if err != nil {
			return fmt.Errorf("failed to decode customer: %w", err)
		}

		r.Equal(c, &foundCustomer)
		return nil
	})

	r.NoError(err)
}

// AssertInternalsCustomerShouldNotBeRegistered checks that a customer with the given details does not exist.
func (td *BadgerCustomerRepositoryTestDriver) AssertInternalsCustomerShouldNotBeRegistered(t *testing.T, c *customer.Customer) {
	t.Helper()
	r := require.New(t)

	err := td.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(getIDKey(c.ID))
		r.ErrorIs(err, badger.ErrKeyNotFound)

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefixID); it.ValidForPrefix(prefixID); it.Next() {
			item := it.Item()
			var foundCustomer customer.Customer
			err := item.Value(func(val []byte) error {
				return gob.NewDecoder(bytes.NewReader(val)).Decode(&foundCustomer)
			})
			if err != nil {
				return err
			}
			if foundCustomer.Name == c.Name && foundCustomer.Email == c.Email && foundCustomer.Phone == c.Phone {
				return errors.New("customer should not be found in the database")
			}
		}
		return nil
	})

	r.NoError(err)
}

// AssertInternalsCustomerShouldNotBeDuplicated iterates through all customers to ensure no duplicates exist.
func (td *BadgerCustomerRepositoryTestDriver) AssertInternalsCustomerShouldNotBeDuplicated(t *testing.T, c *customer.Customer) {
	t.Helper()
	r := require.New(t)
	var count int

	err := td.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefixID); it.ValidForPrefix(prefixID); it.Next() {
			item := it.Item()
			var foundCustomer customer.Customer
			err := item.Value(func(val []byte) error {
				return gob.NewDecoder(bytes.NewReader(val)).Decode(&foundCustomer)
			})
			if err != nil {
				return err
			}
			if foundCustomer.Name == c.Name || foundCustomer.Email == c.Email || foundCustomer.Phone == c.Phone {
				count++
			}
		}
		return nil
	})

	r.NoError(err)
	r.LessOrEqual(count, 1, "customer should not be duplicated in the database")
}
