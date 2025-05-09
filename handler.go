// Package stripe provides a versioned handler abstraction for Stripe API interactions.
// It allows easy switching and support for multiple Stripe API versions at runtime.
// This package is designed to be easily extracted into a standalone module in the future.
//
// To add support for a new Stripe API version, implement the Handler interface and register it.
package gomultistripe

import (
	"context"
	"time"
)

// Customer represents a Stripe customer in a version-agnostic way.
type Customer struct {
	ID        string
	Name      string
	Email     string
	Phone     string
	Postcode  string
	Metadata  map[string]string
	CreatedAt time.Time
}

// PaymentMethod represents a Stripe payment method in a version-agnostic way.
type PaymentMethod struct {
	ID         string
	CustomerID string
	Type       string
	Last4      string
	Brand      string
	ExpMonth   uint
	ExpYear    uint
	IsDefault  bool
	Metadata   map[string]string
	CreatedAt  time.Time
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
	Metadata      map[string]string
	CreatedAt     time.Time
}

// Subscription represents a Stripe subscription in a version-agnostic way.
type Subscription struct {
	ID                string
	CustomerID        string
	Status            string
	PriceID           string
	CurrentPeriodEnd  int64
	CancelAtPeriodEnd bool
	CanceledAt        int64
	Metadata          map[string]string
	CreatedAt         time.Time
}

// CallbackEventType represents the type of Stripe event received.
type CallbackEventType string

const (
	EventSetupIntentSucceeded                 CallbackEventType = "setup_intent.succeeded"
	EventPaymentIntentCanceled                CallbackEventType = "payment_intent.canceled"
	EventPaymentIntentPaymentFailed           CallbackEventType = "payment_intent.payment_failed"
	EventPaymentIntentSucceeded               CallbackEventType = "payment_intent.succeeded"
	EventPaymentIntentAmountCapturableUpdated CallbackEventType = "payment_intent.amount_capturable_updated"

	// Subscription events
	EventCustomerSubscriptionCreated      CallbackEventType = "customer.subscription.created"
	EventCustomerSubscriptionUpdated      CallbackEventType = "customer.subscription.updated"
	EventCustomerSubscriptionDeleted      CallbackEventType = "customer.subscription.deleted"
	EventCustomerSubscriptionTrialWillEnd CallbackEventType = "customer.subscription.trial_will_end"
	EventCustomerSubscriptionPaused       CallbackEventType = "customer.subscription.paused"
	EventCustomerSubscriptionResumed      CallbackEventType = "customer.subscription.resumed"

	// Invoice events
	EventInvoicePaymentSucceeded CallbackEventType = "invoice.payment_succeeded"
	EventInvoicePaymentFailed    CallbackEventType = "invoice.payment_failed"
	EventInvoiceCreated          CallbackEventType = "invoice.created"
	EventInvoiceUpcoming         CallbackEventType = "invoice.upcoming"

	// Refund events
	EventRefundCreated  CallbackEventType = "refund.created"
	EventRefundUpdated  CallbackEventType = "refund.updated"
	EventRefundFailed   CallbackEventType = "refund.failed"
	EventChargeRefunded CallbackEventType = "charge.refunded"
)

// CallbackEvent is a version-agnostic representation of a Stripe webhook event.
type CallbackEvent struct {
	Type CallbackEventType

	// Common metadata fields
	Metadata     map[string]string
	PreAllocated string
	ValidateOnly string

	// SetupIntent fields
	SetupIntentID   string
	PaymentMethodID string
	CardBrand       string
	CardExpMonth    uint
	CardExpYear     uint
	CardLast4       string

	// PaymentIntent fields
	PaymentIntentID  string
	Amount           int64
	AmountCapturable int64
	Status           string

	// Payment error fields
	LastPaymentErrorCode            string
	LastPaymentErrorMsg             string
	LastPaymentErrorDeclineCode     string
	LastPaymentErrorPaymentMethodID string
	LastPaymentErrorChargeID        string

	// Subscription fields
	SubscriptionID    string
	CustomerID        string
	CurrentPeriodEnd  int64
	CancelAtPeriodEnd bool
	CanceledAt        int64
	CreatedAt         time.Time

	// Invoice fields
	InvoiceID    string
	InvoiceLines []InvoiceLine

	// Refund fields
	RefundID     string
	RefundAmount int64
	RefundReason string
	RefundStatus string
	ChargeID     string
	Currency     string
}

type InvoiceLine struct {
	ID             string
	Amount         int64
	Currency       string
	Description    string
	SubscriptionID string
}

// Handler abstracts Stripe API interactions and versioning.
type Handler interface {
	// Version returns the Stripe API version this handler implements.
	Version() string
	// SetSecretKey sets the Stripe secret key for this handler.
	SetSecretKey(secretKey string)
	// SetWebhookSecret sets the Stripe webhook secret for this handler.
	SetWebhookSecret(webhookSecret string)
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
	// CreateSubscription creates a subscription for a customer.
	CreateSubscription(ctx context.Context, customerID string, priceID string) (*Subscription, error)
	// ListSubscriptions lists subscriptions for a customer.
	ListSubscriptions(ctx context.Context, customerID string) ([]*Subscription, error)
	// UpdateSubscription updates a subscription (e.g., change price, cancel at period end).
	UpdateSubscription(ctx context.Context, subscriptionID string, cancelAtPeriodEnd bool, newPriceID string) (*Subscription, error)
	// CancelSubscription cancels a subscription immediately or at period end.
	CancelSubscription(ctx context.Context, subscriptionID string, atPeriodEnd bool) (*Subscription, error)
	// Example: CreateCustomer, Charge, etc. Add more as needed.

	// HandleWebhook processes a Stripe webhook payload and sends events to the channel.
	HandleWebhook(payload []byte, sigHeader string) (*CallbackEvent, error)
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
