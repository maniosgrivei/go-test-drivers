package badgerpoc

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/maniosgrivei/go-test-drivers/customer"
)

// key prefixes to simulate tables and indexes
var (
	prefixID    = []byte("#CS>ID>")
	prefixName  = []byte("#CS>NM>")
	prefixEmail = []byte("#CS>EM>")
	prefixPhone = []byte("#CS>PH>")
)

// BadgerCustomerRepository is an implementation of CustomerRepository that uses an in-memory Badger database.
type BadgerCustomerRepository struct {
	db *badger.DB
}

// Ensure BadgerCustomerRepository implements the CustomerRepository interface.
var _ customer.CustomerRepository = (*BadgerCustomerRepository)(nil)

// NewBadgerCustomerRepository creates and initializes a new BadgerCustomerRepository.
func NewBadgerCustomerRepository() *BadgerCustomerRepository {
	opts := badger.DefaultOptions("").WithInMemory(true)
	opts.Logger = nil // Suppress verbose logging during tests
	db, err := badger.Open(opts)
	if err != nil {
		panic(fmt.Sprintf("failed to open badger database: %v", err))
	}
	return &BadgerCustomerRepository{db: db}
}

// Save adds a new customer to the repository, checking for duplicates first.
func (r *BadgerCustomerRepository) Save(c *customer.Customer) error {
	if r.db == nil {
		return customer.ErrSystem
	}

	err := r.db.Update(func(txn *badger.Txn) error {
		// Check for duplications before inserting.
		if err := checkDuplication(txn, c); err != nil {
			return fmt.Errorf("%w: %w", customer.ErrDuplication, err)
		}

		// Encode the customer struct into bytes.
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(c); err != nil {
			return fmt.Errorf("failed to encode customer: %w", err)
		}

		// Save the customer data.
		key := getIDKey(c.ID)
		if err := txn.Set(key, buf.Bytes()); err != nil {
			return err
		}

		// Save uniqueness indexes.
		if err := txn.Set(getNameKey(c.Name), key); err != nil {
			return err
		}
		if err := txn.Set(getEmailKey(c.Email), key); err != nil {
			return err
		}
		if err := txn.Set(getPhoneKey(c.Phone), key); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// Check if the error is a known duplication or system error.
		if errors.Is(err, customer.ErrDuplication) {
			return err
		}
		return fmt.Errorf("%w: %w", customer.ErrSystem, err)
	}

	return nil
}

// Close closes the Badger database connection.
func (r *BadgerCustomerRepository) Close() error {
	return r.db.Close()
}

// checkDuplication checks if a customer with the same id, name, email, or phone already exists within a transaction.
func checkDuplication(txn *badger.Txn, c *customer.Customer) error {
	var errs []error

	if _, err := txn.Get(getIDKey(c.ID)); !errors.Is(err, badger.ErrKeyNotFound) {
		errs = append(errs, fmt.Errorf("duplicated id: '%s'", c.ID))
	}
	if _, err := txn.Get(getNameKey(c.Name)); !errors.Is(err, badger.ErrKeyNotFound) {
		errs = append(errs, fmt.Errorf("duplicated name: '%s'", c.Name))
	}
	if _, err := txn.Get(getEmailKey(c.Email)); !errors.Is(err, badger.ErrKeyNotFound) {
		errs = append(errs, fmt.Errorf("duplicated email: '%s'", c.Email))
	}
	if _, err := txn.Get(getPhoneKey(c.Phone)); !errors.Is(err, badger.ErrKeyNotFound) {
		errs = append(errs, fmt.Errorf("duplicated phone: '%s'", c.Phone))
	}

	return errors.Join(errs...)
}

// getIDKey generates the database key for a customer.
func getIDKey(id string) []byte {
	return append(prefixID, []byte(id)...)
}

// getNameKey generates the database key for the name index.
func getNameKey(name string) []byte {
	return append(prefixName, []byte(name)...)
}

// getEmailKey generates the database key for the email index.
func getEmailKey(email string) []byte {
	return append(prefixEmail, []byte(email)...)
}

// getPhoneKey generates the database key for the phone index.
func getPhoneKey(phone string) []byte {
	return append(prefixPhone, []byte(phone)...)
}
