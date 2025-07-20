//go:build test

package rest

import (
	"testing"

	"github.com/maniosgrivei/go-test-drivers/customer"
)

// CustomerRESTAPIHandlerTestDriver is the test driver for the
// CustomerRESTAPIHandler.
type CustomerRESTAPIHandlerTestDriver struct {
	restAPI *CustomerRESTAPIHandler
}

var _ customer.CustomerUpperLayerTestDriver = (*CustomerRESTAPIHandlerTestDriver)(nil)

// NewCustomerRESTAPIHandlerTestDriver creates a new test driver for the REST
// customer API handler.
func NewCustomerRESTAPIHandlerTestDriver(restAPI *CustomerRESTAPIHandler) *CustomerRESTAPIHandlerTestDriver {
	return &CustomerRESTAPIHandlerTestDriver{
		restAPI: restAPI,
	}
}

//
// Act

// ActTryToRegisterACustomer simulates an HTTP request to the registration endpoint.
func (td *CustomerRESTAPIHandlerTestDriver) ActTryToRegisterACustomer(
	t *testing.T,
	request map[string]any,
	extraParams map[string]any,
) map[string]any {
	t.Helper()

	t.Log("ActTryToRegisterACustomer not implemented yet")

	t.Fail()

	return nil
}

//
// Assert

// AssertRegistrationShouldSucceed asserts that the HTTP response indicates a successful registration.
func (td *CustomerRESTAPIHandlerTestDriver) AssertRegistrationShouldSucceed(
	t *testing.T,
	result map[string]any,
	extraParams map[string]any,
) {
	t.Helper()

	t.Log("AssertRegistrationShouldSucceed not implemented yet")

	t.Fail()
}

// AssertRegistrationShouldFail asserts that the HTTP response indicates a failure.
func (td *CustomerRESTAPIHandlerTestDriver) AssertRegistrationShouldFail(
	t *testing.T,
	result map[string]any,
	extraParams map[string]any,
) {
	t.Helper()

	t.Log("AssertRegistrationShouldFail not implemented yet")

	t.Fail()
}

// AssertRegistrationShouldFailWithMessage asserts that the HTTP response indicates a failure
// with specific status codes and error messages.
func (td *CustomerRESTAPIHandlerTestDriver) AssertRegistrationShouldFailWithMessage(
	t *testing.T,
	result map[string]any,
	extraParams map[string]any,
	targetMessages ...string,
) {
	t.Helper()

	t.Log("AssertRegistrationShouldFailWithMessage not implemented yet")

	t.Fail()
}
