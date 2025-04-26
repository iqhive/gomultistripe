package v74

import (
	"context"
	"os"
	"testing"

	"github.com/iqhive/gomultistripe"
	"github.com/stripe/stripe-go/v74"
)

func TestHandlerV74_CreateCustomer(t *testing.T) {
	stripe.Key = os.Getenv("STRIPE_API_KEY")
	if stripe.Key == "" {
		t.Skip("STRIPE_API_KEY not set")
	}

	h := gomultistripe.GetHandler("v74")
	if h == nil {
		t.Fatal("Handler for v74 not registered")
	}

	params := &gomultistripe.Customer{
		Name:  "Test User",
		Email: "testuser@example.com",
	}
	cust, err := h.CreateCustomer(context.Background(), params)
	if err != nil {
		t.Fatalf("CreateCustomer failed: %v", err)
	}
	if cust == nil {
		t.Fatal("Expected customer, got nil")
	}
}
