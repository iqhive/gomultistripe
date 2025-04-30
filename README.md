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

## Stripe Version Management Tool

We provide a tool to help manage Stripe Go SDK versions. The `update_stripe_versions` tool:

1. Updates the 5 most recent existing versions to their latest minor and patch releases
2. Automatically adds new major versions by copying the most recent version's files and updating imports
3. Runs tests to verify changes work correctly
4. Commits changes to git (when not in dry-run mode)

### Using the Tool

```bash
# Navigate to the tool directory
cd cmd/update_stripe_versions

# Build and run the tool (makes actual changes)
make run

# Preview changes without making them (dry-run mode)
make dry-run
```

See the [tool's README](cmd/update_stripe_versions/README.md) for more details.

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

## Using Subscriptions

This package provides a version-agnostic way to manage Stripe subscriptions via the `Handler` interface. The following methods are available for subscription management:

- `CreateSubscription(ctx, customerID, priceID)`
- `ListSubscriptions(ctx, customerID)`
- `UpdateSubscription(ctx, subscriptionID, cancelAtPeriodEnd, newPriceID)`
- `CancelSubscription(ctx, subscriptionID, atPeriodEnd)`

### Subscription Model

The `Subscription` struct is defined as:

```go
// Subscription represents a Stripe subscription in a version-agnostic way.
type Subscription struct {
    ID                string
    CustomerID        string
    Status            string
    PriceID           string
    CurrentPeriodEnd  int64
    CancelAtPeriodEnd bool
    CanceledAt        int64
    Created           int64
}
```

### Creating a Subscription

To create a subscription for a customer to a specific price:

```go
sub, err := handler.CreateSubscription(ctx, customerID, priceID)
if err != nil {
    // handle error
}
fmt.Printf("Created subscription: %+v\n", sub)
```

- `customerID`: The ID of the Stripe customer.
- `priceID`: The ID of the Stripe price (recurring product/plan).

### Listing Subscriptions

To list all subscriptions for a customer:

```go
subs, err := handler.ListSubscriptions(ctx, customerID)
if err != nil {
    // handle error
}
for _, sub := range subs {
    fmt.Printf("Subscription: %+v\n", sub)
}
```

### Updating a Subscription

You can update a subscription to change its price or set it to cancel at the end of the current period:

```go
updatedSub, err := handler.UpdateSubscription(ctx, subscriptionID, cancelAtPeriodEnd, newPriceID)
if err != nil {
    // handle error
}
fmt.Printf("Updated subscription: %+v\n", updatedSub)
```
- `subscriptionID`: The ID of the subscription to update.
- `cancelAtPeriodEnd`: If true, the subscription will be canceled at the end of the current period.
- `newPriceID`: (Optional) The new price ID to switch the subscription to. Pass an empty string to leave unchanged.

### Canceling a Subscription

To cancel a subscription immediately or at the end of the period:

```go
canceledSub, err := handler.CancelSubscription(ctx, subscriptionID, atPeriodEnd)
if err != nil {
    // handle error
}
fmt.Printf("Canceled subscription: %+v\n", canceledSub)
```
- `subscriptionID`: The ID of the subscription to cancel.
- `atPeriodEnd`: If true, the subscription will be canceled at the end of the current period; if false, it will be canceled immediately.

### Notes
- All methods require a valid `context.Context` as the first argument.
- The handler instance should be selected for the desired Stripe API version.
- Returned `Subscription` objects contain key information such as status, price, and period end timestamps.
- Error handling is essential for production use.

## Using Callback (Webhook) Handlers

This package provides a version-agnostic way to handle Stripe webhook events via the `CallbackHandler` interface. Each versioned handler implements its own callback handler, which parses Stripe webhook payloads and sends normalized events to a Go channel for processing.

### Supported Events

The following Stripe event types are supported:

| Event Type                              | Object           | Description                                 | Use Case |
|-----------------------------------------|------------------|---------------------------------------------|----------|
| setup_intent.succeeded                  | SetupIntent      | Triggered when a SetupIntent has successfully completed, and the payment method is ready to use for future payments. | Store/confirm payment method for future use |
| payment_intent.canceled                 | PaymentIntent    | Sent when a PaymentIntent is canceled, indicating that the intended payment will not take place. | Update payment status as canceled |
| payment_intent.payment_failed           | PaymentIntent    | Occurs when a PaymentIntent fails, usually due to authentication or payment method issues. | Error handling, dunning, customer notification |
| payment_intent.succeeded                | PaymentIntent    | Fired when a PaymentIntent has been confirmed and the payment is successfully completed. | Confirm successful payment |
| payment_intent.amount_capturable_updated| PaymentIntent    | Triggered when the amount of a PaymentIntent changes, and it is now ready to be captured (useful for manual capture workflows). | Mark payment as ready for capture |
| customer.subscription.created           | Subscription     | Triggered when a new subscription is created for a customer. | Track new signups |
| customer.subscription.updated           | Subscription     | Sent when the subscription changes (like plan upgrades, downgrades, or changes in quantity or billing cycle). | Track plan changes, upgrades, downgrades |
| customer.subscription.deleted           | Subscription     | Occurs when a subscription is canceled or deleted. | Track cancellations or removals |
| customer.subscription.trial_will_end    | Subscription     | Sent a few days before the trial period of a subscription ends. | Remind users about trial ending |
| customer.subscription.paused            | Subscription     | Triggered when a subscription is paused. | Restrict access temporarily |
| customer.subscription.resumed           | Subscription     | Sent when a previously paused subscription is resumed. | Restore access |
| invoice.payment_succeeded               | Invoice          | Fired when a billing invoice for a subscription is successfully paid. | Confirm successful recurring charge |
| invoice.payment_failed                  | Invoice          | Occurs when an invoice payment attempt fails. | Dunning, alerting customers |
| invoice.created                         | Invoice          | Sent when a new invoice (recurring billing) is created. | Record keeping, notification |
| invoice.upcoming                        | Invoice          | Triggered a short time before an invoice for a subscription is finalized. | Notify user of upcoming charge |

### CallbackEvent Fields

The `CallbackEvent` struct contains all the fields you need for billing and account logic. The fields populated depend on the event type. See the table below for the minimum fields per event:

- **Metadata**: All Stripe metadata fields are now available in the `Metadata` map (e.g., `evt.Metadata["SPID"]`, `evt.Metadata["AccountType"]`, etc.).
- **InvoiceLines**: For invoice events, the `InvoiceLines` field contains detailed information about each line item on the invoice.

#### InvoiceLine Structure

```go
// InvoiceLine represents a single line item on a Stripe invoice.
type InvoiceLine struct {
    ID             string
    Amount         int64
    Currency       string
    Description    string
    SubscriptionID string
}
```

#### Example: Accessing Metadata and Invoice Lines

```go
// Accessing metadata fields
spid := evt.Metadata["SPID"]
accountType := evt.Metadata["AccountType"]
externalID := evt.Metadata["AccountExternalID"]

// Accessing invoice lines
for _, line := range evt.InvoiceLines {
    fmt.Printf("Line: %s, Amount: %d, Description: %s\n", line.ID, line.Amount, line.Description)
}
```

| Event Type                              | Required Metadata Fields (in evt.Metadata) | Other Key Fields in CallbackEvent (if present) |
|-----------------------------------------|--------------------------------------------|-------------------------------------------------|
| setup_intent.succeeded                  | SPID, AccountType, AccountExternalID       | SetupIntentID, PaymentMethodID, CardBrand, CardExpMonth, CardExpYear, CardLast4 |
| payment_intent.canceled                 | SPID, AccountType, AccountExternalID       | PaymentIntentID, Amount, PaymentMethodID, PreAllocated |
| payment_intent.payment_failed           | SPID, AccountType, AccountExternalID       | PaymentIntentID, Amount, PaymentMethodID, PreAllocated, LastPaymentErrorCode, LastPaymentErrorMsg, LastPaymentErrorDeclineCode, LastPaymentErrorPaymentMethodID, LastPaymentErrorChargeID, Status, ValidateOnly |
| payment_intent.succeeded                | SPID, AccountType, AccountExternalID       | PaymentIntentID, Amount, PaymentMethodID, PreAllocated, Status, ValidateOnly |
| payment_intent.amount_capturable_updated| SPID, AccountType, AccountExternalID       | PaymentIntentID, Amount, AmountCapturable, Status, ValidateOnly |
| customer.subscription.created           | SPID, AccountType, AccountExternalID       | SubscriptionID, CustomerID, Status, CurrentPeriodEnd, CancelAtPeriodEnd, CanceledAt, Created |
| customer.subscription.updated           | SPID, AccountType, AccountExternalID       | SubscriptionID, CustomerID, Status, CurrentPeriodEnd, CancelAtPeriodEnd, CanceledAt, Created |
| customer.subscription.deleted           | SPID, AccountType, AccountExternalID       | SubscriptionID, CustomerID, Status, CurrentPeriodEnd, CancelAtPeriodEnd, CanceledAt, Created |
| customer.subscription.trial_will_end    | SPID, AccountType, AccountExternalID       | SubscriptionID, CustomerID, Status, CurrentPeriodEnd, CancelAtPeriodEnd, CanceledAt, Created |
| customer.subscription.paused            | SPID, AccountType, AccountExternalID       | SubscriptionID, CustomerID, Status, CurrentPeriodEnd, CancelAtPeriodEnd, CanceledAt, Created |
| customer.subscription.resumed           | SPID, AccountType, AccountExternalID       | SubscriptionID, CustomerID, Status, CurrentPeriodEnd, CancelAtPeriodEnd, CanceledAt, Created |
| invoice.payment_succeeded               | SPID, AccountType, AccountExternalID       | InvoiceID, CustomerID, SubscriptionID, Amount, Status, Created, InvoiceLines |
| invoice.payment_failed                  | SPID, AccountType, AccountExternalID       | InvoiceID, CustomerID, SubscriptionID, Amount, Status, Created, InvoiceLines |
| invoice.created                         | SPID, AccountType, AccountExternalID       | InvoiceID, CustomerID, SubscriptionID, Amount, Status, Created, InvoiceLines |
| invoice.upcoming                        | SPID, AccountType, AccountExternalID       | InvoiceID, CustomerID, SubscriptionID, Amount, Status, Created, InvoiceLines |

### Example: Instantiating and Using a Callback Handler

```go
// For v82 (similar for other versions):
import (
    v82 "github.com/iqhive/gomultistripe/v82"
    "os"
)

func main() {
    // Set your Stripe webhook secret in the environment
    os.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_...")
    handler := v82.NewCallbackHandlerV82()

    // In your HTTP handler for Stripe webhooks:
    func(w http.ResponseWriter, r *http.Request) {
        payload, _ := io.ReadAll(r.Body)
        sigHeader := r.Header.Get("Stripe-Signature")
        err := handler.HandleWebhook(payload, sigHeader)
        if err != nil {
            w.WriteHeader(400)
            return
        }
        w.WriteHeader(200)
    }

    // In a goroutine, process events:
    go func() {
        for evt := range handler.Events() {
            switch evt.Type {
            case "setup_intent.succeeded":
                // Use evt.SPID, evt.AccountType, evt.AccountExternalID, etc.
            case "payment_intent.succeeded":
                // Use evt.PaymentIntentID, evt.Amount, evt.Status, etc.
            // ... handle other event types ...
            }
        }
    }()
}
```

### Notes
- Each versioned handler (e.g., v82, v81, v80, etc.) provides its own `NewCallbackHandlerVXX()` constructor.
- The handler verifies the Stripe webhook signature using the `STRIPE_WEBHOOK_SECRET` environment variable.
- Only events with all required metadata fields are sent to the channel; others are ignored.
- The channel is buffered (size 100) to avoid blocking the webhook handler.
- You are responsible for draining the channel and processing events in your application logic.
- The event struct is version-agnostic and safe to use across all supported versions.

## Adding a New Stripe API Version

To add support for a new Stripe API version (e.g., v83):

1. **Copy an Existing Handler:**
   - Duplicate the most recent handler file (e.g., `handler_v82.go` → `handler_v83.go`).

2. **Update Imports:**
   - Change all Stripe SDK imports to the new version (e.g., `github.com/stripe/stripe-go/v83`).
   - Add the new Stripe SDK version to your `go.mod` using `go get github.com/stripe/stripe-go/v83`.

3. **Rename the Handler Type:**
   - Update the handler struct and registration to match the new version (e.g., `HandlerV83`