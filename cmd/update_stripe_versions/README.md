# Stripe Go SDK Version Updater

This tool automatically updates your local Go Stripe SDK versions to the latest available versions.

## Features

- Fetches all available versions from the [stripe-go repository](https://github.com/stripe/stripe-go)
- Updates the 5 most recent existing versions to their latest minor and patch releases (only if newer)
- Automatically adds new major versions by copying the most recent version's files and updating imports
- Runs tests automatically to verify changes work correctly
- Supports dry-run mode to preview changes without modifying files

## Usage

```bash
# Build the tool
go build

# Run the tool normally (makes actual changes)
./update_stripe_versions

# Run in dry-run mode (no changes will be made)
./update_stripe_versions --dry-run
```

Or use the provided Makefile:

```bash
# Build and run (with changes)
make run

# Build and run in dry-run mode
make dry-run
```

## What It Does

The tool will:
1. Fetch all tags from the Stripe Go SDK repository
2. Find existing Stripe versions in your local project
3. Update the 5 most recent versions to the latest minor/patch releases (only if newer than current)
4. Add any new major versions that don't exist yet
5. Run tests to verify everything still works

## How It Works

The tool uses the GitHub API to fetch tags from the Stripe Go repository, then:

1. For existing versions: Updates go.mod to use the latest minor/patch versions, but only if they're newer than what's already in go.mod
2. For new major versions: Creates new directories, copies files from the latest version, updates import paths

After running the tool, you should see updates in your go.mod file and possibly new version directories.

## Best Practices

- Always run with `--dry-run` first to see what changes will be made
- The tool automatically runs tests, but you may want to manually verify your application works with the updated versions 