//go:build test

package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maniosgrivei/go-test-drivers/customer"
	"github.com/stretchr/testify/require"
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

	r := require.New(t)

	// Prepare request
	body, err := json.Marshal(request)
	r.NoError(err)

	req, err := http.NewRequest(http.MethodPost, "/customers", bytes.NewReader(body))
	r.NoError(err)

	req.Header.Set("Content-Type", "application/json")

	// Record response
	recorder := httptest.NewRecorder()
	td.restAPI.ServeHTTP(recorder, req)

	// Parse response body
	var responseBody map[string]any
	err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
	if err != nil {
		responseBody = nil // Handle cases with no or non-JSON body
	}

	result := map[string]any{
		"id":            "",
		"response_body": responseBody,
		"status_code":   recorder.Code,
		"status":        http.StatusText(recorder.Code),
	}

	if recorder.Code < http.StatusBadRequest {
		result["id"] = getIDFromResponseBody(t, responseBody)
	}

	return result
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

	r := require.New(t)

	// Check status
	expectedStatus := getExpectedStatus(t, extraParams)
	r.IsType(expectedStatus, result["status"])
	r.Equal(expectedStatus, result["status"].(string))

	// Check status code
	expectedStatusCode := getExpectedStatusCode(t, extraParams)
	r.IsType(expectedStatusCode, result["status_code"])
	r.Equal(expectedStatusCode, result["status_code"].(int))

	// Check body for ID
	r.Contains(result, "response_body")
	r.IsType(map[string]any{}, result["response_body"])
	responseBody := result["response_body"].(map[string]any)
	r.NotEmpty(getIDFromResponseBody(t, responseBody))
}

// AssertRegistrationShouldFail asserts that the HTTP response indicates a failure.
func (td *CustomerRESTAPIHandlerTestDriver) AssertRegistrationShouldFail(
	t *testing.T,
	result map[string]any,
	extraParams map[string]any,
) {
	t.Helper()

	r := require.New(t)

	// Check status
	expectedStatus := getExpectedStatus(t, extraParams)
	r.IsType(expectedStatus, result["status"])
	r.Equal(expectedStatus, result["status"].(string))

	// Check status code
	expectedStatusCode := getExpectedStatusCode(t, extraParams)
	r.IsType(expectedStatusCode, result["status_code"])
	r.Equal(expectedStatusCode, result["status_code"].(int))
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

	r := require.New(t)

	td.AssertRegistrationShouldFail(t, result, extraParams)

	// Check error message in body
	r.Contains(result, "response_body")
	r.IsType(map[string]any{}, result["response_body"])
	responseBody := result["response_body"].(map[string]any)

	r.Contains(responseBody, "error")
	r.IsType("", responseBody["error"])
	errorMessage := responseBody["error"].(string)

	r.NotEmpty(errorMessage)
	for _, msg := range targetMessages {
		r.Contains(errorMessage, msg)
	}
}

func getIDFromResponseBody(t *testing.T, responseBody map[string]any) string {
	t.Helper()

	r := require.New(t)

	r.Contains(responseBody, "id")
	r.IsType("", responseBody["id"])

	return responseBody["id"].(string)
}

// getExpectedStatus extracts the expected HTTP status from the extra parameters.
func getExpectedStatus(t *testing.T, extraParams map[string]any) string {
	t.Helper()

	r := require.New(t)

	r.Contains(extraParams, "http_response")
	httpResponse, ok := extraParams["http_response"].(map[string]any)
	r.True(ok)
	r.Contains(httpResponse, "status")
	status, ok := httpResponse["status"].(string)
	r.True(ok)

	return status
}

// getExpectedStatusCode extracts the expected HTTP status code from the extra parameters.
func getExpectedStatusCode(t *testing.T, extraParams map[string]any) int {
	t.Helper()

	r := require.New(t)

	r.Contains(extraParams, "http_response")
	httpResponse, ok := extraParams["http_response"].(map[string]any)
	r.True(ok)
	r.Contains(httpResponse, "status_code")
	statusCode, ok := httpResponse["status_code"].(int)
	r.True(ok)

	return statusCode
}
