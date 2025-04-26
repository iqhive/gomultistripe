// Package v74 provides versioned Stripe API handlers. See handler.go for the interface and registration logic.
package v74

import (
	"context"
	"errors"
	"time"

	gomultistripe "github.com/iqhive/gomultistripe"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/paymentmethod"
	"github.com/stripe/stripe-go/v74/subscription"
)

// Handler implements the Handler interface for Stripe API v74.
type HandlerV74 struct {
}

func NewHandler() *HandlerV74 { return &HandlerV74{} }

func (h *HandlerV74) Version() string { return "v74" }

func (h *HandlerV74) SetSecretKey(secretKey string) {
	stripe.Key = secretKey
}

// CreateCustomer implements the Handler interface for v74.
func (h *HandlerV74) CreateCustomer(ctx context.Context, params *gomultistripe.Customer) (*gomultistripe.Customer, error) {
	stripeParams := &stripe.CustomerParams{
		Name:  stripe.String(params.Name),
		Email: stripe.String(params.Email),
		Phone: stripe.String(params.Phone),
		Address: &stripe.AddressParams{
			PostalCode: stripe.String(params.Postcode),
		},
	}
	cust, err := customer.New(stripeParams)
	if err != nil {
		return nil, err
	}
	return &gomultistripe.Customer{
		ID:    cust.ID,
		Name:  cust.Name,
		Email: cust.Email,
		Phone: cust.Phone,
		Metadata: func() map[string]string {
			if cust.Metadata != nil {
				return cust.Metadata
			} else {
				return make(map[string]string)
			}
		}(),
		Postcode: func() string {
			if cust.Address != nil {
				return cust.Address.PostalCode
			} else {
				return ""
			}
		}(),
		CreatedAt: time.Unix(cust.Created, 0),
	}, nil
}

// UpdateCustomer implements the Handler interface for v74.
func (h *HandlerV74) UpdateCustomer(ctx context.Context, customerID string, params *gomultistripe.Customer) (*gomultistripe.Customer, error) {
	stripeParams := &stripe.CustomerParams{
		Name:  stripe.String(params.Name),
		Email: stripe.String(params.Email),
		Phone: stripe.String(params.Phone),
		Address: &stripe.AddressParams{
			PostalCode: stripe.String(params.Postcode),
		},
	}
	cust, err := customer.Update(customerID, stripeParams)
	if err != nil {
		return nil, err
	}
	return &gomultistripe.Customer{
		ID:    cust.ID,
		Name:  cust.Name,
		Email: cust.Email,
		Phone: cust.Phone,
		Metadata: func() map[string]string {
			if cust.Metadata != nil {
				return cust.Metadata
			} else {
				return make(map[string]string)
			}
		}(),
		Postcode: func() string {
			if cust.Address != nil {
				return cust.Address.PostalCode
			} else {
				return ""
			}
		}(),
		CreatedAt: time.Unix(cust.Created, 0),
	}, nil
}

// GetPaymentMethods implements the Handler interface for v74.
func (h *HandlerV74) GetPaymentMethods(ctx context.Context, customerID string) ([]*gomultistripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}
	iter := paymentmethod.List(params)
	var methods []*gomultistripe.PaymentMethod
	for iter.Next() {
		pm := iter.PaymentMethod()
		methods = append(methods, &gomultistripe.PaymentMethod{
			ID:         pm.ID,
			CustomerID: pm.Customer.ID,
			Metadata: func() map[string]string {
				if pm.Metadata != nil {
					return pm.Metadata
				} else {
					return make(map[string]string)
				}
			}(),
			Type:      string(pm.Type),
			Last4:     pm.Card.Last4,
			Brand:     string(pm.Card.Brand),
			ExpMonth:  uint(pm.Card.ExpMonth),
			ExpYear:   uint(pm.Card.ExpYear),
			CreatedAt: time.Unix(pm.Created, 0),
		})
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return methods, nil
}

