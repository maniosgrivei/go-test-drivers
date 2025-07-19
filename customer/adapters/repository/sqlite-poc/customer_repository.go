package sqlitepoc

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/maniosgrivei/go-test-drivers/customer"
	sqlite "github.com/mattn/go-sqlite3"
)

// SQLiteCustomerRepository is an implementation of CustomerRepository that uses an in-memory SQLite database.
type SQLiteCustomerRepository struct {
	db *sqlx.DB
}

// Ensure SQLiteCustomerRepository implements the CustomerRepository interface.
var _ customer.CustomerRepository = (*SQLiteCustomerRepository)(nil)

// NewSQLiteCustomerRepository creates and initializes a new SQLiteCustomerRepository.
// It sets up an in-memory SQLite database and creates the necessary customer table.
func NewSQLiteCustomerRepository() *SQLiteCustomerRepository {
	db := sqlx.MustConnect("sqlite3", ":memory:")

	schema := `
    CREATE TABLE customers (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL UNIQUE,
        email TEXT NOT NULL UNIQUE,
        phone TEXT NOT NULL UNIQUE
    );`
	db.MustExec(schema)

	return &SQLiteCustomerRepository{db: db}
}

// Save adds a new customer to the repository after checking for duplications.
func (r *SQLiteCustomerRepository) Save(c *customer.Customer) error {
	if r.db == nil {
		return customer.ErrSystem
	}

	query := "INSERT INTO customers (id, name, email, phone) VALUES (?, ?, ?, ?)"
	_, err := r.db.Exec(query, c.ID, c.Name, c.Email, c.Phone)
	if err != nil {
		var sqliteErr sqlite.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite.ErrConstraint {
				// In order to be complient with the acceptance criteria we need
				// to check duplications becouse the SQLite engine returns only
				// the first error it found.
				if err := r.checkDuplication(c); err != nil {
					return fmt.Errorf("%w: %w", customer.ErrDuplication, err)
				}
			}
		}

		return fmt.Errorf("%w: %w", customer.ErrSystem, err)
	}

	return nil
}

// checkDuplication checks if a customer with the same name, email, or phone already exists.
func (r *SQLiteCustomerRepository) checkDuplication(c *customer.Customer) error {
	var errs []error
	var count int

	if err := r.db.Get(&count, "SELECT count(*) FROM customers WHERE id = ?", c.ID); err == nil && count > 0 {
		errs = append(errs, fmt.Errorf("duplicated id: '%s'", c.Name))
	}

	if err := r.db.Get(&count, "SELECT count(*) FROM customers WHERE name = ?", c.Name); err == nil && count > 0 {
		errs = append(errs, fmt.Errorf("duplicated name: '%s'", c.Name))
	}

	if err := r.db.Get(&count, "SELECT count(*) FROM customers WHERE email = ?", c.Email); err == nil && count > 0 {
		errs = append(errs, fmt.Errorf("duplicated email: '%s'", c.Email))
	}

	if err := r.db.Get(&count, "SELECT count(*) FROM customers WHERE phone = ?", c.Phone); err == nil && count > 0 {
		errs = append(errs, fmt.Errorf("duplicated phone: '%s'", c.Phone))
	}

	return errors.Join(errs...)
}
