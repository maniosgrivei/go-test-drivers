# Go Test Drivers

## Building the Foundation for DSL-Driven Tests in Go

<div style="text-align: center;" width= "100%">
  <img src="./doc/images/go-test-drivers-banner.png" alt="Go Test Drivers" width="100%" height="auto" allign="center" />
</div>

What if your tests could speak the language of your business? Instead of
imperative scripts, imagine a suite of tests that reads like a clear
specification of your system's behavior.

This is the principle behind **Go Test Drivers**.

This article is inspired by the video
**[How to Write Acceptance Tests](https://www.youtube.com/watch?v=JDD5EEJgpHU)**
from the
**[Modern Software Engineering](https://www.youtube.com/@ModernSoftwareEngineeringYT)**
YouTube channel by **[Dave Fraley](https://www.google.com/search?q=dave+farley)**.
We will explore how to implement his concept of **"Test Protocol Drivers"** in
Go, treating them as what they are: **specialized mechanisms that translate**
**terms from a Domain-Specific Language (DSL) into direct interactions with**
**a System Under Test (SUT)**.

By the end, you'll have a practical framework for building this foundation,
**unlocking the ability to take the next step**: writing tests that are deeply
decoupled, easy to maintain, and serve as living documentation for your
application.

## The Black-Box Dilemma: Control and Observation
Before we build our first driver, we need to understand two subtle but critical
challenges that high-level acceptance tests face, especially in Go.

1. **The Control Problem: Handling Internals**

Often, acceptance tests are placed in a separate packages to force them to
interact with the system (SUT) just like a real client would. This is great for
ensuring correctness, but it creates a problem: the test code is blocked from
accessing any internal, unexported parts of the system.

This leads to a crucial question: If our test can't reach inside the system, how
can we force it into a specific state—like simulating a database failure or a
specific network error—in a clean and deterministic way?

2. **The Observation Problem: Asserting Side Effects**

This is a direct consequence of the control problem. Real-world systems produce
side effects—a record is created, a file is written, or a message is sent to a
queue. If our test only sees the final output of a function, how can it verify
that the correct side effect actually occurred?

This brings us to the second question: If our test is outside the system, how
can we directly observe and assert that an internal state change—or even an
external one—happened exactly as we intended?

## Case Study

### Functions as Drivers

Imagine you're starting the development of a new CRM system from scratch. At
this moment, nothing is defined—the project is a blank slate. You receive the
first User Story and its Acceptance Criteria.

#### User Story
As a Sales Person, I want to register a new customer on the CRM system by
providing their pertinent data: Name, E-mail, and Phone Number. As a result, I
want to receive the unique ID for that customer.

#### Acceptance Criteria

- Success Case:
  - Upon successful registration, a unique ID for the new customer is returned.

- Validation Errors (Missing Data):
  - If the Name is not provided, the system must reject the registration with a
    specific "Name is required" error.
  - If the E-mail is not provided, the system must reject the registration with
    a specific "E-mail is required" error.
  - If the Phone Number is not provided, the system must reject the
    registration with a specific "Phone Number is required" error.

- Uniqueness Errors (Duplicate Data):
  - If the Name already exists, the system must reject the registration with a
    specific "Name already in use" error.
  - If the E-mail already exists, the system must reject the registration with
    a specific "E-mail already in use" error.
  - If the Phone Number already exists, the system must reject the registration
    with a specific "Phone Number already in use" error.

- Generic Errors:
  - If any other unexpected error occurs (e.g., a database failure), the system
    must reject the registration with a generic error message, instructing the
    user to contact support.

#### Step 1: Code, Results, and a Hidden Problem

So, after tackling that first User Story and following several 
_**Red-Green-Refactor**_ cycles, we have a working set of tests and code, which
you can see in the files below:

- [customer/customer_test.go](https://github.com/maniosgrivei/go-test-drivers/blob/step-01/customer/customer_test.go)
- [customer/customer.go](https://github.com/maniosgrivei/go-test-drivers/blob/step-01/customer/customer.go)

Also, our tests are passing and producing meaningful output that is
understandable even to non-technical people, as you can see here:

- [Step 1 test results](https://github.com/maniosgrivei/go-test-drivers/blob/step-01/doc/test-results/step-01.txt)

On the surface, everything looks perfect. But this seemingly successful result
hides a critical flaw.

##### The Flaw: We Can't Truly Assert Side Effects

Look at this "happy path" test:

```go
func TestRegisterCustomer(t *testing.T) {
	t.Run("should register a customer with valid data", func(t *testing.T) {
		r := require.New(t)

		request := RegisterRequest{
			Name:  dfltName,
			Email: dfltEmail,
			Phone: dfltPhone,
		}

        Init()

		id, err := Register(request)

		r.NoError(err)
		r.NotEmpty(id)
	})

    // ...

}
```

The problem is that the only thing we really assert is the return value of the
function. Nothing ensures that the customer was actually registered in the
repository.

You might argue that the test for duplications certifies that our repository
contains the data, but that's not entirely true. When we look at the
implementation, that test only confirms that the **INDEXES** contain the data;
there is nothing proving that the complete customer object itself was correctly
added to the `customer` slice. We are blind to the true side effect.