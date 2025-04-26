# Stripe Handler Versioning System

This package implements a versioned handler abstraction for Stripe API interactions, allowing support for multiple Stripe API versions at runtime. The design enables easy upgrades, backward compatibility, and clear separation of logic for each Stripe API version.

The primary usecase for this package is to support the [Elements](https://docs.stripe.com/elements) product, so the package functions are designed to be used with that.

The main reason for the package is that when using Stripe's API in Go, you end up with version-specific package references all throughout your code, and as your code base grows, it becomes more and more difficult to upgrade to a new version of the Stripe API. This package allows you to upgrade to a new version of the Stripe API by making a single change to your codebase.

## Supported Stripe API Versions

This package currently supports the following Stripe API versions:

- **v75**
- **v76**
- **v78**
- **v79**
- **v80**
- **v81**
- **v82**

We will be setting up automation to automatically add new versions to the package as they are released.

Each version has its own handler implementation and can be selected at runtime.

## Overview

- **Versioned Handlers:** Each supported Stripe API version has its own handler implementation (e.g., `handler_v80.go` for v80, `handler_v81.go` for v81, `handler_v82.go` for v82).
- **Interface Abstraction:** All handlers implement the `Handler` interface defined in `handler.go`.
- **Registration:** Handlers self-register via their `init()` function, making them available for runtime selection.
- **Version-Agnostic Models:** Common types (`Customer`, `PaymentMethod`, `PaymentIntent`) are defined in a version-agnostic way and mapped to/from Stripe SDK types in each handler.

## Directory Structure

- `handler.go`: Defines the `Handler` interface, version-agnostic models, and handler registry.
- `handler_vXX.go`: Implements the handler for Stripe API version XX (e.g., `handler_v80.go` for v80, `handler_v81.go` for v81, `handler_v82.go` for v82).
- `handler_test.go`: Contains tests for handler registration and basic functionality.

## Handler Interface

All handlers must implement the following interface:

## Adding a New Stripe API Version

To add support for a new Stripe API version (e.g., v83):

1. **Copy an Existing Handler:**
   - Duplicate the most recent handler file (e.g., `handler_v82.go` → `handler_v83.go`).

2. **Update Imports:**
   - Change all Stripe SDK imports to the new version (e.g., `github.com/stripe/stripe-go/v83`).
   - Add the new Stripe SDK version to your `go.mod` using `go get github.com/stripe/stripe-go/v83`.

3. **Rename the Handler Type:**
   - Update the handler struct and registration to match the new version (e.g., `HandlerV83`).

4. **Update the Version Method:**
   - Ensure `Version()` returns the correct version string (e.g., `"v83"`).

5. **Review API Changes:**
   - Check the [Stripe API changelog](https://stripe.com/docs/upgrades#api-changelog) and the Go SDK release notes for breaking changes or new features.
   - Update method implementations as needed to accommodate API changes.

6. **Register the Handler:**
   - Ensure the handler is registered in its `init()` function:
     ```go
     func init() {
         gomultistripe.RegisterHandler(&HandlerV83{})
     }
     ```

7. **Test the Handler:**
   - Add or update tests in `handler_test.go` to cover the new version.

8. **Update the README:**
   - Add the new version to the supported versions list above.

## Conventions and Best Practices

- **Documentation:** Update this README and add comments to new handlers to explain any version-specific logic. Always update the supported versions list when adding a new version.

## Example: Creating a New Handler

Suppose Stripe releases API version v83. To add support:

1. Copy `handler_v82.go` to `handler_v83.go`.
2. Update all imports from `v82` to `v83`.
3. Rename the handler struct to `HandlerV83` and update the `Version()` method.
4. Update the `init()` function to register `HandlerV83`.
5. Run `go get github.com/stripe/stripe-go/v83` to add the new SDK version to your dependencies.
6. Review the Stripe API changelog and Go SDK docs for any changes.
7. Update the implementation as needed.
8. Test thoroughly.
9. Update the README.

## Troubleshooting

- **Missing Stripe SDK Version:** If you see errors like `no required module provides package github.com/stripe/stripe-go/vXX`, run `go get github.com/stripe/stripe-go/vXX` to add the required version to your `go.mod`.
- **Package Name Mismatch in Tests:** Ensure the package name in your test files matches the implementation file (e.g., `package v74`).
- **Handler Not Registered:** Make sure your handler is registered in an `init()` function.
