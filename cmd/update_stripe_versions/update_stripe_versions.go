package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Version represents a semantic version of the Stripe Go SDK
type Version struct {
	Major int
	Minor int
	Patch int
	Tag   string
}

// UpdateStripeVersions updates existing versions and adds new ones
// If dryRun is true, it will only print planned actions without making changes
func UpdateStripeVersions(debug bool, dryRun bool) error {
	// Find versions we already have
	existingVersions, baseDir, err := findExistingVersions()
	if err != nil {
		return fmt.Errorf("error finding existing versions: %v", err)
	}

	// Get tags from GitHub API
	tags, err := getStripeGoTags()
	if err != nil {
		return fmt.Errorf("error fetching tags: %v", err)
	}

	// Update 5 most recent existing versions
	if err := updateExistingVersions(baseDir, existingVersions, tags, dryRun); err != nil {
		return fmt.Errorf("error updating existing versions: %v", err)
	}

	// Add new major versions (>80)
	if err := addNewMajorVersions(baseDir, existingVersions, tags, dryRun); err != nil {
		return fmt.Errorf("error adding new major versions: %v", err)
	}

	// If not in dry-run mode, run tests and commit changes
	if !dryRun {
		fmt.Println("Running tests to verify changes...")
		if err := runTests(baseDir); err != nil {
			return fmt.Errorf("tests failed after updating versions: %v", err)
		}
		fmt.Println("Tests passed!")

		// Commit changes to git
		if err := commitChanges(baseDir); err != nil {
			return fmt.Errorf("failed to commit changes: %v", err)
		}
		fmt.Println("Changes committed to git!")
	}

	return nil
}

// runTests runs go test for the entire project
func runTests(baseDir string) error {
	cmd := exec.Command("go", "test", "./...")
	cmd.Stdout = os.Stdout
	cmd.Dir = baseDir
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// commitChanges commits the changes to git
func commitChanges(baseDir string) error {
	// Check if we're in a git repository
	if _, err := os.Stat(baseDir + "/.git"); os.IsNotExist(err) {
		fmt.Println("Not a git repository, skipping commit")
		return nil
	}

	// Get current date for commit message
	currentDate := time.Now().Format("2006-01-02")
	commitMsg := fmt.Sprintf("Update Stripe SDK versions (%s)", currentDate)

	// Add changes
	cmd := exec.Command("git", "add", "go.mod", "go.sum")
	cmd.Dir = baseDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %v", err)
	}

	// Add any new version directories
	cmd = exec.Command("git", "add", "v*")
	cmd.Dir = baseDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// It's okay if this fails (e.g., no new version directories)
	_ = cmd.Run()

	// Commit changes
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = baseDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %v", err)
	}

	return nil
}

func getStripeGoTags() ([]Version, error) {
	// Get tags from GitHub API
	url := "https://api.github.com/repos/stripe/stripe-go/tags?per_page=100"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var tagList []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &tagList); err != nil {
		return nil, err
	}

	// Regular expression for version tags
	re := regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

	// Extract versions
	var versions []Version
	for _, tag := range tagList {
		matches := re.FindStringSubmatch(tag.Name)
		if matches != nil {
			major, _ := strconv.Atoi(matches[1])
			minor, _ := strconv.Atoi(matches[2])
			patch, _ := strconv.Atoi(matches[3])
			versions = append(versions, Version{
				Major: major,
				Minor: minor,
				Patch: patch,
				Tag:   tag.Name,
			})
		}
	}

	// Sort versions by major, minor, patch (descending)
	sort.Slice(versions, func(i, j int) bool {
		if versions[i].Major != versions[j].Major {
			return versions[i].Major > versions[j].Major
		}
		if versions[i].Minor != versions[j].Minor {
			return versions[i].Minor > versions[j].Minor
		}
		return versions[i].Patch > versions[j].Patch
	})

	return versions, nil
}

func findExistingVersions() ([]Version, string, error) {
	baseDir := "."
	if _, err := os.Stat(baseDir + "/go.mod"); err != nil {
		baseDir = "../.."
	}
	if _, err := os.Stat(baseDir + "/go.mod"); err != nil {
		log.Fatalf("error finding existing versions in %s: %v", baseDir, err)
	}

	vers, err := findExistingVersionsInDir(baseDir)
	log.Printf("existing versions in %s: %v", baseDir, vers)
	if err != nil {
		log.Printf("error finding existing versions in %s: %v", ".", err)
		return nil, baseDir, err
	}

	if len(vers) == 0 {
		log.Printf("no existing versions found")
		return nil, baseDir, fmt.Errorf("no existing versions found")
	}

	return vers, baseDir, nil
}

