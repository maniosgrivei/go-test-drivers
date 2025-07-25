package customer_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/maniosgrivei/go-test-drivers/customer"
	"github.com/maniosgrivei/go-test-drivers/customer/adapters/presentation/rest"
	badgerpoc "github.com/maniosgrivei/go-test-drivers/customer/adapters/repository/badger-poc"
	reference "github.com/maniosgrivei/go-test-drivers/customer/adapters/repository/reference"
	sqlitepoc "github.com/maniosgrivei/go-test-drivers/customer/adapters/repository/sqlite-poc"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

//
// Tests
//
//

// shouldRegisterACustomerWithValidData tests the successful registration of a
// customer.
func shouldRegisterACustomerWithValidData(
	t *testing.T,
	testDriver *customer.CustomerServiceTestDriver,
	request map[string]any,
	extraArgs map[string]any,
) {
	t.Helper()

	// Given that
	testDriver.ArrangeInternalsNoCustomerIsRegistered(t)

	// When we
	result := testDriver.ActTryToRegisterACustomer(t, request, extraArgs)
	// with valid data

	// Then the
	testDriver.AssertRegistrationShouldSucceed(t, result, extraArgs)

	// And the
	testDriver.AssertInternalsCustomerShouldBeProperlyRegistered(t, request)
}

// shouldRejectARegistrationWithInvalidData tests the rejection of a customer
// registration due to invalid data.
func shouldRejectARegistrationWithInvalidData(
	t *testing.T,
	testDriver *customer.CustomerServiceTestDriver,
	request map[string]any,
	extraArgs map[string]any,
	findOnError []string,
) {
	t.Helper()

	// Given that
	testDriver.ArrangeInternalsNoCustomerIsRegistered(t)

	// When we
	result := testDriver.ActTryToRegisterACustomer(t, request, extraArgs)
	// with invalid data

	// Then the
	testDriver.AssertRegistrationShouldFailWithMessage(t, result, extraArgs, findOnError...)

	// And the
	testDriver.AssertInternalsCustomerShouldNotBeRegistered(t, request)
}

// shouldRejectARegistrationWithDuplicatedData tests the rejection of a customer
// registration due to duplicated data.
func shouldRejectARegistrationWithDuplicatedData(
	t *testing.T,
	testDriver *customer.CustomerServiceTestDriver,
	referenceRequest, request map[string]any,
	extraArgs map[string]any,
	findOnError []string,
) {
	t.Helper()

	// Given that
	testDriver.ArrangeInternalsSomeCustomersAreRegistered(t, referenceRequest)

	// When we
	result := testDriver.ActTryToRegisterACustomer(t, request, extraArgs)
	// with duplicated data

	// Then the
	testDriver.AssertRegistrationShouldFailWithMessage(t, result, extraArgs, findOnError...)

	// And the
	testDriver.AssertInternalsCustomerShouldNotBeRegistered(t, request)
}

// shouldNotRegisterTheSameUserTwice tests that the same user cannot be
// registered twice.
func shouldNotRegisterTheSameUserTwice(
	t *testing.T,
	testDriver *customer.CustomerServiceTestDriver,
	referenceCustomer map[string]any,
	extraArgs map[string]any,
) {
	t.Helper()

	// Given that
	testDriver.ArrangeInternalsSomeCustomersAreRegistered(t, referenceCustomer)

	// When we
	result := testDriver.ActTryToRegisterACustomer(t, referenceCustomer, extraArgs)
	// twice

	// Then the
	testDriver.AssertRegistrationShouldFailWithMessage(
		t, result, extraArgs,
		customer.ErrDuplication.Error(), "duplicated name", "duplicated email", "duplicated phone",
	)

	// And
	testDriver.AssertInternalsCustomerShouldNotBeDuplicated(t, referenceCustomer)
}

// shouldReturnAGenericSystemErrorOnFailure tests that a generic system error is
// returned on failure.
func shouldReturnAGenericSystemErrorOnFailure(
	t *testing.T,
	customerTestDriver *customer.CustomerServiceTestDriver,
	referenceCustomer map[string]any,
	extraArgs map[string]any,
) {
	t.Helper()

	// Given that
	customerTestDriver.ArrangeInternalsNoCustomerIsRegistered(t)

	// And
	customerTestDriver.ArrangeInternalsSomethingCausingAProblem(t)

	// When we
	result := customerTestDriver.ActTryToRegisterACustomer(t, referenceCustomer, extraArgs)
	// with valid data

	// Them the
	customerTestDriver.AssertRegistrationShouldFailWithMessage(t, result, extraArgs, "system error", "contact support")
}

//
// Test Suite
//
//

