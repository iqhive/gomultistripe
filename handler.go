// Package stripe provides a versioned handler abstraction for Stripe API interactions.
// It allows easy switching and support for multiple Stripe API versions at runtime.
// This package is designed to be easily extracted into a standalone module in the future.
//
// To add support for a new Stripe API version, implement the Handler interface and register it.
package gomultistripe

import (
	"context"
)

// Customer represents a Stripe customer in a version-agnostic way.
type Customer struct {
	ID       string
	Name     string
	Email    string
	Phone    string
	Postcode string
}

// PaymentMethod represents a Stripe payment method in a version-agnostic way.
type PaymentMethod struct {
	ID        string
	Type      string
	Last4     string
	Brand     string
	ExpMonth  uint
	ExpYear   uint
	IsDefault bool
}

// PaymentIntent represents a Stripe payment intent in a version-agnostic way.
type PaymentIntent struct {
	ID            string
	Amount        int64
	Currency      string
	Status        string
	ClientSecret  string
	CustomerID    string
	PaymentMethod string
}

// Handler abstracts Stripe API interactions and versioning.
type Handler interface {
	// Version returns the Stripe API version this handler implements.
	Version() string
	// CreateCustomer creates a customer in Stripe for this version.
	CreateCustomer(ctx context.Context, params *Customer) (*Customer, error)
	// UpdateCustomer updates a customer in Stripe for this version.
	UpdateCustomer(ctx context.Context, customerID string, params *Customer) (*Customer, error)
	// GetPaymentMethods retrieves payment methods for a customer in Stripe for this version.
	GetPaymentMethods(ctx context.Context, customerID string) ([]*PaymentMethod, error)
	// AttachPaymentMethod attaches a payment method to a customer (required for Elements flow).
	AttachPaymentMethod(ctx context.Context, customerID string, paymentMethodID string) (*PaymentMethod, error)
	// DetachPaymentMethod detaches a payment method from a customer (for secure removal).
	DetachPaymentMethod(ctx context.Context, paymentMethodID string) error
	// CreatePaymentIntent creates a PaymentIntent for secure payment confirmation.
	CreatePaymentIntent(ctx context.Context, params *PaymentIntent) (*PaymentIntent, error)
	// RetrievePaymentIntent retrieves a PaymentIntent by ID.
	RetrievePaymentIntent(ctx context.Context, paymentIntentID string) (*PaymentIntent, error)
	// Example: CreateCustomer, Charge, etc. Add more as needed.
}

// registry holds all registered Stripe handlers by version.
// NB: go init functions run in series, so using a map to register handlers should be thread
//
//	safe, assuming there are no other registrations other than the ones in init() functions
var registry = make(map[string]Handler)

// RegisterHandler registers a handler for a specific Stripe API version.
func RegisterHandler(h Handler) {
	registry[h.Version()] = h
}

// GetHandler returns the handler for the given version, or nil if not found.
func GetHandler(version string) Handler {
	return registry[version]
}
