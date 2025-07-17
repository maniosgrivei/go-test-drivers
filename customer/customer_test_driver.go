//go:build test

package customer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//
// Dependencies

// CustomerRepositoryTestDriver is a test driver for the CustomerRepository.
// It provides methods for arranging, acting, and asserting on the
// CustomerRepository.
type CustomerRepositoryTestDriver interface {
	//
	// Arrange

	// ArrangeInternalsNoCustomerIsRegistered initializes the repository to a
	// clean state.
	ArrangeInternalsNoCustomerIsRegistered(t *testing.T)

	// ArrangeInternalsSomeCustomersAreRegistered populates the repository with
	// the given customers.
	ArrangeInternalsSomeCustomersAreRegistered(t *testing.T, customers []*Customer)

	// ArrangeInternalsSomethingCausingAProblem corrupts the internal state to
	// ensure subsequent function calls will result in a system error.
	ArrangeInternalsSomethingCausingAProblem(t *testing.T)

	//
	// Assert

	// AssertInternalsCustomerShouldBeProperlyRegistered asserts that the
	// customer is properly registered in the internal data structures.
	AssertInternalsCustomerShouldBeProperlyRegistered(t *testing.T, customer *Customer)

	// AssertInternalsCustomerShouldNotBeRegistered asserts that the customer is not
	// present in the internal data structures.
	AssertInternalsCustomerShouldNotBeRegistered(t *testing.T, customer *Customer)

	// AssertInternalsCustomerShouldNotBeDuplicated asserts that the customer is not
	// duplicated in the internal data structures. IDs are not compared.
	AssertInternalsCustomerShouldNotBeDuplicated(t *testing.T, customer *Customer)
}

//
// Test Driver

// CustomerServiceTestDriver is a test driver for the CustomerService.
// It provides methods for arranging, acting, and asserting on the
// CustomerService.
type CustomerServiceTestDriver struct {
	*CustomerService

	repositoryTD CustomerRepositoryTestDriver
}

// NewCustomerServiceTestDriver creates a new instance of
// CustomerServiceTestDriver.
func NewCustomerServiceTestDriver(
	customerService *CustomerService,
	repositoryTD CustomerRepositoryTestDriver,
) *CustomerServiceTestDriver {
	return &CustomerServiceTestDriver{
		CustomerService: customerService,
		repositoryTD:    repositoryTD,
	}
}

//
// Arrange

// ArrangeInternalsNoCustomerIsRegistered initializes the repository to a clean
// state.
func (td *CustomerServiceTestDriver) ArrangeInternalsNoCustomerIsRegistered(t *testing.T) {
	t.Helper()

	td.repositoryTD.ArrangeInternalsNoCustomerIsRegistered(t)
}

// ArrangeInternalsSomeCustomersAreRegistered populates the repository with the
// given customers.
//
// It looks for the following attributes in each map:
// - id: string
// - name: string
// - email: string
// - phone: string
func (td *CustomerServiceTestDriver) ArrangeInternalsSomeCustomersAreRegistered(t *testing.T, customerMaps ...map[string]any) {
	t.Helper()

	td.ArrangeInternalsNoCustomerIsRegistered(t)

	customers := make([]*Customer, len(customerMaps))
	for i, cm := range customerMaps {
		c := getCustomerFromMap(t, cm)

		if c.ID == "" {
			var err error
			c.ID, err = GenerateID(c.Name, time.Now())
			require.NoError(t, err)
			require.NotEmpty(t, c.ID)
		}

		customers[i] = c
	}

	td.repositoryTD.ArrangeInternalsSomeCustomersAreRegistered(t, customers)
}

// ArrangeInternalsSomethingCausingAProblem corrupts the internal state to
// ensure subsequent function calls will result in a system error.
func (td *CustomerServiceTestDriver) ArrangeInternalsSomethingCausingAProblem(t *testing.T) {
	t.Helper()

	td.repositoryTD.ArrangeInternalsSomethingCausingAProblem(t)
}

//
// Act

// ActTryToRegisterACustomer attempts to register a customer using data from a
// map.
//
// It looks for the following optional attributes in the `request` map:
// - name: string
// - email: string
// - phone: string
//
// It returns a map containing:
// - id: string
// - err: error
func (td *CustomerServiceTestDriver) ActTryToRegisterACustomer(t *testing.T, request map[string]any) map[string]any {
	t.Helper()

	name := getOptionalStringFromMap(t, request, "name")
	email := getOptionalStringFromMap(t, request, "email")
	phone := getOptionalStringFromMap(t, request, "phone")

	id, err := td.Register(&RegisterRequest{Name: name, Email: email, Phone: phone})

	// Update the request with the generated customer ID.
	request["id"] = id

	return map[string]any{
		"id":  id,
		"err": err,
	}
}