// AttachPaymentMethod attaches a payment method to a customer.
func (h *HandlerV74) AttachPaymentMethod(ctx context.Context, customerID string, paymentMethodID string) (*gomultistripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}
	pm, err := paymentmethod.Attach(paymentMethodID, params)
	if err != nil {
		return nil, err
	}
	return &gomultistripe.PaymentMethod{
		ID:         pm.ID,
		CustomerID: pm.Customer.ID,
		Metadata: func() map[string]string {
			if pm.Metadata != nil {
				return pm.Metadata
			} else {
				return make(map[string]string)
			}
		}(),
		Type:      string(pm.Type),
		Last4:     pm.Card.Last4,
		Brand:     string(pm.Card.Brand),
		ExpMonth:  uint(pm.Card.ExpMonth),
		ExpYear:   uint(pm.Card.ExpYear),
		CreatedAt: time.Unix(pm.Created, 0)}, nil
}

// DetachPaymentMethod detaches a payment method from a customer.
func (h *HandlerV74) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	return err
}

// CreatePaymentIntent creates a PaymentIntent for secure payment confirmation.
func (h *HandlerV74) CreatePaymentIntent(ctx context.Context, params *gomultistripe.PaymentIntent) (*gomultistripe.PaymentIntent, error) {
	stripeParams := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(params.Amount),
		Currency:      stripe.String(params.Currency),
		Customer:      stripe.String(params.CustomerID),
		PaymentMethod: stripe.String(params.PaymentMethod),
		Confirm:       stripe.Bool(true),
	}
	pi, err := paymentintent.New(stripeParams)
	if err != nil {
		return nil, err
	}
	return &gomultistripe.PaymentIntent{
		ID:           pi.ID,
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
		Status:       string(pi.Status),
		ClientSecret: pi.ClientSecret,
		CustomerID:   pi.Customer.ID,
		CreatedAt:    time.Unix(pi.Created, 0),
		Metadata: func() map[string]string {
			if pi.Metadata != nil {
				return pi.Metadata
			} else {
				return make(map[string]string)
			}
		}(),
		PaymentMethod: func() string {
			if pi.PaymentMethod != nil {
				return pi.PaymentMethod.ID
			} else {
				return ""
			}
		}(),
	}, nil
}

// RetrievePaymentIntent retrieves a PaymentIntent by ID.
func (h *HandlerV74) RetrievePaymentIntent(ctx context.Context, paymentIntentID string) (*gomultistripe.PaymentIntent, error) {
	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}
	return &gomultistripe.PaymentIntent{
		ID:           pi.ID,
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
		Status:       string(pi.Status),
		ClientSecret: pi.ClientSecret,
		CustomerID:   pi.Customer.ID,
		CreatedAt:    time.Unix(pi.Created, 0),
		Metadata: func() map[string]string {
			if pi.Metadata != nil {
				return pi.Metadata
			} else {
				return make(map[string]string)
			}
		}(),
		PaymentMethod: func() string {
			if pi.PaymentMethod != nil {
				return pi.PaymentMethod.ID
			} else {
				return ""
			}
		}(),
	}, nil
}

// CreateSubscription implements the Handler interface for v74.
func (h *HandlerV74) CreateSubscription(ctx context.Context, customerID string, priceID string) (*gomultistripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{Price: stripe.String(priceID)},
		},
	}
	s, err := subscription.New(params)
	if err != nil {
		return nil, err
	}
	return &gomultistripe.Subscription{
		ID:         s.ID,
		CustomerID: s.Customer.ID,
		Status:     string(s.Status),
		PriceID: func() string {
			if len(s.Items.Data) > 0 && s.Items.Data[0].Price != nil {
				return s.Items.Data[0].Price.ID
			}
			return ""
		}(),
		CurrentPeriodEnd:  s.CancelAt,
		CancelAtPeriodEnd: s.CancelAtPeriodEnd,
		CanceledAt:        s.CanceledAt,
		CreatedAt:         time.Unix(s.Created, 0),
		Metadata: func() map[string]string {
			if s.Metadata != nil {
				return s.Metadata
			} else {
				return make(map[string]string)
			}
		}(),
	}, nil
}

