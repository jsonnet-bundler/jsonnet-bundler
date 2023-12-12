package pkg

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Version describes a specific version of a package
type Version struct {
	Version string `json:"version"`
	Source  string `json:"source"`
}

type Versions []Version

// Package describes a package referenced in a registry
type Package struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Source      string   `json:"source"`
	Versions    Versions `json:"versions"`
}

// GitRegistry implements the Registry interface and supports registries in git
type GitRegistry struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Source      string `yaml:"source"`
	PackageFile string `yaml:"filename"`
}

// NewGitRegistry creates an instance of GitRegistry
func NewGitRegistry(name, description, source, packageFile string) *GitRegistry {
	return &GitRegistry{
		Name:        name,
		Description: description,
		Source:      source,
		PackageFile: packageFile,
	}
}

func (v Versions) Len() int {
	return len(v)
}

func (v Versions) Less(i, j int) bool {
	// Split version strings into parts based on dots.
	partsA := strings.Split(v[i].Version, ".")
	partsB := strings.Split(v[j].Version, ".")

	// Compare each part of the version strings.
	for k := 0; k < len(partsA) && k < len(partsB); k++ {

		// Convert parts to integers for numerical comparison.
		numA, errA := strconv.Atoi(partsA[k])
		numB, errB := strconv.Atoi(partsB[k])

		// If conversion fails, fall back to lexicographical comparison.
		if errA == nil && errB == nil {
			if numA < numB {
				return true
			} else if numA > numB {
				return false
			}
		} else {
			if partsA[k] < partsB[k] {
				return true
			} else if partsA[k] > partsB[k] {
				return false
			}
		}
	}

	// If all common parts are equal, the shorter version string should come first.
	return len(partsA) < len(partsB)
}

// Swap swaps the elements with indexes i and j.
func (v Versions) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (r GitRegistry) normalizeName() string {
	cleanName := filepath.Clean(r.Name)
	return strings.ToLower(cleanName)
}

func (r GitRegistry) getCacheDir(create bool) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cacheDir := filepath.Join(homeDir, ".cache", "jsonnet-bundler", "registries", r.normalizeName())
	if create {
		err = os.MkdirAll(cacheDir, os.ModePerm)
		if err != nil {
			return "", err
		}

	}

	return cacheDir, nil
}

func (r GitRegistry) CleanCache() error {
	cacheDir, err := r.getCacheDir(false)
	if err != nil {
		return err
	}
	return os.RemoveAll(cacheDir)
}

func (r GitRegistry) run(ctx context.Context, args ...string) (string, error) {
	cacheDir, err := r.getCacheDir(true)
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = cacheDir
	cmd.Stderr = nil
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

func (r GitRegistry) Init(ctx context.Context) error {
	cacheDir, err := r.getCacheDir(true)
	if err != nil {
		return err
	}
	_, err = os.Stat(filepath.Join(cacheDir, ".git"))
	if err == nil {
		// already initialized
		return nil
	} else if !os.IsNotExist(err) {
		// some fs error
		return err
	}
	// file not found, start init process
	_, err = r.run(ctx, "init")
	if err != nil {
		return err
	}
	_, err = r.run(ctx, "remote", "add", "origin", r.Source)
	if err != nil {
		return err
	}
	_, err = r.run(ctx, "sparse-checkout", "set", "--skip-checks", r.PackageFile)
	if err != nil {
		return err
	}
	_, err = r.run(ctx, "fetch", "--filter=blob:none", "--depth=1")
	if err != nil {
		return err
	}

	output, err := r.run(ctx, "symbolic-ref", "HEAD")
	if err != nil {
		return err
	}

	parts := strings.Split(output, "/")
	branch := parts[len(parts)-1]

	_, err = r.run(ctx, "checkout", branch)
	if err != nil {
		return err
	}

	return nil
}

// Update to latest version
func (r GitRegistry) Update(ctx context.Context) error {
	_, err := r.run(ctx, "pull")
	if err != nil {
		return err
	}
	return nil
}

// Search queries the data for some text
func (r GitRegistry) Search(ctx context.Context, query string) ([]Package, error) {
	cacheDir, err := r.getCacheDir(true)
	if err != nil {
		return nil, err
	}
	fileContent, err := os.ReadFile(filepath.Join(cacheDir, r.PackageFile))
	if err != nil {
		return nil, err
	}

	// Create an instance of the struct to hold the data
	var entries []Package

	// Unmarshal the JSON data into the struct
	err = json.Unmarshal(fileContent, &entries)
	if err != nil {
		return nil, err
	}

	filtered := []Package{}

	q := strings.ToLower(query)
	for _, item := range entries {
		if strings.Contains(strings.ToLower(item.Name), q) || strings.Contains(strings.ToLower(item.Description), q) {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

func UpdateRegistries(ctx context.Context) error {
	for _, registry := range Registries.Entries {
		err := registry.Init(ctx)
		if err != nil {
			return err
		}
		err = registry.Update(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func SearchPackage(ctx context.Context, query string) (map[string][]Package, error) {
	results := make(map[string][]Package)

	for _, registry := range Registries.Entries {
		err := registry.Init(ctx)
		if err != nil {
			return nil, err
		}
		result, err := registry.Search(ctx, query)
		if err != nil {
			return nil, err
		}
		results[registry.Name] = result
	}

	return results, nil
}