//
// Assert

// AssertRegistrationShouldSucceed asserts that the registration was successful.
//
// It looks for the following attributes in the `result` map:
// - id: string
// - err: error
func (td *CustomerServiceTestDriver) AssertRegistrationShouldSucceed(t *testing.T, result map[string]any) {
	t.Helper()

	r := require.New(t)

	if errVal, ok := result["err"]; ok && errVal != nil {
		r.NoError(errVal.(error))
	}

	id := getStringFromMap(t, result, "id")
	r.NotEmpty(id)
}

// AssertRegistrationShouldFail asserts that the registration failed.
//
// It looks for the following attributes in the `result` map:
// - id: string
// - err: error
func (td *CustomerServiceTestDriver) AssertRegistrationShouldFail(t *testing.T, result map[string]any) {
	t.Helper()

	r := require.New(t)

	if idVal, ok := result["id"]; ok {
		id, isString := idVal.(string)
		r.True(isString)
		r.Empty(id)
	}

	r.Contains(result, "err")
	err, ok := result["err"].(error)
	r.True(ok, "result 'err' field should be an error type")
	r.Error(err)
}

// AssertRegistrationShouldFailWithMessage asserts that the registration failed
// with the given message(s).
//
// It looks for the following attributes in the `result` map:
// - id: string
// - err: error
func (td *CustomerServiceTestDriver) AssertRegistrationShouldFailWithMessage(t *testing.T, result map[string]any, targetMessages ...string) {
	t.Helper()

	td.AssertRegistrationShouldFail(t, result)

	errorMessage := result["err"].(error).Error()
	for _, msg := range targetMessages {
		require.Contains(t, errorMessage, msg)
	}
}

// AssertInternalsCustomerShouldBeProperlyRegistered asserts that the customer
// is properly registered in the internal data structures.
//
// It looks for the following attributes in the `customerData` map:
// - id: string
// - name: string
// - email: string
// - phone: string
func (td *CustomerServiceTestDriver) AssertInternalsCustomerShouldBeProperlyRegistered(t *testing.T, customerData map[string]any) {
	t.Helper()

	customer := getCustomerFromMap(t, customerData)

	td.repositoryTD.AssertInternalsCustomerShouldBeProperlyRegistered(t, customer)
}

// AssertInternalsCustomerShouldNotBeRegistered asserts that the customer is not
// present in the internal data structures.
//
// It looks for the following optional attributes in the `customerData` map:
// - id: string
// - name: string
// - email: string
// - phone: string
func (td *CustomerServiceTestDriver) AssertInternalsCustomerShouldNotBeRegistered(t *testing.T, customerData map[string]any) {
	t.Helper()

	customer := getCustomerFromMap(t, customerData)

	td.repositoryTD.AssertInternalsCustomerShouldNotBeRegistered(t, customer)
}

// AssertInternalsCustomerShouldNotBeDuplicated asserts that the customer is not
// duplicated in the internal data structures. IDs are not compared.
func (td *CustomerServiceTestDriver) AssertInternalsCustomerShouldNotBeDuplicated(t *testing.T, customerData map[string]any) {
	t.Helper()

	customer := getCustomerFromMap(t, customerData)

	td.repositoryTD.AssertInternalsCustomerShouldNotBeDuplicated(t, customer)
}

//
// Internal Helpers

// getCustomerFromMap extracts a customer from a map.
//
// It looks for the following attributes in the `data` map:
// - id: string (optional)
// - name: string
// - email: string
// - phone: string
func getCustomerFromMap(t *testing.T, data map[string]any) *Customer {
	t.Helper()

	return &Customer{
		ID:    getOptionalStringFromMap(t, data, "id"),
		Name:  getStringFromMap(t, data, "name"),
		Email: getStringFromMap(t, data, "email"),
		Phone: getStringFromMap(t, data, "phone"),
	}
}

// getStringFromMap safely extracts a required string value from a map.
func getStringFromMap(t *testing.T, data map[string]any, key string) string {
	t.Helper()
	r := require.New(t)

	r.Contains(data, key, "map should contain required key '%s'", key)
	val, ok := data[key]
	r.True(ok)

	strVal, ok := val.(string)
	r.True(ok, "value for key '%s' should be a string", key)

	return strVal
}

// getOptionalStringFromMap safely extracts an optional string value from a map.
// It returns an empty string if the key does not exist.
func getOptionalStringFromMap(t *testing.T, data map[string]any, key string) string {
	t.Helper()
	r := require.New(t)

	val, ok := data[key]
	if !ok {
		return ""
	}

	strVal, ok := val.(string)
	r.True(ok, "value for key '%s' should be a string", key)

	return strVal
}
