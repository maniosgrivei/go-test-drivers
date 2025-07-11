# Go Test Drivers

### Building the Foundation for DSL-Driven Tests in Go

<div style="text-align: center;" width= "100%">
<img src="./doc/images/go-test-drivers-banner.png" alt="Go Test Drivers" width="100%" height="auto" allign="center" \>
</div>

**What if your tests could speak the language of your business?** Instead of
imperative scripts, imagine a suite of tests that reads like a clear
specification of your system's behavior.

Our guiding principle is that **acceptance tests should validate the**
**integrated system in an environment as close to production as possible**.
Therefore, in this article, **we will deliberately avoid replacing our core**
**components with test doubles or mocks**.

To achieve this, we'll turn to a powerful concept from software engineering: the
"_**Seam**_", first introduced by
**<a href="https://www.google.com/search?q=michael+feathers" target="_blank">Michael Feathers</a>**:
**A *Seam* is a place in the code where you can alter the program's behavior**
**without editing the code in that place**.

This is the principle behind _**Go Test Drivers**_.

This article, inspired by the video
**<a href="https://www.youtube.com/watch?v=JDD5EEJgpHU" target="_blank">How to Write Acceptance Tests</a>**
from the
**<a href="https://www.youtube.com/@ModernSoftwareEngineeringYT" target="_blank">Modern Software Engineering</a>**
YouTube channel by
**<a href="https://www.google.com/search?q=dave+farley" target="_blank">Dave Farley</a>**,
will show you how to implement "_**Test Protocol Drivers**_" in **Go**. We will
build them as what they are: **specialized mechanisms that translate terms**
**from a _Domain-Specific Language (DSL)_ into direct interactions with a**
_**System Under Test (SUT)**_. 

By the end, you'll have a practical framework for building this foundation,
**unlocking the ability to take the next step**: **writing tests that are**
**deeply decoupled, easy to maintain, and serve as living documentation for**
**your application**.

-----

## Table of Contents