func findExistingVersionsInDir(parentDir string) ([]Version, error) {
	// Find directories named 'v\d+'
	re := regexp.MustCompile(`^v(\d+)$`)

	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return nil, err
	}

	var versions []Version
	for _, entry := range entries {
		if entry.IsDir() {
			matches := re.FindStringSubmatch(entry.Name())
			if matches != nil {
				major, _ := strconv.Atoi(matches[1])
				versions = append(versions, Version{
					Major: major,
					Minor: 0, // We don't know minor/patch from directory names
					Patch: 0,
					Tag:   entry.Name(),
				})
			}
		}
	}

	// Sort by major version (descending)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Major > versions[j].Major
	})

	return versions, nil
}

func updateExistingVersions(baseDir string, existingVersions []Version, allTags []Version, dryRun bool) error {
	// Get latest minor/patch for each major version
	latestVersions := make(map[int]Version)
	for _, v := range allTags {
		if existing, ok := latestVersions[v.Major]; !ok || v.Minor > existing.Minor || (v.Minor == existing.Minor && v.Patch > existing.Patch) {
			latestVersions[v.Major] = v
		}
	}

	// Get current versions from go.mod
	currentVersions, err := getCurrentVersionsFromGoMod(baseDir)
	if err != nil {
		return fmt.Errorf("failed to get current versions from go.mod: %v", err)
	}

	// Limit to 5 most recent versions
	limit := 5
	if len(existingVersions) < limit {
		limit = len(existingVersions)
	}

	// Update go.mod file
	for i := 0; i < limit; i++ {
		v := existingVersions[i]
		if latest, ok := latestVersions[v.Major]; ok {
			// Get current version for this major version
			current, exists := currentVersions[v.Major]

			// Skip if we don't have this version in go.mod
			if !exists {
				if dryRun {
					fmt.Printf("[DRY RUN] Would add v%d.%d.%d (not currently in go.mod)\n",
						latest.Major, latest.Minor, latest.Patch)
				} else {
					fmt.Printf("Adding v%d.%d.%d (not currently in go.mod)\n",
						latest.Major, latest.Minor, latest.Patch)
				}
			} else {
				// Compare versions - only update if the new version is newer
				if latest.Minor < current.Minor ||
					(latest.Minor == current.Minor && latest.Patch <= current.Patch) {
					if config.Debug() {
						fmt.Printf("Skipping v%d: current v%d.%d.%d is already at or newer than v%d.%d.%d\n",
							v.Major, current.Major, current.Minor, current.Patch,
							latest.Major, latest.Minor, latest.Patch)
					}
					continue
				}

				if dryRun {
					fmt.Printf("[DRY RUN] Would update v%d from v%d.%d.%d to v%d.%d.%d\n",
						v.Major, current.Major, current.Minor, current.Patch,
						latest.Major, latest.Minor, latest.Patch)
					continue
				}

				fmt.Printf("Updating v%d from v%d.%d.%d to v%d.%d.%d\n",
					v.Major, current.Major, current.Minor, current.Patch,
					latest.Major, latest.Minor, latest.Patch)
			}

			// Skip actual update in dry-run mode
			if dryRun {
				continue
			}

			// Use go get to update the module
			cmd := exec.Command("go", "get", fmt.Sprintf("github.com/stripe/stripe-go/v%d@%s", v.Major, latest.Tag))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to update v%d: %v", v.Major, err)
			}
		}
	}

	// Update go.mod and go.sum
	if !dryRun {
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = baseDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return nil
}

