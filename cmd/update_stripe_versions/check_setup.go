package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// checkSetup verifies that we can access the GitHub API to fetch Stripe tags
func checkSetup() error {
	// Check if we can access the GitHub API
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	url := "https://api.github.com/repos/stripe/stripe-go/tags?per_page=1"
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to access GitHub API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned non-OK status: %s", resp.Status)
	}

	// Check write permissions in the current directory
	testFile := "test_write_permission.tmp"
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		return fmt.Errorf("failed to write to current directory: %v", err)
	}

	// Clean up test file
	os.Remove(testFile)

	return nil
}
