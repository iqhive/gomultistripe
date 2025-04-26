package stripe

import (
	"encoding/json"
	"os"
	"sync"

	gomultistripe "github.com/iqhive/gomultistripe"
	stripe "github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/webhook"
)

type CallbackHandlerV79 struct {
	events chan gomultistripe.CallbackEvent
	mu     sync.Once
}

func NewCallbackHandlerV79() *CallbackHandlerV79 {
	return &CallbackHandlerV79{
		events: make(chan gomultistripe.CallbackEvent, 100),
	}
}

func (h *CallbackHandlerV79) Events() <-chan gomultistripe.CallbackEvent {
	return h.events
}

func (h *CallbackHandlerV79) HandleWebhook(payload []byte, sigHeader string) error {
	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	event, err := webhook.ConstructEvent(payload, sigHeader, secret)
	if err != nil {
		return err
	}

	switch event.Type {
	case stripe.EventTypeSetupIntentSucceeded:
		var intent stripe.SetupIntent
		if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
			return err
		}
		pm := intent.PaymentMethod
		var pmID, brand, last4 string
		var expMonth, expYear uint
		if pm != nil && pm.Card != nil {
			pmID = pm.ID
			brand = string(pm.Card.Brand)
			last4 = pm.Card.Last4
			expMonth = uint(pm.Card.ExpMonth)
			expYear = uint(pm.Card.ExpYear)
		}
		cbEvent := gomultistripe.CallbackEvent{
			Type:            gomultistripe.EventSetupIntentSucceeded,
			Metadata:        make(map[string]string),
			SetupIntentID:   intent.ID,
			PaymentMethodID: pmID,
			CardBrand:       brand,
			CardExpMonth:    expMonth,
			CardExpYear:     expYear,
			CardLast4:       last4,
		}
		for k, v := range intent.Metadata {
			cbEvent.Metadata[k] = v
		}
		h.events <- cbEvent
	case stripe.EventTypePaymentIntentCanceled,
		stripe.EventTypePaymentIntentPaymentFailed,
		stripe.EventTypePaymentIntentSucceeded,
		stripe.EventTypePaymentIntentAmountCapturableUpdated:
		var intent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
			return err
		}
		preAllocated := intent.Metadata["PreAllocated"]
		validateOnly := intent.Metadata["ValidateOnly"]
		pmID := ""
		if intent.PaymentMethod != nil {
			pmID = intent.PaymentMethod.ID
		}
		evt := gomultistripe.CallbackEvent{
			Type:            gomultistripe.CallbackEventType(event.Type),
			Metadata:        make(map[string]string),
			PreAllocated:    preAllocated,
			ValidateOnly:    validateOnly,
			PaymentIntentID: intent.ID,
			Amount:          intent.Amount,
			Status:          string(intent.Status),
			PaymentMethodID: pmID,
		}
		for k, v := range intent.Metadata {
			evt.Metadata[k] = v
		}
		if event.Type == stripe.EventType(gomultistripe.EventPaymentIntentAmountCapturableUpdated) {
			evt.AmountCapturable = intent.AmountCapturable
		}
		if event.Type == stripe.EventType(gomultistripe.EventPaymentIntentPaymentFailed) {
			if intent.LastPaymentError != nil {
				evt.LastPaymentErrorCode = string(intent.LastPaymentError.Code)
				evt.LastPaymentErrorMsg = ""
				if intent.LastPaymentError.Err != nil {
					evt.LastPaymentErrorMsg = intent.LastPaymentError.Err.Error()
				}
				evt.LastPaymentErrorDeclineCode = string(intent.LastPaymentError.DeclineCode)
				if intent.LastPaymentError.PaymentMethod != nil {
					evt.LastPaymentErrorPaymentMethodID = intent.LastPaymentError.PaymentMethod.ID
				}
				evt.LastPaymentErrorChargeID = intent.LastPaymentError.ChargeID
			}
		}
		h.events <- evt
	case stripe.EventTypeCustomerSubscriptionCreated,
		stripe.EventTypeCustomerSubscriptionUpdated,
		stripe.EventTypeCustomerSubscriptionDeleted,
		stripe.EventTypeCustomerSubscriptionTrialWillEnd,
		stripe.EventTypeCustomerSubscriptionPaused,
		stripe.EventTypeCustomerSubscriptionResumed:
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return err
		}
		cbEvent := gomultistripe.CallbackEvent{
			Type:              gomultistripe.CallbackEventType(event.Type),
			Metadata:          make(map[string]string),
			SubscriptionID:    sub.ID,
			CustomerID:        sub.Customer.ID,
			Status:            string(sub.Status),
			CurrentPeriodEnd:  sub.CancelAt,
			CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
			CanceledAt:        sub.CanceledAt,
			Created:           sub.Created,
		}
		for k, v := range sub.Metadata {
			cbEvent.Metadata[k] = v
		}
		h.events <- cbEvent
	case stripe.EventTypeInvoicePaymentSucceeded,
		stripe.EventTypeInvoicePaymentFailed,
		stripe.EventTypeInvoiceCreated,
		stripe.EventTypeInvoiceUpcoming:
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			return err
		}

		cbEvent := gomultistripe.CallbackEvent{
			Type:       gomultistripe.CallbackEventType(event.Type),
			Metadata:   make(map[string]string),
			InvoiceID:  inv.ID,
			CustomerID: inv.Customer.ID,
			Amount:     inv.AmountDue,
			Status:     string(inv.Status),
			Created:    inv.Created,
		}
		for k, v := range inv.Metadata {
			cbEvent.Metadata[k] = v
		}
		if inv.Lines != nil {
			for _, line := range inv.Lines.Data {
				gmline := gomultistripe.InvoiceLine{
					ID:          line.ID,
					Amount:      line.Amount,
					Currency:    string(line.Currency),
					Description: line.Description,
				}
				if line.Subscription != nil {
					gmline.SubscriptionID = line.Subscription.ID
				}
				cbEvent.InvoiceLines = append(cbEvent.InvoiceLines, gmline)
			}
		}
		h.events <- cbEvent
	}
	return nil
}
