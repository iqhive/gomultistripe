package main

import (
	"fmt"
	"os"
)

func main() {
	loadConfig()

	if config.DryRun() {
		fmt.Println("Running in DRY RUN mode - no changes will be made")
	}

	fmt.Println("Checking setup...")
	if err := checkSetup(); err != nil {
		fmt.Printf("Setup check failed: %v\n", err)
		fmt.Println("Please ensure you have internet access and permission to write to the current directory.")
		os.Exit(1)
	}

	fmt.Println("Updating Stripe Go SDK versions...")

	err := UpdateStripeVersions(config.Debug(), config.DryRun())
	if err != nil {
		fmt.Printf("Error updating Stripe versions: %v\n", err)
		os.Exit(1)
	}

	if config.DryRun() {
		fmt.Println("Dry run completed. No changes were made.")
	} else {
		fmt.Println("Stripe Go SDK versions successfully updated!")
	}
}