// getCurrentVersionsFromGoMod parses go.mod to get the current versions
func getCurrentVersionsFromGoMod(baseDir string) (map[int]Version, error) {
	// Read go.mod file
	data, err := os.ReadFile(baseDir + "/go.mod")
	if err != nil {
		// Try parent directory if not found
		parentData, parentErr := os.ReadFile("../../go.mod")
		if parentErr != nil {
			return nil, fmt.Errorf("failed to read go.mod: %v", err)
		}
		data = parentData
	}

	content := string(data)

	// Extract stripe-go version dependencies
	re := regexp.MustCompile(`github\.com/stripe/stripe-go/v(\d+)\s+v(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	versions := make(map[int]Version)
	for _, match := range matches {
		if len(match) != 5 {
			continue
		}

		major, _ := strconv.Atoi(match[1])
		vMajor, _ := strconv.Atoi(match[2]) // This should equal major
		minor, _ := strconv.Atoi(match[3])
		patch, _ := strconv.Atoi(match[4])

		if major != vMajor {
			fmt.Printf("Warning: Inconsistent major versions in go.mod: %d vs %d\n", major, vMajor)
		}

		versions[major] = Version{
			Major: major,
			Minor: minor,
			Patch: patch,
			Tag:   fmt.Sprintf("v%d.%d.%d", major, minor, patch),
		}
	}

	return versions, nil
}

func addNewMajorVersions(baseDir string, existingVersions []Version, allTags []Version, dryRun bool) error {
	// Find the largest existing major version
	var maxExistingMajor int
	for _, v := range existingVersions {
		if v.Major > maxExistingMajor {
			maxExistingMajor = v.Major
		}
	}

	// Get current versions from go.mod
	currentVersions, err := getCurrentVersionsFromGoMod(baseDir)
	if err != nil {
		return fmt.Errorf("failed to get current versions from go.mod: %v", err)
	}

	// Find new major versions not in our existing directories
	var newMajorVersions []Version
	for _, v := range allTags {
		found := false
		for _, existing := range existingVersions {
			if existing.Major == v.Major {
				found = true
				break
			}
		}

		if !found {
			// Only add the latest minor/patch for each new major
			alreadyAdded := false
			for i, added := range newMajorVersions {
				if added.Major == v.Major {
					// Replace if this is a newer minor/patch
					if v.Minor > added.Minor || (v.Minor == added.Minor && v.Patch > added.Patch) {
						newMajorVersions[i] = v
					}
					alreadyAdded = true
					break
				}
			}
			if !alreadyAdded {
				newMajorVersions = append(newMajorVersions, v)
			}
		}
	}

	// Sort new versions by major
	sort.Slice(newMajorVersions, func(i, j int) bool {
		return newMajorVersions[i].Major < newMajorVersions[j].Major
	})

	// Get source directory (highest existing version)
	if len(existingVersions) == 0 {
		return fmt.Errorf("no existing versions found")
	}

	sourceDir := fmt.Sprintf("v%d", maxExistingMajor)

	// Add each new major version
	for _, v := range newMajorVersions {
		destDir := fmt.Sprintf("v%d", v.Major)

		// Check if we already have this version in go.mod but missing directory
		current, hasCurrent := currentVersions[v.Major]

		if dryRun {
			if hasCurrent {
				fmt.Printf("[DRY RUN] Would add new major version directory: %s (v%d.%d.%d) - already in go.mod as v%d.%d.%d\n",
					destDir, v.Major, v.Minor, v.Patch, current.Major, current.Minor, current.Patch)
			} else {
				fmt.Printf("[DRY RUN] Would add new major version: %s (v%d.%d.%d)\n",
					destDir, v.Major, v.Minor, v.Patch)
			}
			continue
		}

		if hasCurrent {
			fmt.Printf("Adding new major version directory: %s (v%d.%d.%d) - already in go.mod as v%d.%d.%d\n",
				destDir, v.Major, v.Minor, v.Patch, current.Major, current.Minor, current.Patch)
		} else {
			fmt.Printf("Adding new major version: %s (v%d.%d.%d)\n",
				baseDir+"/"+destDir, v.Major, v.Minor, v.Patch)
		}

		// Create directory
		if config.Debug() {
			fmt.Printf("Creating directory %s\n", baseDir+"/"+destDir)
		}
		if err := os.MkdirAll(baseDir+"/"+destDir, 0755); err != nil {
			return err
		}

		// Copy files from latest existing version
		if config.Debug() {
			fmt.Printf("Copying files from %s to %s\n", baseDir+"/"+sourceDir, baseDir+"/"+destDir)
		}
		if err := copyDir(baseDir+"/"+sourceDir, baseDir+"/"+destDir); err != nil {
			return err
		}

		// Update imports in new directory
		if config.Debug() {
			fmt.Printf("Updating imports in %s\n", baseDir+"/"+destDir)
		}
		if err := updateImports(baseDir, destDir, maxExistingMajor, v.Major); err != nil {
			return err
		}

		// Add new version to go.mod (if not already there)
		if config.Debug() {
			fmt.Printf("Adding new version to go.mod: %s\n", v.Tag)
		}
		if !hasCurrent {
			cmd := exec.Command("go", "get", fmt.Sprintf("github.com/stripe/stripe-go/v%d@%s", v.Major, v.Tag))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	// Update go.mod and go.sum
	if !dryRun {
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from source
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip if we're at the source directory itself
		if relPath == "." {
			return nil
		}

		// Construct destination path
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// Copy file
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(dstPath, data, info.Mode())
		}
	})
}

func updateImports(baseDir string, dir string, oldMajor, newMajor int) error {
	return filepath.Walk(baseDir+"/"+dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only process Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Read file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		content := string(data)

		// Replace imports
		// oldImport := fmt.Sprintf("github.com/stripe/stripe-go/v%d", oldMajor)
		// newImport := fmt.Sprintf("github.com/stripe/stripe-go/v%d", newMajor)

		oldImport := fmt.Sprintf("v%d", oldMajor)
		newImport := fmt.Sprintf("v%d", newMajor)
		content = strings.ReplaceAll(content, oldImport, newImport)

		oldStruct := fmt.Sprintf("HandlerV%d", oldMajor)
		newStruct := fmt.Sprintf("HandlerV%d", newMajor)
		content = strings.ReplaceAll(content, oldStruct, newStruct)

		// Write updated content
		log.Printf("Updating %s", path)
		return os.WriteFile(path, []byte(content), info.Mode())
	})
}
