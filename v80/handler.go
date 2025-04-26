// Package stripe provides versioned Stripe API handlers. See handler.go for the interface and registration logic.
package stripe

import (
	"context"

	gomultistripe "github.com/iqhive/gomultistripe"
	"github.com/stripe/stripe-go/v80"
	"github.com/stripe/stripe-go/v80/customer"
	"github.com/stripe/stripe-go/v80/paymentintent"
	"github.com/stripe/stripe-go/v80/paymentmethod"
)

// HandlerV80 implements the Handler interface for Stripe API v80.
type HandlerV80 struct{}

func (h *HandlerV80) Version() string { return "v80" }

func (h *HandlerV80) CreateCustomer(ctx context.Context, params *gomultistripe.Customer) (*gomultistripe.Customer, error) {
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
	return &gomultistripe.Customer{ID: cust.ID, Name: cust.Name, Email: cust.Email, Phone: cust.Phone, Postcode: func() string {
		if cust.Address != nil {
			return cust.Address.PostalCode
		} else {
			return ""
		}
	}()}, nil
}

func (h *HandlerV80) UpdateCustomer(ctx context.Context, customerID string, params *gomultistripe.Customer) (*gomultistripe.Customer, error) {
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
	return &gomultistripe.Customer{ID: cust.ID, Name: cust.Name, Email: cust.Email, Phone: cust.Phone, Postcode: func() string {
		if cust.Address != nil {
			return cust.Address.PostalCode
		} else {
			return ""
		}
	}()}, nil
}

func (h *HandlerV80) GetPaymentMethods(ctx context.Context, customerID string) ([]*gomultistripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}
	iter := paymentmethod.List(params)
	var methods []*gomultistripe.PaymentMethod
	for iter.Next() {
		pm := iter.PaymentMethod()
		methods = append(methods, &gomultistripe.PaymentMethod{
			ID:       pm.ID,
			Type:     string(pm.Type),
			Last4:    pm.Card.Last4,
			Brand:    string(pm.Card.Brand),
			ExpMonth: uint(pm.Card.ExpMonth),
			ExpYear:  uint(pm.Card.ExpYear),
		})
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return methods, nil
}

func (h *HandlerV80) AttachPaymentMethod(ctx context.Context, customerID string, paymentMethodID string) (*gomultistripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}
	pm, err := paymentmethod.Attach(paymentMethodID, params)
	if err != nil {
		return nil, err
	}
	return &gomultistripe.PaymentMethod{
		ID:       pm.ID,
		Type:     string(pm.Type),
		Last4:    pm.Card.Last4,
		Brand:    string(pm.Card.Brand),
		ExpMonth: uint(pm.Card.ExpMonth),
		ExpYear:  uint(pm.Card.ExpYear),
	}, nil
}

func (h *HandlerV80) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	return err
}

func (h *HandlerV80) CreatePaymentIntent(ctx context.Context, params *gomultistripe.PaymentIntent) (*gomultistripe.PaymentIntent, error) {
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
		PaymentMethod: func() string {
			if pi.PaymentMethod != nil {
				return pi.PaymentMethod.ID
			} else {
				return ""
			}
		}(),
	}, nil
}

func (h *HandlerV80) RetrievePaymentIntent(ctx context.Context, paymentIntentID string) (*gomultistripe.PaymentIntent, error) {
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
		PaymentMethod: func() string {
			if pi.PaymentMethod != nil {
				return pi.PaymentMethod.ID
			} else {
				return ""
			}
		}(),
	}, nil
}

func init() {
	gomultistripe.RegisterHandler(&HandlerV80{})
}