// ListSubscriptions implements the Handler interface for v74.
func (h *HandlerV74) ListSubscriptions(ctx context.Context, customerID string) ([]*gomultistripe.Subscription, error) {
	params := &stripe.SubscriptionListParams{Customer: stripe.String(customerID)}
	iter := subscription.List(params)
	var subs []*gomultistripe.Subscription
	for iter.Next() {
		s := iter.Subscription()
		subs = append(subs, &gomultistripe.Subscription{
			ID:         s.ID,
			CustomerID: s.Customer.ID,
			Status:     string(s.Status),
			PriceID: func() string {
				if len(s.Items.Data) > 0 && s.Items.Data[0].Price != nil {
					return s.Items.Data[0].Price.ID
				}
				return ""
			}(),
			CurrentPeriodEnd:  s.CancelAt,
			CancelAtPeriodEnd: s.CancelAtPeriodEnd,
			CanceledAt:        s.CanceledAt,
			CreatedAt:         time.Unix(s.Created, 0),
			Metadata: func() map[string]string {
				if s.Metadata != nil {
					return s.Metadata
				} else {
					return make(map[string]string)
				}
			}(),
		})
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}

// UpdateSubscription implements the Handler interface for v74.
func (h *HandlerV74) UpdateSubscription(ctx context.Context, subscriptionID string, cancelAtPeriodEnd bool, newPriceID string) (*gomultistripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(cancelAtPeriodEnd),
	}
	if newPriceID != "" {
		params.Items = []*stripe.SubscriptionItemsParams{{
			Price: stripe.String(newPriceID),
		}}
	}
	s, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, err
	}
	return &gomultistripe.Subscription{
		ID:         s.ID,
		CustomerID: s.Customer.ID,
		Status:     string(s.Status),
		PriceID: func() string {
			if len(s.Items.Data) > 0 && s.Items.Data[0].Price != nil {
				return s.Items.Data[0].Price.ID
			}
			return ""
		}(),
		CurrentPeriodEnd:  s.CancelAt,
		CancelAtPeriodEnd: s.CancelAtPeriodEnd,
		CanceledAt:        s.CanceledAt,
		CreatedAt:         time.Unix(s.Created, 0),
		Metadata: func() map[string]string {
			if s.Metadata != nil {
				return s.Metadata
			} else {
				return make(map[string]string)
			}
		}(),
	}, nil
}

// CancelSubscription implements the Handler interface for v74.
func (h *HandlerV74) CancelSubscription(ctx context.Context, subscriptionID string, atPeriodEnd bool) (*gomultistripe.Subscription, error) {
	params := &stripe.SubscriptionCancelParams{
		InvoiceNow: stripe.Bool(!atPeriodEnd),
		Prorate:    stripe.Bool(!atPeriodEnd),
	}
	s, err := subscription.Cancel(subscriptionID, params)
	if err != nil {
		return nil, err
	}
	return &gomultistripe.Subscription{
		ID:         s.ID,
		CustomerID: s.Customer.ID,
		Status:     string(s.Status),
		PriceID: func() string {
			if len(s.Items.Data) > 0 && s.Items.Data[0].Price != nil {
				return s.Items.Data[0].Price.ID
			}
			return ""
		}(),
		CurrentPeriodEnd:  s.CancelAt,
		CancelAtPeriodEnd: s.CancelAtPeriodEnd,
		CanceledAt:        s.CanceledAt,
		CreatedAt:         time.Unix(s.Created, 0),
		Metadata: func() map[string]string {
			if s.Metadata != nil {
				return s.Metadata
			} else {
				return make(map[string]string)
			}
		}(),
	}, nil
}

// ErrInvalidParams is returned when params are not of the expected type.
var ErrInvalidParams = errors.New("invalid params type for this handler version")

func init() {
	gomultistripe.RegisterHandler(NewHandler())
}
