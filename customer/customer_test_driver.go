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
// Inverse Dependencies

// CustomerUpperLayerTestDriver is a test driver for the Customer upper layer
// components. It provides methods for acting and asserting on the Customer
// Presentation layer.
type CustomerUpperLayerTestDriver interface {
	//
	// Act

	// ActTryToRegisterACustomer attempts to register a customer using data from
	// a map.
	ActTryToRegisterACustomer(t *testing.T, request map[string]any, extraArgs map[string]any) map[string]any

	//
	// Assert

	// AssertRegistrationShouldSucceed asserts that the registration was
	// successful.
	AssertRegistrationShouldSucceed(t *testing.T, result map[string]any, extraArgs map[string]any)

	// AssertRegistrationShouldFail asserts that the registration failed.
	AssertRegistrationShouldFail(t *testing.T, result map[string]any, extraArgs map[string]any)

	// AssertRegistrationShouldFailWithMessage asserts that the registration failed
	AssertRegistrationShouldFailWithMessage(t *testing.T, result map[string]any, extraArgs map[string]any, targetMessages ...string)
}

//
// Test Driver

// CustomerServiceTestDriver is a test driver for the CustomerService.
// It provides methods for arranging, acting, and asserting on the
// CustomerService.
type CustomerServiceTestDriver struct {
	*CustomerService

	repositoryTD CustomerRepositoryTestDriver
	upperLayerTD CustomerUpperLayerTestDriver
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

// NewCustomerServiceTestDriverWithPresentation creates a new instance of
// CustomerServiceTestDriver with a presentation layer test driverf.
func NewCustomerServiceTestDriverWithPresentation(
	customerService *CustomerService,
	repositoryTD CustomerRepositoryTestDriver,
	presentationTD CustomerUpperLayerTestDriver,
) *CustomerServiceTestDriver {
	return &CustomerServiceTestDriver{
		CustomerService: customerService,
		repositoryTD:    repositoryTD,
		upperLayerTD:    presentationTD,
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
func (td *CustomerServiceTestDriver) ArrangeInternalsSomeCustomersAreRegistered(
	t *testing.T,
	customerMaps ...map[string]any,
) {
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
func (td *CustomerServiceTestDriver) ActTryToRegisterACustomer(
	t *testing.T,
	request map[string]any,
	extraArgs map[string]any,
) map[string]any {
	t.Helper()

	if td.upperLayerTD != nil {
		result := td.upperLayerTD.ActTryToRegisterACustomer(t, request, extraArgs)

		// Update the request with the generated customer ID.
		request["id"] = result["id"]

		return result
	}

	name := GetOptionalStringFromMap(t, request, "name")
	email := GetOptionalStringFromMap(t, request, "email")
	phone := GetOptionalStringFromMap(t, request, "phone")

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
func (td *CustomerServiceTestDriver) AssertRegistrationShouldSucceed(
	t *testing.T,
	result map[string]any,
	extraArgs map[string]any,
) {
	t.Helper()

	if td.upperLayerTD != nil {
		td.upperLayerTD.AssertRegistrationShouldSucceed(t, result, extraArgs)
		return
	}

	r := require.New(t)

	if errVal, ok := result["err"]; ok && errVal != nil {
		r.NoError(errVal.(error))
	}

	id := GetStringFromMap(t, result, "id")
	r.NotEmpty(id)
}

// AssertRegistrationShouldFail asserts that the registration failed.
//
// It looks for the following attributes in the `result` map:
// - id: string
// - err: error
func (td *CustomerServiceTestDriver) AssertRegistrationShouldFail(
	t *testing.T,
	result map[string]any,
	extraArgs map[string]any,
) {
	t.Helper()

	if td.upperLayerTD != nil {
		td.upperLayerTD.AssertRegistrationShouldFail(t, result, extraArgs)

		return
	}

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
func (td *CustomerServiceTestDriver) AssertRegistrationShouldFailWithMessage(
	t *testing.T,
	result map[string]any,
	extraArgs map[string]any,
	targetMessages ...string,
) {
	t.Helper()

	if td.upperLayerTD != nil {
		td.upperLayerTD.AssertRegistrationShouldFailWithMessage(t, result, extraArgs, targetMessages...)

		return
	}

	td.AssertRegistrationShouldFail(t, result, extraArgs)

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
func (td *CustomerServiceTestDriver) AssertInternalsCustomerShouldBeProperlyRegistered(
	t *testing.T,
	customerData map[string]any,
) {
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
func (td *CustomerServiceTestDriver) AssertInternalsCustomerShouldNotBeRegistered(
	t *testing.T,
	customerData map[string]any,
) {
	t.Helper()

	customer := getCustomerFromMap(t, customerData)

	td.repositoryTD.AssertInternalsCustomerShouldNotBeRegistered(t, customer)
}

// AssertInternalsCustomerShouldNotBeDuplicated asserts that the customer is not
// duplicated in the internal data structures. IDs are not compared.
func (td *CustomerServiceTestDriver) AssertInternalsCustomerShouldNotBeDuplicated(
	t *testing.T,
	customerData map[string]any,
) {
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
		ID:    GetOptionalStringFromMap(t, data, "id"),
		Name:  GetStringFromMap(t, data, "name"),
		Email: GetStringFromMap(t, data, "email"),
		Phone: GetStringFromMap(t, data, "phone"),
	}
}

//
// Map Data Helpers

// GetStringFromMap safely extracts a required string value from a map.
func GetStringFromMap(t *testing.T, data map[string]any, key string) string {
	t.Helper()
	r := require.New(t)

	r.Contains(data, key, "map should contain required key '%s'", key)
	val, ok := data[key]
	r.True(ok)

	strVal, ok := val.(string)
	r.True(ok, "value for key '%s' should be a string", key)

	return strVal
}

// GetOptionalStringFromMap safely extracts an optional string value from a map.
// It returns an empty string if the key does not exist.
func GetOptionalStringFromMap(t *testing.T, data map[string]any, key string) string {
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