// TestRegisterCustomer is the acceptance test suite for the customer registration
// use case.
func TestRegisterCustomer(t *testing.T) {
	for _, variant := range []string{
		referenceSUTVariant,
		sqliteSUTVariant,
		badgerSUTVariant,
		referenceRESTSUTVariant,
		sqliteRESTSUTVariant,
		badgerRESTSUTVariant,
	} {
		t.Run(fmt.Sprintf("with system variant %s", variant), func(t *testing.T) {
			customerTestDriver := sutSetup(t, variant)

			t.Run("should register a customer with valid data", func(t *testing.T) {
				testData := loadYAMLTestData(t, "./data/valid-cases.yaml")

				cases := extractCases(t, testData)
				for title, bundle := range cases {
					caseData := bundleToCaseData(t, bundle)
					request := extractRequest(t, caseData)
					extraArgs := extractExtraArgs(t, caseData)

					t.Run(title, func(t *testing.T) {
						shouldRegisterACustomerWithValidData(t, customerTestDriver, request, extraArgs)
					})
				}
			})

			t.Run("should reject a registration with invalid data", func(t *testing.T) {
				testData := loadYAMLTestData(t, "./data/invalidation-cases.yaml")

				cases := extractCases(t, testData)
				for title, bundle := range cases {
					caseData := bundleToCaseData(t, bundle)
					request := extractRequest(t, caseData)
					extraArgs := extractExtraArgs(t, caseData)
					findOnError := extractFindOnError(t, caseData)

					t.Run(title, func(t *testing.T) {
						shouldRejectARegistrationWithInvalidData(t, customerTestDriver, request, extraArgs, findOnError)
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
					extraArgs := extractExtraArgs(t, caseData)
					findOnError := extractFindOnError(t, caseData)

					t.Run(title, func(t *testing.T) {
						shouldRejectARegistrationWithDuplicatedData(
							t, customerTestDriver, referenceRequest, request, extraArgs, findOnError,
						)
					})
				}
			})

			t.Run("should not register the same user twice", func(t *testing.T) {
				referenceCustomer := loadYAMLTestData(t, "./data/reference-customer.yaml")
				extraArgs := loadYAMLTestData(t, "./data/conflict-extra-args.yaml")

				shouldNotRegisterTheSameUserTwice(t, customerTestDriver, referenceCustomer, extraArgs)
			})

			t.Run("should return a generic system error on failure", func(t *testing.T) {
				referenceCustomer := loadYAMLTestData(t, "./data/reference-customer.yaml")
				extraArgs := loadYAMLTestData(t, "./data/server-error-extra-args.yaml")

				shouldReturnAGenericSystemErrorOnFailure(t, customerTestDriver, referenceCustomer, extraArgs)
			})
		})
	}
}

//
// SUT Setup

const (
	referenceSUTVariant     = "reference"
	referenceRESTSUTVariant = "reference-rest"

	sqliteSUTVariant     = "sqlite"
	sqliteRESTSUTVariant = "sqlite-rest"

	badgerSUTVariant     = "badger"
	badgerRESTSUTVariant = "badger-rest"
)

// sutSetup creates a new CustomerService and CustomerServiceTestDriver for the
// given SUT variant.
func sutSetup(t *testing.T, variant string) *customer.CustomerServiceTestDriver {
	t.Helper()

	var (
		customerRepository        customer.CustomerRepository
		repositoryTestDriver      customer.CustomerRepositoryTestDriver
		customerService           *customer.CustomerService
		customerServiceTestDriver *customer.CustomerServiceTestDriver
	)

	// Setup repository
	switch variant {
	case referenceSUTVariant, referenceRESTSUTVariant:
		repo := reference.NewReferenceCustomerRepository()
		customerRepository = repo
		repositoryTestDriver = reference.NewReferenceCustomerRepositoryTestDriver(repo)

	case sqliteSUTVariant, sqliteRESTSUTVariant:
		repo := sqlitepoc.NewSQLiteCustomerRepository()
		customerRepository = repo
		repositoryTestDriver = sqlitepoc.NewSQLiteCustomerRepositoryTestDriver(repo)

	case badgerSUTVariant, badgerRESTSUTVariant:
		repo := badgerpoc.NewBadgerCustomerRepository()
		customerRepository = repo
		repositoryTestDriver = badgerpoc.NewBadgerCustomerRepositoryTestDriver(repo)

	default:
		t.Fatalf("unknown SUT variant: %s", variant)
	}

	customerService = customer.NewCustomerService(customerRepository)

	// Setup presentation
	switch variant {
	case referenceRESTSUTVariant, sqliteRESTSUTVariant, badgerRESTSUTVariant:
		restAPIHandler := rest.NewCustomerRESTAPIHandler(customerService)

		restPresentationTestDriver := rest.NewCustomerRESTAPIHandlerTestDriver(restAPIHandler)

		customerServiceTestDriver = customer.NewCustomerServiceTestDriverWithPresentation(
			customerService, repositoryTestDriver, restPresentationTestDriver,
		)

	default:
		customerServiceTestDriver = customer.NewCustomerServiceTestDriver(customerService, repositoryTestDriver)
	}

	return customerServiceTestDriver
}

//
// Test Data Helpers

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

// extractExtraArgs extracts the extra args map from the given case data.
//
// It looks for the following attributes:
// - extra_args: map[string]any
func extractExtraArgs(t *testing.T, caseData map[string]any) map[string]any {
	r := require.New(t)

	r.Contains(caseData, "extra_args")
	r.NotNil(caseData["extra_args"])
	r.IsType(map[string]any{}, caseData["extra_args"])

	return caseData["extra_args"].(map[string]any)
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
