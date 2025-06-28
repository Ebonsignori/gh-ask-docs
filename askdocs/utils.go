package askdocs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
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
	// Try GH_THEME first (GitHub CLI sets this)
	switch os.Getenv("GH_THEME") {
	case "light":
		return true
	case "dark":
		return false
	}

	// Check COLORFGBG environment variable (set by some terminals)
	// Format is usually "foreground;background" where light background is high numbers
	if colorfgbg := os.Getenv("COLORFGBG"); colorfgbg != "" {
		parts := strings.Split(colorfgbg, ";")
		if len(parts) >= 2 {
			if bg, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
				// Light backgrounds typically have high color numbers (7, 15, etc.)
				return bg >= 7
			}
		}
	}

	// Check for known light terminal programs
	termProgram := os.Getenv("TERM_PROGRAM")
	switch termProgram {
	case "Apple_Terminal":
		// macOS Terminal defaults to light theme
		return true
	case "iTerm.app":
		// iTerm2 - can't reliably detect, assume dark as it's more common
		return false
	case "vscode":
		// VS Code integrated terminal - assume follows editor theme, default dark
		return false
	}

	// Check if we're in a known IDE terminal that might be light
	if os.Getenv("VSCODE_INJECTION") != "" || os.Getenv("TERM_PROGRAM") == "vscode" {
		return false // VS Code defaults to dark
	}

	// Platform-specific defaults
	switch runtime.GOOS {
	case "windows":
		// Windows terminal traditionally light, but newer Windows Terminal is dark
		// Check if we're in newer Windows Terminal
		if os.Getenv("WT_SESSION") != "" {
			return false // Windows Terminal defaults to dark
		}
		return true // Traditional Windows console is light
	case "darwin":
		// macOS Terminal.app defaults to light, but most developers use dark
		return false
	default:
		// Linux and others - most terminals default to dark
		return false
	}
}

func ExitCouldNotAnswer() {
	fmt.Println("⚠️  The AI could not answer your question.")
	os.Exit(1)
}

func Fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
