//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"reflect"

	"github.com/stripe/stripe-go/v82"
)

func main() {
	t := reflect.TypeOf(stripe.EventTypeRefundCreated)
	fmt.Println("Type:", t)
	fmt.Println("Refund created:", stripe.EventTypeRefundCreated)
	fmt.Println("Refund updated:", stripe.EventTypeRefundUpdated)
	fmt.Println("Refund failed:", stripe.EventTypeRefundFailed)
	fmt.Println("Charge refunded:", stripe.EventTypeChargeRefunded)
}
