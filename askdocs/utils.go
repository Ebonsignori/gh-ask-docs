package askdocs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// stripANSI removes ANSI escape codes from a string.
var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

func StripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

func AutoLink(url, text string) string {
	if url == text {
		return "<" + url + ">"
	}
	return "[" + EscapeMarkdown(text) + "](" + url + ")"
}

func EscapeMarkdown(s string) string {
	return strings.ReplaceAll(s, "]", `\]`)
}

// SupportedVersions represents the structure of the supported versions JSON file
type SupportedVersions struct {
	LastUpdated       string   `json:"lastUpdated"`
	SupportedVersions []string `json:"supportedVersions"`
	LatestVersion     string   `json:"latestVersion"`
}

// LoadSupportedVersions loads the supported enterprise versions from the JSON file
func LoadSupportedVersions() (*SupportedVersions, error) {
	// Get the executable path
	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	// Build path to data directory relative to executable
	dataPath := filepath.Join(filepath.Dir(execPath), "data", "supported-versions.json")

	// If that doesn't exist, try relative to current working directory (for development)
	if _, statErr := os.Stat(dataPath); os.IsNotExist(statErr) {
		dataPath = filepath.Join("data", "supported-versions.json")
	}

	// Read the file
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var versions SupportedVersions
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, err
	}

	return &versions, nil
}

// IsVersionSupported checks if a given enterprise server version is supported
func IsVersionSupported(version string) bool {
	versions, err := LoadSupportedVersions()
	if err != nil {
		// Fallback to hardcoded versions if file loading fails
		hardcodedVersions := []string{"3.11", "3.12", "3.13", "3.14", "3.15", "3.16", "3.17"}
		for _, v := range hardcodedVersions {
			if v == version {
				return true
			}
		}
		return false
	}

	for _, v := range versions.SupportedVersions {
		if v == version {
			return true
		}
	}
	return false
}

func NormalizeVersion(v string) string {
	switch v {
	case "free-pro-team", "enterprise-cloud":
		return v + "@latest"
	}

	// Handle enterprise-server versions
	if strings.HasPrefix(v, "enterprise-server@") {
		// Extract version number
		versionPart := strings.TrimPrefix(v, "enterprise-server@")

		// Allow "latest" through
		if versionPart == "latest" {
			return v
		}

		// Check if the specific version is supported
		if IsVersionSupported(versionPart) {
			return v
		}

		// If version is not supported, fall back to latest supported version
		versions, err := LoadSupportedVersions()
		if err == nil && versions.LatestVersion != "" {
			return "enterprise-server@" + versions.LatestVersion
		}

		// Ultimate fallback
		return "enterprise-server@3.15"
	}

	return "free-pro-team@latest"
}

func IsLight() bool {
	switch os.Getenv("GH_THEME") {
	case "light":
		return true
	case "dark":
		return false
	}
	return runtime.GOOS == "windows"
}

func ExitCouldNotAnswer() {
	fmt.Println("⚠️  The AI could not answer your question.")
	os.Exit(1)
}

func Fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
