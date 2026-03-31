package platform

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// PlatformConfig represents the minimal structure of platform.yaml
// We only care about the version field
type PlatformConfig struct {
	Version string `yaml:"version"`
}

// ExtractVersion attempts to read and parse platform.yaml to extract the version field.
// It implements graceful degradation: any error results in returning "latest" as the default.
//
// Search priority:
// 1. Check if any argument in args looks like a YAML file path
// 2. Check ./platform.yaml in current directory
// 3. Default to "latest"
//
// Returns:
// - version string (e.g., "v1.0.0") or "latest" on any error
// - error (always nil for now, kept for future enhancements)
func ExtractVersion(args []string) (string, error) {
	yamlPath := LocatePlatformYaml(args)

	// Check if file exists
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		// File doesn't exist, default to latest
		return "latest", nil
	}

	// Read file
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		// Can't read file, default to latest
		return "latest", nil
	}

	// Parse YAML (only version field)
	var config PlatformConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		// Parse error, default to latest
		return "latest", nil
	}

	// Validate version is not empty
	if config.Version == "" {
		return "latest", nil
	}

	return config.Version, nil
}

// LocatePlatformYaml searches for platform.yaml file path in the following order:
// 1. Check if any argument ends with .yaml or .yml (explicit file path)
// 2. Default to ./platform.yaml
func LocatePlatformYaml(args []string) string {
	// Check if any argument looks like a YAML file
	for _, arg := range args {
		if strings.HasSuffix(arg, ".yaml") || strings.HasSuffix(arg, ".yml") {
			return arg
		}
	}

	// Default location
	return "./platform.yaml"
}
