package customer_test

import (
	"os"
	"testing"

	"github.com/maniosgrivei/go-test-drivers/customer"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRegisterCustomer(t *testing.T) {
	t.Run("should register a customer with valid data", func(t *testing.T) {
		testData := loadYAMLTestData(t, "./data/valid-cases.yaml")

		cases := extractCases(t, testData)
		for title, bundle := range cases {
			caseData := bundleToCaseData(t, bundle)
			request := extractRequest(t, caseData)

			t.Run(title, func(t *testing.T) {
				// Given that
				customer.ArrangeInternalsNoCustomerIsRegistered(t)

				// When we
				result := customer.ActTryToRegisterACustomer(t, request)
				// with valid data

				// Then the
				customer.AssertRegistrationShouldSucceed(t, result)

				// And the
				customer.AssertInternalsCustomerShouldBeProperlyRegistered(t, request)
			})
		}
	})

	t.Run("should reject a registration with invalid data", func(t *testing.T) {
		testData := loadYAMLTestData(t, "./data/invalidation-cases.yaml")

		cases := extractCases(t, testData)
		for title, bundle := range cases {
			caseData := bundleToCaseData(t, bundle)
			request := extractRequest(t, caseData)
			findOnError := extractFindOnError(t, caseData)

			t.Run(title, func(t *testing.T) {
				// Given that
				customer.ArrangeInternalsNoCustomerIsRegistered(t)

				// When we
				result := customer.ActTryToRegisterACustomer(t, request)
				// with invalid data

				// Then the
				customer.AssertRegistrationShouldFailWithMessage(t, result, findOnError...)

				// And the
				customer.AssertInternalsCustomerShouldNotBeRegistered(t, request)
			})
		}
	})

	t.Run("should reject a registration with duplicated data", func(t *testing.T) {
		testData := loadYAMLTestData(t, "./data/duplication-cases.yaml")

		referenceRequest := extractReferenceRequest(t, testData)

		cases := extractCases(t, testData)
		for title, bundle := range cases {
			caseData := bundleToCaseData(t, bundle)
			request := extractRequest(t, caseData)
			findOnError := extractFindOnError(t, caseData)

			t.Run(title, func(t *testing.T) {
				// Given that
				customer.ArrangeInternalsSomeCustomersAreRegistered(t, referenceRequest)

				// When we
				result := customer.ActTryToRegisterACustomer(t, request)
				// with duplicated data

				// Then the
				customer.AssertRegistrationShouldFailWithMessage(t, result, findOnError...)

				// And the
				customer.AssertInternalsCustomerShouldNotBeRegistered(t, request)
			})
		}
	})

	t.Run("should not register the same user twice", func(t *testing.T) {
		referenceCustomer := loadYAMLTestData(t, "./data/reference-customer.yaml")

		// Given that
		customer.ArrangeInternalsSomeCustomersAreRegistered(t, referenceCustomer)

		// When we
		result := customer.ActTryToRegisterACustomer(t, referenceCustomer)
		// twice

		// Then the
		customer.AssertRegistrationShouldFailWithMessage(t, result, customer.ErrDuplication.Error(), "duplicated name", "duplicated email", "duplicated phone")

		// And
		customer.AssertInternalsCustomerShouldNotBeDuplicated(t, referenceCustomer)
	})

	t.Run("should return a generic system error on failure", func(t *testing.T) {
		referenceCustomer := loadYAMLTestData(t, "./data/reference-customer.yaml")

		// Given that
		customer.ArrangeInternalsNoCustomerIsRegistered(t)

		// And
		customer.ArrangeInternalsSomethingCausingAProblem(t)

		// When we
		result := customer.ActTryToRegisterACustomer(t, referenceCustomer)
		// with valid data

		// Them the
		customer.AssertRegistrationShouldFailWithMessage(t, result, "system error", "contact support")

		// And the
		customer.AssertInternalsCustomerShouldNotBeRegistered(t, referenceCustomer)
	})
}

// loadYAMLTestData loads content of a YAML test data file into a
// `map[string]any`.
func loadYAMLTestData(t *testing.T, path string) map[string]any {
	t.Helper()

	r := require.New(t)

	f, err := os.Open(path)
	r.NoError(err)
	r.NotNil(f)
	defer f.Close()

	var td map[string]any
	err = yaml.NewDecoder(f).Decode(&td)
	r.NoError(err)
	r.NotNil(r)

	return td
}

// extractReferenceRequest extracts the reference request from the given test
// data.
//
// It looks for the following attributes:
// - reference_request: map[string]any
func extractReferenceRequest(t *testing.T, testData map[string]any) map[string]any {
	t.Helper()

	r := require.New(t)

	r.Contains(testData, "reference_request")
	r.NotNil(testData["reference_request"])
	r.IsType(map[string]any{}, testData["reference_request"])

	return testData["reference_request"].(map[string]any)
}

// extractCases extracts the test cases from the given test data.
//
// It looks for the following attributes:
// - cases: map[string]any
func extractCases(t *testing.T, testData map[string]any) map[string]any {
	r := require.New(t)

	r.Contains(testData, "cases")
	r.NotNil(testData["cases"])
	r.IsType(map[string]any{}, testData["cases"])

	return testData["cases"].(map[string]any)
}

// bundleToCaseData converts a bundle (any) to a map[string]any.
func bundleToCaseData(t *testing.T, bundle any) map[string]any {
	r := require.New(t)

	r.NotNil(bundle)
	r.IsType(map[string]any{}, bundle)
	caseData := bundle.(map[string]any)

	return caseData
}

// extractRequest extracts the request from the given case data.
//
// It looks for the following attributes:
// - request: map[string]any
func extractRequest(t *testing.T, caseData map[string]any) map[string]any {
	r := require.New(t)

	r.Contains(caseData, "request")
	r.NotNil(caseData["request"])
	r.IsType(map[string]any{}, caseData["request"])

	return caseData["request"].(map[string]any)
}

// extractFindOnError extracts the `find_on_error` attribute from the given case
// data.
//
// It looks for the following attributes:
// - find_on_error: []string
func extractFindOnError(t *testing.T, caseData map[string]any) []string {
	r := require.New(t)

	r.Contains(caseData, "find_on_error")
	r.NotNil(caseData["find_on_error"])
	r.IsType([]any{}, caseData["find_on_error"])

	vals := caseData["find_on_error"].([]any)

	findOnErrors := make([]string, len(vals))
	for i, val := range vals {
		r.IsType("", val)
		findOnErrors[i] = val.(string)
	}

	return findOnErrors
}
