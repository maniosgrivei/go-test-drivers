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

#### User Story B: A New Format for Customer ID

> As a ***Customer***, I want to receive a more meaningful and easy-to-record
> *Customer ID*.

-----

##### Acceptance Criteria

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

#### User Story C: Enhanced Customer Data Validation

> As a ***Compliance Officer***, I want the system to perform stricter
> validation on customer data to ensure data quality.

-----

##### Acceptance Criteria

  - The **customer name** must be validated for length, allowed characters, and
    structure.
  - The **customer e-mail** must be validated for length, characters, and a
    valid format.
  - The **customer phone** must be validated for length and a valid international
    format.

These stories introduce complex rules. Let's analyze the impact of each one and
see how our testing framework holds up under pressure.

-----

#### Implementing the New Customer ID Format

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

#### Implementing Enhanced Customer Data Validation

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

#### User Story D: Data Storage Flexibility

> As a **System Administrator** preparing to deploy the CRM, I need assurance
> that the system can integrate with my company's certified storage
> infrastructure (e.g., SQL or NoSQL based), so that it will comply with our
> mandatory data policies.

-----

##### Acceptance Criteria

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

#### Implementing Data Storage PoCs

Before we can implement the different storage backends, we must first refactor
our code to be more robust and organized.

-----

##### Moving Everithing to `struct`s

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

##### Decoupling Service and Repository

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

###### A Two-Level Driver: Separating Concerns

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

###### Composing the System with Dependency Injection

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