- [The Black-Box Dilemma: Control and Observation](#the-black-box-dilemma-control-and-observation)
- [Case Study](#case-study)
  - [The First Attempt: A Vicious Cycle](#the-first-attempt-a-vicious-cycle)
  - [The Insight: A Pattern in the Chaos](#the-insight-a-pattern-in-the-chaos)
  - [Hitting a Wall: Crossing the Package Boundary](#hitting-a-wall-crossing-the-package-boundary)
  - [A Hardware Analogy: The Test Pad Pattern](#a-hardware-analogy-the-test-pad-pattern)
  - [Toward a Generic Driver: Decoupling from Data Structures](#toward-a-generic-driver-decoupling-from-data-structures)
  - [The Payoff: Evolving the System with Minimal Test Changes](#the-payoff-evolving-the-system-with-minimal-test-changes)
  - [The Core Refactoring: Introducing the Repository Pattern](#the-core-refactoring-introducing-the-repository-pattern)
  - [AI-Assisted Development: Generating the REST Layer](#ai-assisted-development-generating-the-rest-layer)
- [The Go Test Driver Pattern (Formal Definition)](#the-go-test-driver-pattern)
  - [Application](#application)
  - [Architectural Variants](#architectural-variants)
  - [Components](#components)
- [Putting the Pattern into Practice](#putting-the-pattern-into-practice)
  - [Step 1: Define the Skeleton](#step-1-define-the-skeleton)
  - [Step 2: Build a Flexible Test Harness](#step-2-build-a-flexible-test-harness)
  - [Step 3: Write Gherkin Scenarios](#step-3-write-gherkin-scenarios)
  - [Step 4: From Scenarios to Test Signatures](#step-4-from-scenarios-to-test-signatures)
  - [Step 5: The Red-Green-Refactor Cycle](#step-5-the-red-green-refactor-cycle)
  - [Step 6: Evolving the System and the Tests](#step-6-evolving-the-system-and-the-tests)
- [Practical Usage and Tooling](#practical-usage-and-tooling)
- [Conclusion](#conclusion)
- [Next Steps: Toward a True DSL](#next-steps-toward-a-true-dsl)

-----

## The Black-Box Dilemma: Control and Observation

Before we build our first driver, we need to understand two subtle but critical
challenges that high-level acceptance tests face, especially when we choose to
avoid test doubles.

1.  **The Control Problem: Handling Internals**

    Often, acceptance tests are placed in a separate package to force them to
    interact with the *System Under Test (SUT)* just like a real client would.
    This is great for ensuring correctness, but it creates a problem: the test
    code is blocked from accessing any internal, unexported parts of the system.

    This leads to a crucial question: If our test can't reach inside the system,
    how can we force it into a specific state—like simulating a database
    failure—in a clean and deterministic way?

2.  **The Observation Problem: Asserting Side Effects**

    This is a direct consequence of the control problem. Real-world systems
    produce side effects—a record is created, a file is written, or a message is
    sent to a queue. If our test only sees the final output of a function, how
    can it verify that the correct side effect actually occurred?

    This brings us to the second question: If our test is outside the system,
    how can we directly observe and assert that an internal state change
    happened exactly as we intended?

-----

## Case Study

Imagine you work for a world-class software company, tasked with building a new
flagship product from scratch: an Ultimate CRM System.

A product of this scale must be designed from the perspective of multiple
stakeholders. The definition of a "user" goes far beyond the person at a desk
operating the web interface. For this project, our stakeholders include:

-   The **End-User**, like a _Sales Person_, who needs an efficient and intuitive
    interface for their daily work.

-   The **Client**, represented by their _Engineers_ and _System Administrators_,
    who are responsible for deploying, maintaining, and integrating the system
    within their company's infrastructure.

-   The **Compliance Officer**, who must ensure the system respects industry
    standards and guarantees data quality.

-   The **System Integrator**, who will connect the CRM to other enterprise
    systems like an ERP or a corporate website.

Each of these personas has valid needs that represent real business value.

At this early stage, nothing is defined—the project is a blank slate. With this
context in mind, you receive the first User Story and its Acceptance Criteria.

-----

### User Story A: Registering a New Customer

> As a _**Sales Person**_, I want to register a new customer on the CRM system by
> providing their pertinent data: _Name_, _E-mail_, and _Phone Number_. As a
> result, I want to receive the unique _ID_ for that customer.

-----

#### Acceptance Criteria

  - **Success Case**:
      - Upon successful registration, a unique ID for the new customer is
        returned.
  - **Validation Errors**:
      - If any required field is not provided, the system must reject the
        registration with a specific error.
  - **Uniqueness Errors**:
      - If the Name, E-mail, or Phone Number already exists, the system must
        reject the registration with a specific error.
  - **Generic Errors**:
      - If any other unexpected error occurs, the system must reject the
        registration with a generic error message.

-----

### The First Attempt: A Vicious Cycle

So, after tackling that first User Story with a pure _**Red-Green-Refactor**_
approach, our first "happy path" test looks simple and clean:

```go
func TestRegisterCustomer(t *testing.T) {
	t.Run("should register a customer with valid data", func(t *testing.T) {
		r := require.New(t)

		request := &RegisterRequest{Name: dfltName, Email: dfltEmail, Phone: dfltPhone}

		Init()

		id, err := Register(request)

		r.NoError(err)
		r.NotEmpty(id)
	})

	// ...
}
```

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-01" target="_blank">Step 1</a>**_.

But this passing test hides a critical flaw: we are blind to the true side
effect. We only assert the return value, but nothing ensures the customer was
actually registered in the repository.

To solve this, we must directly inspect the internal state. But this "fix"
comes at a steep price. Look at what our once-simple test becomes:

```go
func TestRegisterCustomer(t *testing.T) {
    t.Run("should register a customer with valid data", func(t *testing.T) {
		r := require.New(t)

		// Arrange: Set the desired initial state.
		customers = make([]*Customer, 0)
		idIndex = make(map[string]*Customer)
		nameIndex = make(map[string]*Customer)
		emailIndex = make(map[string]*Customer)
		phoneIndex = make(map[string]*Customer)

		// Act: Perform an action on the system.
		request := &RegisterRequest{Name: dfltName, Email: dfltEmail, Phone: dfltPhone}

		id, err := Register(request)

		// Assert: Check the results.
		r.NoError(err)
		r.NotEmpty(id)

		// Assert: Check the side-effects.
		customer := Customer{ID: id, Name: dfltName, Email: dfltEmail, Phone: dfltPhone}

		r.Contains(customers, &customer)

		r.Contains(idIndex, id)
		r.EqualValues(idIndex[id], &customer)

		r.Contains(nameIndex, dfltName)
		r.EqualValues(nameIndex[dfltName], &customer)

		r.Contains(emailIndex, dfltEmail)
		r.EqualValues(emailIndex[dfltEmail], &customer)

		r.Contains(phoneIndex, dfltPhone)
		r.EqualValues(phoneIndex[dfltPhone], &customer)
	})

	// ...
}
```
_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-02" target="_blank">Step 2</a>**_.

This is a vicious cycle. To gain observability, our clean test has ballooned
into a messy, fragile script, tightly coupled to implementation details. This is
not readable, not scalable, and a nightmare to maintain. We have gained control,
but we have lost simplicity.

-----

### The Insight: A Pattern in the Chaos

However, even inside this complexity, a powerful pattern is emerging. Notice the
structure of our test: we _**Arrange**_ state, _**Act**_ on the system, and 
_**Assert**_ the outcome. If this feels familiar, it's because it directly
mirrors the "_**Given-When-Then**_" formula of _**Gherkin**_.

Our true goal is revealed: we want our Go tests to read like clean, high-level
specifications. This insight leads to a major refactoring on test: encapsulating
the messy details into helper functions with a clear, _**AAA+**_ naming
convention:

- `Arrange...`: A function used to set up a desired initial state by using the
  system's public API.

- `ArrangeInternals...`: A function that sets up state by "cheating" and
  manipulating the system's internal state directly. This prefix signals that we
  are reaching into unexported parts of our package.

- `Act...`: A function that performs the single action on the system that we
  want to test.

- `Assert...`: A function that verifies the results of the action based on the
  function's return values (e.g., outputs and errors).

- `AssertInternals...`: A function that verifies the side effects of the
  action by inspecting the system's internal state directly.

Our tests now look like this:

```go
func TestRegisterCustomer(t *testing.T) {
    t.Run("should register a customer with valid data", func(t *testing.T) {
		// Given that
		ArrangeInternalsNoCustomerIsRegistered(t)

		// When we
		id, err := ActTryToRegisterACustomer(t, &RegisterRequest{Name: dfltName, Email: dfltEmail, Phone: dfltPhone})
		// with valid data

		// Then the
		AssertRegistrationShouldSucceed(t, id, err)

		// And the
		AssertInternalsCustomerShouldBeProperlyRegistered(t, &Customer{ID: id, Name: dfltName, Email: dfltEmail, Phone: dfltPhone})
	})

	// ...
}
```

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-03" target="_blank">Step 3</a>**_.

It is very close to a _Gherking_ specification:

```gherkin
Given that no customer is registered
When we try to register a customer with valid data
Then the registration should succeed
And the customer should be properly registered
```

We have achieved clarity, but our helpers are still reliant on directly
manipulating the internal state from inside our package. 

-----

### Leave Home: Crossing the Package Boundary

Now it's time for our tests to "_leave home_". We'll move them from the cozy
confines of the `customer` package into their own `customer_test` package in
another directory. This forces them to behave like a privileged client,
interacting with our system not through its public API, but through our
dedicated, test-only helper functions.

This move creates a pivotal crisis: our tests **no longer compile**: They cannot
see the test-only helper functions in the main package.

To solve that, we need to **Isolate the Helpers** by moving them into a new
file, `customer_test_driver.go`, within the `customer` package. At the very top
of this new file, we add the build constraint `//go:build test`. This tag tells
the Go compiler to only include this file when the test tag is active.

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-04" target="_blank">Step 4</a>**_.

_**NOTE**: To ensure your code editor's language server (`gopls`) can see these_
_helpers, it's strongly recommended to add the `test` tag to your editor's Go_
_extension settings_.

This awkwardness isn't a flaw in Go; it's a powerful signal. It tells us that
our helper functions have effectively become a special, privileged interface to
our system's internals.

-----

### A Hardware Analogy: The Test Pad Pattern

Having worked for many years in electronic circuit development, I find the best
way to understand this is through an analogy. Think of our `customer` package as
a finished circuit board sealed in an enclosure. The end-user only interacts
with its public ports.

However, engineers intentionally place _**test pads**_ on that board—special
points for technicians to stimulate and measure the circuit's internal state.
Our `...Internals...` helper functions are these test pads.

The `-tags test` flag is the act of **removing the enclosure**. A regular
`go build` keeps the enclosure on. By adding the flag, we, as the system's
technicians, gain access to these restricted points. In software, this concept
is known as a **"Seam,"** a term coined by **Michael Feathers**.

Viewing our pattern as an intentional "Seam" reframes our approach. We are not
hacking; we are building a professional interface for deep testing.

To be crystal clear: **this entire testing framework has zero impact on your
production binary**. Any file marked with `//go:build test` is ignored by a
standard `go build`.

-----

### Toward a Generic Driver: Decoupling from Data Structures

The next challenge is that our helper functions are tightly coupled to concrete Go
structs (e.g., `RegisterRequest`). What happens when we add a REST API layer to
our service? It will likely require a different data structure. This would force
us to change our helper function signatures, making it impossible to create a
single, stable testing interface for all layers of the application.

The solution is to decouple our helpers from specific structs by using generic
input formats, like `map[string]any`, and moving our test data into external
YAML files.

After refactoring, even our most complex test case becomes a clean, high-level
specification:

```go
func TestRegisterCustomer(t *testing.T) {
	// ...

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

	// ...
}
```

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-05" target="_blank">Step 5</a>**_.

The power of this approach is that our helper functions now form a
**stable testing API**. Their signatures are no longer tied to any one
implementation's data structures. We can add new layers to our service, each
with different data shapes, while the high-level tests that call these helpers
remain unchanged.

However, this approach has a clear trade-off: we sacrifice Go's compile-time
safety for data-driven flexibility. The complexity of validating the generic
data is pushed into our Test Driver, which must now be more defensive.

-----

### The Payoff: Evolving the System with Minimal Test Changes

Now comes the real test of our design. A good architecture should be resilient
to change and easy to evolve. Let's see how our testing framework handles two
new, complex feature requests from the business.

We have received two new User Stories:

-----

### User Story B: A New Format for Customer ID

> As a ***Customer***, I want to receive a more meaningful and easy-to-record
> *Customer ID*.

-----

#### Acceptance Criteria

The Customer ID must have the following format: `CCCC-YYMD-DXXX`

  - Twelve characters total, composed of three groups of four characters
    separated by hyphens.
  - Only uppercase letters (A-Z) and numbers (0-9) are allowed.
  - The **first group (`CCCC`)** is the name code.
  - The **second group (`YYMD`)** contains most of the registration date.
  - The **third group (`DXXX`)** contains the rest of the date and the unique
    code.
  - **Example:** `JHND-25F1-5A0R`.

-----

### User Story C: Enhanced Customer Data Validation

> As a ***Compliance Officer***, I want the system to perform stricter
> validation on customer data to ensure data quality.

-----

#### Acceptance Criteria

  - The **customer name** must be validated for length, allowed characters, and
    structure.
  - The **customer e-mail** must be validated for length, characters, and a
    valid format.
  - The **customer phone** must be validated for length and a valid international
    format.

These stories introduce complex rules. Let's analyze the impact of each one and
see how our testing framework holds up under pressure.

-----

### Implementing the New Customer ID Format

Implementing the new ID algorithm requires a fairly complex sequence of
operations involving strings, dates, and times.

However, from the perspective of our high-level acceptance test, very little has
changed. The `Register` function still accepts a request and returns a `string`
ID. The internal complexity of *how* that string is generated is an
implementation detail.

Therefore, we made a crucial decision: **the complex ID generation logic is**
**tested exhaustively at the *Unit Test* level**. Our acceptance tests don't
need to re-verify it.

As a result, the only impact on our entire test driver was a tiny change to the
`Arrange...` helper, which now calls the new `GenerateID` function instead of
creating a simple numeric ID:

```diff
func ArrangeInternalsSomeCustomersAreRegistered(t *testing.T, customerMaps ...map[string]any) {
	t.Helper()

	ArrangeInternalsNoCustomerIsRegistered(t)

-	for id, cm := range customerMaps {
+	for _, cm := range customerMaps {
		c := getCustomerFromMap(t, cm)

		if c.ID == "" {
-			c.ID = fmt.Sprintf("%d", id)
+			var err error
+			c.ID, err = GenerateID(c.Name, time.Now())
+			require.NoError(t, err)
			require.NotEmpty(t, c.ID)
		}

		customers = append(customers, c)
		idIndex[c.ID] = c
		nameIndex[c.Name] = c
		emailIndex[c.Email] = c
		phoneIndex[c.Phone] = c
	}
}
```

This is a huge win. A significant internal change was implemented and verified
with almost zero impact on our high-level acceptance tests.

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-06" target="_blank">Step 6</a>**_.

-----

### Implementing Enhanced Customer Data Validation

This User Story also resulted in a complex set of new validation functions.
Unlike the ID change, these changes directly affect the high-level behavior of
our system, as we now have many new, specific error messages that should be
returned to the user.

This might seem like it would require a massive testing effort. However, because
we designed a data-driven framework, the impact was again minimal.

The only structural change to our test suite was to wrap the original "happy
path" test into a loop to run over a new YAML file of valid test cases. The
test logic itself remains identical:


```diff
	t.Run("should register a customer with valid data", func(t *testing.T) {
-		referenceCustomer := loadYAMLTestData(t, "./data/reference-customer.yaml")
+		testData := loadYAMLTestData(t, "./data/valid-cases.yaml")

-		// Given that
-		customer.ArrangeInternalsNoCustomerIsRegistered(t)
+		cases := extractCases(t, testData)
+		for title, bundle := range cases {
+			caseData := bundleToCaseData(t, bundle)
+			request := extractRequest(t, caseData)

-		// When we
-		result := customer.ActTryToRegisterACustomer(t, referenceCustomer)
-		// with valid data
+			t.Run(title, func(t *testing.T) {
+				// Given that
+				customer.ArrangeInternalsNoCustomerIsRegistered(t)

-		// Then the
-		customer.AssertRegistrationShouldSucceed(t, result)
+				// When we
+				result := customer.ActTryToRegisterACustomer(t, request)
+				// with valid data

-		// And the
-		customer.AssertInternalsCustomerShouldBeProperlyRegistered(t, referenceCustomer)
+				// Then the
+				customer.AssertRegistrationShouldSucceed(t, result)
+
+				// And the
+				customer.AssertInternalsCustomerShouldBeProperlyRegistered(t, request)
+			})
+		}
	})
```

All the new, complex validation rules were tested simply by adding new cases to
our both `valid-cases.yaml` and `invalidation-cases.yaml` files. Our acceptance
tests validated this new feature with zero changes to the Go test code, only to
the external data sets. This is the power of a truly decoupled testing
framework.

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-07" target="_blank">Step 7</a>**_.

-----

### Hitting the Wall And Jumpping It

Our pragmatic approach, following the eXtreme Programming philosophy, has led us
to build more than a simple prototype; we have a strong acceptance test
scaffolding to guide the evolution of our system.

But now we face a true architectural challenge—a step toward a real-world
system—led by this User Story:

-----

### User Story D: Data Storage Flexibility

> As a **System Administrator** preparing to deploy the CRM, I need assurance
> that the system can integrate with my company's certified storage
> infrastructure (e.g., SQL or NoSQL based), so that it will comply with our
> mandatory data policies.

-----

#### Acceptance Criteria

For this initial architectural validation, the following outcomes are sufficient:

- The system's architecture must fully decouple the core service logic from the
  specific data storage mechanism.

- Two Prooves of Concept (PoC) must be implemented:
  - A PoC using a SQL-based technology.
  - A PoC using a NoSQL-based technology.
  - Both PoC are considered successful when they allow the **entire existing**
    **suite of acceptance tests to pass** without any changes to the test code
	itself.

-----

### Implementing Data Storage PoCs

Before we can implement the different storage backends, we must first refactor
our code to be more robust and organized.

-----

#### Moving Everithing to `struct`s

Our first, and most important, action is to finally eliminate the package-level
global variables. We'll create a `CustomerService` struct to encapsulate the
`customers` slice and all the index maps. Consequently, our package-level
functions (like `Register`) will become methods bound to this new struct.

At the same time, we will formalize our test helpers. Instead of a loose
collection of functions, we'll create our first true **`TestDriver`** type. This
driver struct will embed the `CustomerService`, giving it direct, white-box
access to the service's internal state for testing. The result looks like this:

```go
// ...

type CustomerServiceTestDriver struct {
	*CustomerService
}

func NewCustomerServiceTestDriver(customerService *CustomerService) *CustomerServiceTestDriver {
	return &CustomerServiceTestDriver{
		CustomerService: customerService,
	}
}

func (td *CustomerServiceTestDriver) ArrangeInternalsNoCustomerIsRegistered(t *testing.T) {
	t.Helper()

	td.customers = make([]*Customer, 0)
	// ...
}

// ...
```

By embedding a pointer to `*CustomerService`, our `CustomerServiceTestDriver`
can directly manipulate the encapsulated state (the slices and maps) for test
setup and verification. All our test helper functions now become methods on this
driver type.

This refactoring naturally required a small change in our test setup to
instantiate these new types. However, the core logic and high-level structure
of the tests remained completely unchanged:

```diff
// ...

func TestRegisterCustomer(t *testing.T) {
+	customerTestDriver := customer.NewCustomerServiceTestDriver(customer.NewCustomerService())
+
	t.Run("should register a customer with valid data", func(t *testing.T) {
		testData := loadYAMLTestData(t, "./data/valid-cases.yaml")

		cases := extractCases(t, testData)
		for title, bundle := range cases {
			caseData := bundleToCaseData(t, bundle)
			request := extractRequest(t, caseData)

			t.Run(title, func(t *testing.T) {
				// Given that
-				customer.ArrangeInternalsNoCustomerIsRegistered(t)
+				customerTestDriver.ArrangeInternalsNoCustomerIsRegistered(t)

				// When we
-				result := customer.ActTryToRegisterACustomer(t, request)
+				result := customerTestDriver.ActTryToRegisterACustomer(t, request)
				// with valid data

				// Then the
-				customer.AssertRegistrationShouldSucceed(t, result)
+				customerTestDriver.AssertRegistrationShouldSucceed(t, result)

				// And the
-				customer.AssertInternalsCustomerShouldBeProperlyRegistered(t, request)
+				customerTestDriver.AssertInternalsCustomerShouldBeProperlyRegistered(t, request)
			})
		}
	})

	// ...
}
```

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-08" target="_blank">Step 8</a>**_.

-----

#### Decoupling Service and Repository

In this step, we make the pivotal architectural change that solves the problems
of inflexibility and tight coupling. We will decouple the service from its data
storage by introducing a `CustomerRepository` interface.

All the code that was previously handling our in-memory storage has been moved
to a new reference package, where it's used to create a "`reference`
implementation" of this new interface. This allows our CustomerService to
depend on the `Repository` contract, not a specific implementation.

This decoupling of our production code allows for a similar, powerful separation
in our test code.

-----

##### A Two-Level Driver: Separating Concerns

Our single Test Driver is now split into two, each with a clear responsibility:

1. `CustomerServiceTestDriver`: This remains the high-level driver that our
  tests interact with. Its job is to handle generic data (`like map[string]any`)
  and translate test intentions into calls on the service or the repository
  driver.

2. `CustomerRepositoryTestDriver`: This is a new, lower-level driver. It
  is an implementation of the `Repository` interface, specifically for a
  certain storage backend. Its methods work with concrete Go types (like
  `*Customer`) and know how to manipulate that specific backend.

Here is the interface for our new repository-level driver:

```go
type CustomerRepositoryTestDriver interface {
	//
	// Arrange
	ArrangeInternalsNoCustomerIsRegistered(*testing.T)
	ArrangeInternalsSomeCustomersAreRegistered(*testing.T, []*Customer)
	ArrangeInternalsSomethingCausingAProblem(*testing.T)

	//
	// Assert
	AssertInternalsCustomerShouldBeProperlyRegistered(*testing.T, *Customer)
	AssertInternalsCustomerShouldNotBeRegistered(*testing.T, *Customer)
	AssertInternalsCustomerShouldNotBeDuplicated(*testing.T, *Customer)
}
```

The `CustomerServiceTestDriver` now holds a reference to a
`CustomerRepositoryTestDriver` and delegates the internal assertion work,
acting as a translator from generic maps to concrete types. Here is an example
of this delegation:

_File: `customer_test_driver.go`_
```go
func (td *CustomerServiceTestDriver) AssertInternalsCustomerShouldNotBeDuplicated(t *testing.T, customerData map[string]any) {
	t.Helper()

	customer := getCustomerFromMap(t, customerData)

	// Calling the method on the repository test driver dependency:
	td.repositoryTD.AssertInternalsCustomerShouldNotBeDuplicated(t, customer)
}
```

_File: `customer_repository_test_driver.go`_
```go
func (td *ReferenceCustomerRepositoryTestDriver) AssertInternalsCustomerShouldNotBeDuplicated(t *testing.T, c *customer.Customer) {
	t.Helper()

	r := require.New(t)

	// Performing the appropriate actions.
	r.LessOrEqual(sliceCountCustomeOccurrences(td.customers, c), 1)
	r.LessOrEqual(mapCountCustomeOccurrences(td.idIndex, c), 1)
	r.LessOrEqual(mapCountCustomeOccurrences(td.nameIndex, c), 1)
	r.LessOrEqual(mapCountCustomeOccurrences(td.emailIndex, c), 1)
	r.LessOrEqual(mapCountCustomeOccurrences(td.phoneIndex, c), 1)
}
```

-----

##### Composing the System with Dependency Injection

With this new architecture, the only change needed in our acceptance tests is in
the initial setup, where we now compose our service and drivers using
Dependency Injection:

```diff
// ...

func TestRegisterCustomer(t *testing.T) {
-	customerTestDriver := customer.NewCustomerServiceTestDriver(customer.NewCustomerService())
+	customerRepository := reference.NewReferenceCustomerRepository()
+	customerService := customer.NewCustomerService(customerRepository)
+
+	customerRepositoryTestDriver := reference.NewReferenceCustomerRepositoryTestDriver(customerRepository)
+	customerTestDriver := customer.NewCustomerServiceTestDriver(customerService, customerRepositoryTestDriver)

	// ...
}
```

Our acceptance test suite is now stable and fully decoupled. We are finally
ready to implement the **Data Storage PoCs**. The only change we'll need to
make is in the test suite setup above, where we can swap in different
repository implementations. The tests themselves will remain untouched.

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-09" target="_blank">Step 9</a>**_.

-----

#### An AI-Generated SQLite PoC

Now that we have clean `Repository` and `RepositoryTestDriver` interfaces, and a
reference implementation that passes all our tests, we can try something new:
generating our SQLite PoC using an AI assistant.

This is a powerful test of our architecture: if our test suite is good enough,
it should be able to guide the development of a brand-new component, even one
written by an AI.

-----

##### Preparing the Test Harness for Multiple Backends

Before we can generate the new code, we need to make some adjustments to our
acceptance tests so they can run against different backend implementations.

1.  First, we create a new package `customer/adapters/repository/sqlite`
    containing two files—`customer_repository.go` and
    `customer_repository_test_driver.go`—with the empty structs that will
    implement our `customer.CustomerRepository` and
    `customer.CustomerRepositoryTestDriver` interfaces.

2.  Next, we introduce the concept of a **SUT Variant** to our test suite. We
    create a setup function, `sutSetup(t *testing.T, variant string)`, that
    constructs and returns a fully configured test driver for a specific
    backend. This allows us to parameterize our entire test run, like so:

```diff
 // ...

 func TestRegisterCustomer(t *testing.T) {
-    customerRepository := reference.NewReferenceCustomerRepository()
-    customerService := customer.NewCustomerService(customerRepository)
-    customerRepositoryTestDriver := reference.NewReferenceCustomerRepositoryTestDriver(customerRepository)
-    customerTestDriver := customer.NewCustomerServiceTestDriver(customerService, customerRepositoryTestDriver)
+    for _, variant := range []string{"reference", "sqlite"} {
+        t.Run(fmt.Sprintf("with system variant %s", variant), func(t *testing.T) {
+            customerTestDriver := sutSetup(t, variant)
+
+            t.Run("should register a customer with valid data", func(t *testing.T) {
+                // ... all the existing tests are now run inside this loop
+            })
+        })
+    }
 }
```

-----

##### Prompting the AI for the SQLite Implementation

With our test harness ready to validate multiple backends, we can now ask the AI
to write the code. We provided the AI tool with the full context of our
`customer` package, the `reference` implementation, and the empty files in the
new `sqlite` package.

Once the context was set, we sent the following prompt:

> I need you to complete the implementations of the
> `SQLiteCustomerRepositoryTestDriver` and `SQLiteCustomerRepository` types.
> Take the `reference` implementations as an example, use `SQLite` and `sqlx`,
> and ensure the new code will pass all existing acceptance tests.

-----

##### The Result: A Working, AI-Assisted PoC

The AI generated the code in about two minutes. As is common with AI-generated
code, it was a solid first draft but not perfect—several tests failed on the
first run.

The **debugging** process, guided entirely by the failures reported by our
acceptance test suite, took roughly ten minutes. It required minor adjustments
to both the `SQLiteCustomerRepository` and its corresponding test driver.

In the end, all tests passed against the new SQLite backend. The PoC worked
exactly as intended, proving that a robust test suite is the ultimate guide for
development, whether the developer is human or not. As expected, its performance
was slower than the in-memory reference, which is perfectly acceptable given the
overhead of a real SQL engine.

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-10" target="_blank">Step 10</a>**_.

-----

#### Repeating the Recipe With a BadgerDB (NoSQL) PoC

With the success of the first AI-assisted PoC, it's time to repeat the
experiment for our NoSQL requirement. To keep things simple while proving the
concept, we chose `BadgerDB`: a fast, embedded, key-value store for Go.

The preparation was identical to the SQLite step: we created the new package
and empty types, then added the `badger` SUT variant to our test harness.

The generation process was also the same. We provided the AI with the relevant
context and the following prompt:

> I need you to complete the implementations of the
> `BadgerCustomerRepositoryTestDriver` and `BadgerCustomerRepository` types.
> Take the `reference` implementations as an example, use `BadgerDB`, and ensure
> the new code will pass all existing acceptance tests.

-----

##### A More Challenging Result

This time, however, the results were not as clean as with the `SQLite` case. The
AI seemed to struggle more with the non-relational logic of a key-value
database.

The debugging process was far more complex, but once again, **our acceptance**
**test suite was the indispensable guide**. It objectively and relentlessly
flagged every issue in the AI's logic until the implementation was correct.

Ultimately, all tests passed. This is a critical lesson: even when an
implementation is suboptimal or difficult to produce, a robust test suite
ensures it is still **correct**. The tests passed, guaranteeing our system's
behavior, while also revealing the implementation's poor performance compared to
both the in-memory and SQLite versions.

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-11" target="_blank">Step 11</a>**_.

-----

#### Tidying Up the Test Suite

At this point, the `for` loop iterating through our SUT variants has added a new
layer of indentation, making the `TestRegisterCustomer` function a bit messy and
difficult to read.

To restore clarity, we can isolate the core test logic into its own function,
like this:

```go
// ...

func shouldRegisterACustomerWithValidData(t *testing.T, testDriver *customer.CustomerServiceTestDriver, request map[string]any, extraArgs map[string]any) {
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

// ...
```

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-12" target="_blank">Step 12</a>**_.

-----

### Toward User-Facing: Adding a Presentation Layer

By today's standards, a system like the one we are proposing has at least one
transport layer that allows different types of clients (web UIs, mobile apps,
third-party systems) to integrate with it. Our `CustomerService` could also
evolve into an ***Event-Driven*** service or be part of a ***CQRS***
architecture, connecting with a *message broker* (like Kafka), and so on. We
could still have many layers until we arrive at the end-user interface.

Such a demand could be expressed by a User Story like this:

-----

### User Story E: REST API for Integration

> As a **System Integrator**, I need all services in the CRM to expose a REST
> API, so that I can communicate with them using standard HTTP protocols.

-----

#### Acceptance Criteria

A `POST /customers` endpoint should be created with the following behaviors:

  - **Success Case:**
      - A request with valid, non-duplicate customer data must return a
        `201 Created` HTTP status.
      - The response body must contain the newly generated customer ID.
  - **Client Error Cases:**
      - A request with missing or invalid data must return a `400 Bad Request`
        HTTP status.
      - A request with data that duplicates an existing customer must return a
        `409 Conflict` HTTP status.
      - The response body for all client errors must contain a clear error
        message.
  - **Server Error Case:**
      - If an unexpected internal system failure occurs, the API must return a
        `500 Internal Server Error` HTTP status.

### A New Challenge: Reusing Tests Across a Stack of Layers

In this scenario, and others we might encounter, our acceptance tests must
remain valid and flexible enough to be applied to all *SUT variants*.

One way to do this is by extracting a generic interface from our
`CustomerTestDriver` type and implementing it for each new layer. That way, each
layer's driver would implement the interface and delegate responsibilities to
the layer below, forming a cascade of *Test Drivers*.

However, let's try an alternative, didactic approach: the
***Optional Inverted Dependency Injection*** with ***Selective Delegation***. It
sounds complicated but is simpler than the names make it appear.

If you look at our `CustomerRepositoryTestDriver` interface, you might have
noted that it contains only the `...Internals...` methods. This is logical
because everything that is a dependency of our `CustomerService` is an
*internal* part of it.

Conversely, if we want to extract an interface to test layers above our
`CustomerService`, we only need the "non-internal" methods. This is what our
`CustomerUpperLayerTestDriver` looks like:

```go
type CustomerUpperLayerTestDriver interface {
	//
	// Act
	ActTryToRegisterACustomer(t *testing.T, request map[string]any, extraArgs map[string]any) map[string]any

	//
	// Assert
	AssertRegistrationShouldSucceed(t *testing.T, result map[string]any, extraArgs map[string]any)
	AssertRegistrationShouldFail(t *testing.T, result map[string]any, extraArgs map[string]any)
	AssertRegistrationShouldFailWithMessage(t *testing.T, result map[string]any, extraArgs map[string]any, targetMessages ...string)
}
```

Our `CustomerServiceTestDriver` now has an `upperLayerTD` property and an
optional `NewCustomerServiceTestDriverWithUpperLayer` constructor to use
upper-layer test driver implementations as an optional inverse dependency. A very
simple selective delegation mechanism transfers the responsibility for calls to
the upper layer when it is set. This mechanism looks like this:

```go
// ...

func (td *CustomerServiceTestDriver) AssertRegistrationShouldSucceed(t *testing.T, result map[string]any, extraArgs map[string]any) {
	t.Helper()

	if td.upperLayerTD != nil {
		td.upperLayerTD.AssertRegistrationShouldSucceed(t, result, extraArgs)

		return
	}

	// Otherwise, perform the original direct call to the service.
	//...
}

// ...
```

To supply specific test data to the upper layers, we also added the
`extraArgs map[string]any` parameter to each method on the
`CustomerUpperLayerTestDriver` interface. In our test data, we can now add an
`extra_args` field, which for this scenario contains the expected HTTP
response status and status code.

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-13" target="_blank">Step 13</a>**_.

-----

### Continue Betting in AI-Assisted Development: Generating the REST Layer

The process for implementing our REST API layer mirrored the repository PoCs. We
began by creating the necessary package at `customer/adapters/presentation/rest`
and scaffolding empty `CustomerRESTAPIHandler` and
`CustomerRESTAPIHandlerTestDriver` types.

With the empty files in place, we provided the AI tool with our full codebase as
context and then sent the following, more structured prompt:

> I need you to complete the implementations of the `CustomerRESTAPIHandler`
> and `CustomerRESTAPIHandlerTestDriver` types. The `CustomerRESTAPIHandler`
> must have a REST HTTP handler that wraps the `Register` method of the
> `CustomerService`. The `CustomerRESTAPIHandlerTestDriver` must implement the
> `CustomerUpperLayerTestDriver` interface, performing its checks on the HTTP
> response by using the expected HTTP status codes provided in the test data.
> The final code must pass all existing acceptance tests by correctly
> implementing the logic for the following status codes: 201 for success, 400
> for invalid data, 409 for duplicates, and 500 for system failures.

The AI generated the code in about two minutes, and the result was a solid first
draft that was very close to passing our tests. The `CustomerRESTAPIHandler`
required only a minor tweak. The `CustomerRESTAPIHandlerTestDriver` needed more
significant adjustments, but using our test suite as a guide, the entire
debugging process was completed in about 10 minutes.

_**NOTE**: see the codebase at that stage here:_
_**<a href="https://github.com/maniosgrivei/go-test-drivers/blob/step-14" target="_blank">Step 14</a>**_.

-----

## The Go Test Driver Pattern

Our case study has led us on a journey, revealing the challenges of evolving a
system guided by acceptance tests, and in turn, evolving the testing suite
itself so that all variants of a system can be validated by the same tests.

In a real-world situation, however, we can anticipate these challenges and
scaffold our systems with a reusable test suite from the beginning, saving
significant time and refactoring effort.

This is possible because we can now define the pattern we've discovered:
the **Go Test Driver Pattern**.

-----

### Application

The **Go Test Driver** pattern is applied to allow a single set of acceptance
tests, conceived at the *service level*, to be reused to validate the
*System Under Test (SUT)* in its many *variants*, which may be composed of
different *upper layers* (like APIs) and *dependencies* (like databases).

It also aims to supply the foundation for a higher-level construct: a
**Domain-Specific Language (DSL)** for testing.

-----

### Architectural Variants

-----

#### The Direct Dependency Injection Variant

This is the standard and simplest way to build test drivers for each layer. It
consists of creating a new test driver implementation for each layer that
(re)implements the methods from the service-level driver.

Its advantage is that it is direct and simple. Its main drawback is the
potentially long **delegation chain** required for the `...Internals...`
methods, as each layer must pass the call down to the next.

-----

#### The Optional Inverted Dependency Injection Variant

This non-standard and slightly more complex approach was the one we used in this
article. It consists of making the test drivers for the upper layers optional,
inverse dependencies of the main service-level test driver.

Its advantages are the reduced interface required for upper-layer drivers and
the elimination of the delegation chain for internal methods. Its main drawbacks
are that it is more complex and can be counter-intuitive.

-----

### Components

-----

#### The Service-Level Test Driver

This is the core of the pattern, as all other components are derived from or
interact with it.

The **Service-Level Test Driver** is a concrete type that contains the methods
for our **Extended Arrange-Act-Assert (AAA+)** pattern. These methods
encapsulate the low-level test logic and provide a clean API for writing
acceptance tests.

Its methods should receive test data in a generic way, such as `map[string]any`,
and be capable of parsing that data into the specific types used by the
service.

Its constructor receives both the service and the dependency test drivers. In
the case of the inverted dependency variant, an optional constructor must also
receive the upper-layer test drivers.

-----

#### The Dependency Test Driver Interfaces

This defines the contract for test drivers for service dependencies, such as
repositories, event brokers, or mailing services.

The **Dependency Test Driver Interfaces** are extracted from the Service-Level
Test Driver but contain **only the `...Internals...` methods**. They may also
contain extra methods for arranging or asserting the internal state of a
specific dependency.

They must never contain `Act...` methods, and it's not recommended that they
contain non-internal `Arrange...` or `Assert...` methods.

-----

#### The Dependency Test Drivers

This is the controlled way to expose the internal state of service dependencies
for testing purposes only.

**Dependency Test Drivers** are concrete types that implement the corresponding
Dependency Test Driver Interface. Their purpose is to expose the internal
elements of a dependency in a controlled way, to be used by the Service-Level
Test Driver.

Its constructor typically receives the underlying dependency that is being
exposed for tests.

-----

#### The Upper-Layer Test Driver Interfaces

This is the contract that makes our acceptance tests compatible with our SUT, no
matter how many layers exist between our tests and our core service.

It is an interface extracted from our Service-Level Test Driver, but its shape
depends on the architectural variant we choose to use.

-----

##### When Using The Direct Dependency Injection Variant

In this case, the **Upper-Level Test Driver Interface** must contain **all**
methods from the Service-Level Test Driver. We would then replace direct calls
to the concrete service driver in our tests with calls to this interface.

-----

##### When Using The Optional Inverted Dependency Injection Variant

In this case, the **Upper-Level Test Driver Interface** must contain **only the**
**non-internal** `Arrange...`, `Act...` and `Assert...` methods. It must not
contain any of the `...Internal...` methods. Our tests will continue to call the
concrete Service-Level Test Driver directly.

-----

#### The Upper-Level Test Driver

This is the concrete implementation of our test primitives for each layer on top
of our service that we want to test. The implementation details depend on the
variant we are using.

-----

##### When Using The Direct Dependency Injection Variant

This driver needs a reference to the layer it is testing (e.g., the REST
handler) and a reference to the underlying test driver it is wrapping (which
would also be typed as the Upper-Level Test Driver Interface). It must implement
the primitives for all non-internal methods and **delegate** all calls for
internal methods to the underlying driver.

-----

##### When Using The Optional Inverted Dependency Injection Variant

This driver only needs a reference to the upper layer it is testing. It only
needs to implement the methods defined in the (smaller) Upper-Level Test Driver
Interface. It is then used as an **inverse dependency** of the Service-Level
Test Driver, which will selectively **upward-delegate** calls to it.

-----

### Putting the Pattern into Practice

Our case study led us on a didactic journey, starting from a simple design and
evolving it step-by-step. This was useful for discovery, but in a real-world
project, we can anticipate these challenges and apply the finished pattern from
the beginning, saving significant time and effort.

What follows is an opinionated guide to scaffolding a new project using the
**Go Test Driver** pattern. While we still follow the YAGNI principle, we start
with a slightly more realistic foundation than global variables, assuming at
least a concrete type to encapsulate our service and its core dependencies.

-----

#### Step 1: Define the Skeleton

We begin by creating the most basic elements of our architecture, all within the
same initial package:
-   **Entities:** Go structs for our core domain (e.g., the `Customer` struct).
-   **Service:** A concrete type to orchestrate the business logic (e.g.,
    `CustomerService`).
-   **Dependencies:** Initial, types for any essential dependencies (e.g.,
    an `InMemoryRepository`).

In parallel, in a file marked with `//go:build test`, we create the initial
concrete types for our `ServiceLevelTestDriver` and `DependencyTestDriver`.

-----

#### Step 2: Build a Flexible Test Harness

Even if dependencies and upper layers are not yet defined, a modern software
product will almost certainly have multiple variants. We can therefore
**preventively** create a minimal test harness that encourages this decoupling
from the start. This includes:
-   A small setup function that can build our SUT with different variants.
-   A basic structure for loading test data from external files (e.g., YAML).

The advantage of this minimal "boilerplate" is that it acts as a **guardrail**.
It makes building a decoupled system the path of least resistance and makes
adding a new SUT variant (like a new database) a trivial change to the test
setup.

However, we must be careful. This initial harness should be a thin, flexible
skeleton, not a heavy, rigid framework. We are setting up a structure to guide
our evolution, not pre-emptively building features we don't need.

-----

#### Step 3: Write Gherkin Scenarios

Before writing any Go test code, we translate our User Stories and Acceptance
Criteria into Gherkin-like scenarios. This is a mandatory step that forces us to
think about behavior first. We can start with just one or two of the most
complex or risky scenarios.

To succeed, we must adopt the correct perspective, which I call
**The Service-Level Point of View**.

Drawing from our hardware analogy, we are the technicians, and our test drivers
are the "test pads." Our Gherkin scenarios should be written from this
privileged perspective: not so low-level that they are coupled to implementation
details, but not so high-level that they ignore the critical dependencies we
need to control and the side effects we need to assert.

-----

#### Step 4: From Scenarios to Test Signatures

Once we have our first Gherkin scenarios, we can write our first Go tests,
translating the Gherkin steps directly into our `Arrange/Act/Assert` structure.
This process will naturally define the required method signatures for our test
driver interfaces.

-----

#### Step 5: The Red-Green-Refactor Cycle

The development process follows the classic TDD cycle: write a test and watch it
fail (**Red**), write the minimal code to make the test pass (**Green**), and then
improve both the test and the functional code (**Refactor**).

Applying this to our pattern requires an upfront investment. Initially, building
the test drivers, data schemas, and helper functions requires more effort than a
simple unit test.

However, once this test harness is stable, the effort to add new test cases
drops dramatically. You can validate complex new behaviors simply by adding a new
YAML file, allowing you to write new acceptance tests at a remarkable speed.

-----

#### Step 6: Evolving the System and the Tests

As the system expands, the test suite must co-evolve with it. We can think of
this expansion in two dimensions:

-   **Horizontal Expansion:** Adding new layers or dependencies (e.g., a gRPC
    endpoint or a new database backend). This requires implementing new test
    drivers for those components and adding a new SUT variant to the test
    harness.

-   **Vertical Expansion:** Adding new features to an existing service (e.g., an
    `UpdateCustomer` or `DeleteCustomer` method). This typically only requires
    adding new test driver methods and new YAML test data, reusing the existing
    drivers.

The key benefit of this pattern is the reusability of your testing components.
The `Arrange` method you build to populate a repository for a registration test
will be reused to set up the state for an update or a delete test. The
`extra_args` you use to check an HTTP status code for a REST API can be reused
to check a gRPC status code.

The result is a virtuous cycle: the more the system expands, the cheaper it
becomes to add the next test, **provided the initial framework was properly**
**conceived**.

-----

## Practical Usage and Tooling

Beyond the initial design, discipline and good tooling are essential to get the
most out of this pattern.

The standard Go toolchain offers great support for this workflow. With some
minor setup and the right commands, you can achieve fine-grained test selection
and get a clear, graphical view of your test coverage. Let's see how.

-----

### Code Editor Setup (`gopls`)

As mentioned in our case study, your code editor will likely show errors when
trying to find your test helper functions. This is because the `//go:build test`
constraint hides them from the default build process.

To solve this, you must configure the Go language server, `gopls`, to include
the `test` build tag.

If you are using **VS Code** or a variant like VSCodium, go to
**File \> Preferences \> Settings**, search for **gopls**, find the
**Go: Build Tags** option, and add `test` in the text field. For other editors,
consult their documentation on how to configure build tags for `gopls`.

-----

### Running Specific Tests

The `go test -run` command accepts a regular expression that filters which tests
to run. Because our test names form a predictable path (e.g.,
`TestGroup/Variant/Scenario/Case`), we can run tests with fine granularity.

  - **Run a specific test case in a single system variant**

    ```sh
    go test -v -tags test \
	-run '^TestRegisterCustomer/with_system_variant_reference/should_register_a_customer_with_valid_data/when_is_a_company$' \
	github.com/maniosgrivei/go-test-drivers/test/acceptance/customer
    ```

  - **Run all tests for a single system variant**

    ```sh
    go test -v -tags test \
	-run '^TestRegisterCustomer/with_system_variant_sqlite$' \
	github.com/maniosgrivei/go-test-drivers/test/acceptance/customer
    ```

  - **Run a specific scenario across all system variants**

    ```sh
    go test -v -tags test \
	-run '^TestRegisterCustomer/with_system_variant_.*/should_register_a_customer_with_valid_data$' \
	github.com/maniosgrivei/go-test-drivers/test/acceptance/customer
    ```

  - **Run a specific scenario across specific system variants**

    ```sh
    go test -v -tags test \
	-run '^TestRegisterCustomer/with_system_variant_(sqlite|badger)/should_register_a_customer_with_valid_data$' \
	github.com/maniosgrivei/go-test-drivers/test/acceptance/customer
    ```

-----

### Measuring Test Coverage

A common question is, "How do I get the test coverage of my application code if
my tests are in a different package?" The Go toolchain provides a simple answer.

  - **Calculating Coverage**

    The `-coverpkg` flag tells the `go test` command which package(s) to measure
    coverage against. In our case study, the following command runs our
    acceptance tests while measuring coverage for all code inside the `customer`
    package and its sub-packages:

    ```sh
    go test -tags test \
	-coverpkg=./customer/... \
	-run '^TestRegisterCustomer$' \
    github.com/maniosgrivei/go-test-drivers/test/acceptance/customer
    ```

    Result:

    ```
    ok  github.com/maniosgrivei/go-test-drivers/test/acceptance/customer 2.033s coverage: 94.3% of statements in ./customer/...
    ```

  - **Visualizing the Coverage Profile**

    You can generate a graphical, browser-based report of your test coverage.
    The `-coverprofile` flag writes the coverage data to a file, and
    `go tool cover` can then convert this file to HTML.

    The following command chain calculates the coverage profile and immediately
    opens the visual report in your browser:

    ```sh
    go test -tags test \
	-coverpkg=./customer/... -coverprofile=/tmp/cover.out \
	-run '^TestRegisterCustomer$' \
    github.com/maniosgrivei/go-test-drivers/test/acceptance/customer \
    && go tool cover -html=/tmp/cover.out
    ```

-----

## Conclusion

We began this journey with a simple goal: to write clean, decoupled acceptance
tests in Go without resorting to traditional mocks. This led us down a path of
discovery, from the pain of testing tightly-coupled code to the power of a
data-driven framework, culminating in the formal **Go Test Driver** pattern. We
demonstrated that a well-designed test suite is not a burden but an asset that
guides evolution, absorbs change, and even validates AI-generated code.

But where does this pattern fit in the broader testing landscape? The Go Test
Driver pattern is a powerful tool for acceptance and integration testing, but
it does not replace **unit tests**. The two must live in harmony.

Our acceptance tests, running against real components via the "Test Pad" Seam,
are perfect for validating the integrated behavior of our system—the "what."
They ensure all the pieces work together as expected from the
**Service-Level Point of View**.

However, for testing the "how"—the complex internal logic of a single component
in complete isolation—**unit tests using traditional test doubles are often the**
**superior tool**. The complex ID generation algorithm from our case study, for
example, is a perfect candidate for exhaustive unit testing, where a mocked
time source can be used to verify every edge case.

Ultimately, Test-Driven Development is a discipline of design. The Go Test
Driver pattern is a key that unlocks this discipline for acceptance testing,
allowing us to build robust, maintainable, and truly testable systems in Go.

-----

### Next Steps: Toward a True DSL

By employing the Go Test Driver pattern, we achieved high levels of readability
and maintainability in our acceptance tests. Our test cases are clean, and the
framework is flexible enough to handle significant evolution in our system.

However, the tests still contain technical details (like `result` maps and
`extraArgs`) that can be unfamiliar to non-technical stakeholders.

The final step in this journey is to achieve truly business-comprehensible
tests by establishing a **Domain-Specific Language (DSL)**. We can use our
existing test drivers as the foundational building blocks for this new, higher
level of abstraction.

I will explore this topic in-depth in a future article. But for now, let's look
at a small sample of what this DSL could look like:

```go
//go:build test

package dsl

// ...

// RegisterCustomer is a high-level DSL function that encapsulates the entire
// "register customer" scenario.
func (d *CRMDSL) RegisterCustomer(t *testing.T, request map[string]any, extraArgs map[string]any) {
	t.Helper()

	// The DSL method encapsulates the full Act-Assert sequence.
	result := d.customerTestDriver.ActTryToRegisterACustomer(t, request, extraArgs)

	d.customerTestDriver.AssertRegistrationShouldSucceed(t, result, extraArgs)

	d.customerTestDriver.AssertInternalsCustomerShouldBeProperlyRegistered(t, request, extraArgs)
}

// ...
```
