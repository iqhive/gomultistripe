.PHONY: build run dry-run clean

# Build the tool
build:
	go build -o update_stripe_versions main.go

# Run the tool
run: build
	./update_stripe_versions

# Run the tool in dry-run mode (no changes made)
dry-run: build
	./update_stripe_versions --dry-run

# Clean up
clean:
	rm -f update_stripe_versions 