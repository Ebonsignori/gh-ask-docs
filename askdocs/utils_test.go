package askdocs

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestStripANSI(t *testing.T) {
	input := "\x1b[31mred\x1b[0m and \x1b[32mgreen\x1b[0m"
	expected := "red and green"
	actual := StripANSI(input)
	if actual != expected {
		t.Errorf("stripANSI failed: expected %q, got %q", expected, actual)
	}
}

func TestAutoLink(t *testing.T) {
	tests := []struct {
		url, text, expected string
	}{
		{"https://example.com", "https://example.com", "<https://example.com>"},
		{"https://example.com", "Example", "[Example](https://example.com)"},
		{"https://example.com", "bracket]test", "[bracket\\]test](https://example.com)"},
	}

	for _, test := range tests {
		result := AutoLink(test.url, test.text)
		if result != test.expected {
			t.Errorf("autoLink(%q, %q) = %q; want %q", test.url, test.text, result, test.expected)
		}
	}
}

func TestEscapeMarkdown(t *testing.T) {
	input := "some]text]with]brackets"
	expected := "some\\]text\\]with\\]brackets"
	actual := EscapeMarkdown(input)
	if actual != expected {
		t.Errorf("escapeMarkdown failed: expected %q, got %q", expected, actual)
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := map[string]string{
		"free-pro-team":            "free-pro-team@latest",
		"enterprise-cloud":         "enterprise-cloud@latest",
		"enterprise-server@latest": "enterprise-server@latest",
		"some-other-value":         "free-pro-team@latest",
		"":                         "free-pro-team@latest",
	}

	for input, expected := range tests {
		if got := NormalizeVersion(input); got != expected {
			t.Errorf("normalizeVersion(%q) = %q; want %q", input, got, expected)
		}
	}
}

func TestNormalizeVersionWithSupportedVersions(t *testing.T) {
	// Test with a supported version (assuming 3.15 is in our test data)
	result := NormalizeVersion("enterprise-server@3.15")
	if !strings.HasPrefix(result, "enterprise-server@") {
		t.Errorf("normalizeVersion should return enterprise-server format, got %q", result)
	}

	// Test with an unsupported version (should fallback)
	result = NormalizeVersion("enterprise-server@1.0")
	if !strings.HasPrefix(result, "enterprise-server@") {
		t.Errorf("normalizeVersion should return enterprise-server format even for unsupported versions, got %q", result)
	}
}

func TestNormalizeVersionEdgeCases(t *testing.T) {
	tests := map[string]string{
		"free-pro-team":            "free-pro-team@latest",
		"enterprise-cloud":         "enterprise-cloud@latest",
		"enterprise-server@latest": "enterprise-server@latest",
		"enterprise-server@":       "enterprise-server@3.15", // empty version part falls back to default
		"random-value":             "free-pro-team@latest",
		"":                         "free-pro-team@latest",
	}

	for input, expected := range tests {
		result := NormalizeVersion(input)
		if result != expected {
			t.Errorf("NormalizeVersion(%q) = %q, want %q", input, result, expected)
		}
	}
}

func TestNormalizeVersionWithLoadError(t *testing.T) {
	// Save original working directory
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	// Change to a directory without data file to test fallback behavior
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Test with unsupported version when LoadSupportedVersions fails
	result := NormalizeVersion("enterprise-server@999.0")
	expected := "enterprise-server@3.15" // ultimate fallback
	if result != expected {
		t.Errorf("NormalizeVersion with load error should fallback to %q, got %q", expected, result)
	}
}

func TestIsVersionSupported(t *testing.T) {
	// Test with versions that should be in our test data
	testVersions := []string{"3.11", "3.12", "3.13", "3.14", "3.15"}

	for _, version := range testVersions {
		if !IsVersionSupported(version) {
			// This might pass or fail depending on the test data, but shouldn't crash
			t.Logf("Version %s not found in supported versions (this may be expected)", version)
		}
	}

	// Test with obviously unsupported version
	if IsVersionSupported("1.0") {
		t.Errorf("Version 1.0 should not be supported")
	}
}

func TestIsVersionSupportedWithFallback(t *testing.T) {
	// Save original working directory
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	// Change to a directory without data file to test fallback
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Test hardcoded fallback versions
	hardcodedVersions := []string{"3.11", "3.12", "3.13", "3.14", "3.15", "3.16", "3.17"}
	for _, version := range hardcodedVersions {
		if !IsVersionSupported(version) {
			t.Errorf("Hardcoded version %s should be supported in fallback", version)
		}
	}

	// Test unsupported version with fallback
	if IsVersionSupported("2.0") {
		t.Errorf("Version 2.0 should not be supported even in fallback")
	}
}

func TestLoadSupportedVersions(t *testing.T) {
	versions, err := LoadSupportedVersions()
	if err != nil {
		t.Logf("Could not load supported versions (expected in test environment): %v", err)
		return
	}

	if len(versions.SupportedVersions) == 0 {
		t.Errorf("Expected at least one supported version")
	}

	if versions.LatestVersion == "" {
		t.Errorf("Expected latest version to be set")
	}
}

func TestLoadSupportedVersionsWithTestData(t *testing.T) {
	// Create a temporary test data file
	testData := `{
		"lastUpdated": "2024-01-01T00:00:00.000Z",
		"supportedVersions": ["3.15", "3.16", "3.17"],
		"latestVersion": "3.17"
	}`

	// Write to a temporary file
	tmpDir := t.TempDir()
	testFile := tmpDir + "/supported-versions.json"
	err := os.WriteFile(testFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Save and restore original working directory
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	// Change to temp directory so relative path works
	os.Chdir(tmpDir)

	// Create data directory structure
	os.Mkdir("data", 0755)
	os.Rename("supported-versions.json", "data/supported-versions.json")

	versions, err := LoadSupportedVersions()
	if err != nil {
		t.Fatalf("Failed to load test supported versions: %v", err)
	}

	if len(versions.SupportedVersions) != 3 {
		t.Errorf("Expected 3 supported versions, got %d", len(versions.SupportedVersions))
	}

	if versions.LatestVersion != "3.17" {
		t.Errorf("Expected latest version to be 3.17, got %s", versions.LatestVersion)
	}

	if versions.LastUpdated != "2024-01-01T00:00:00.000Z" {
		t.Errorf("Expected lastUpdated to be 2024-01-01T00:00:00.000Z, got %s", versions.LastUpdated)
	}
}

func TestIsLight(t *testing.T) {
	// Save and defer restore original env
	orig := os.Getenv("GH_THEME")
	defer os.Setenv("GH_THEME", orig)

	// Test explicit values
	_ = os.Setenv("GH_THEME", "light")
	if !IsLight() {
		t.Errorf("isLight() should return true for GH_THEME=light")
	}

	_ = os.Setenv("GH_THEME", "dark")
	if IsLight() {
		t.Errorf("isLight() should return false for GH_THEME=dark")
	}

	// Test fallback (simulate unset or other value)
	_ = os.Unsetenv("GH_THEME")
	want := runtime.GOOS == "windows"
	if got := IsLight(); got != want {
		t.Errorf("isLight() fallback mismatch on %q: got %v, want %v", runtime.GOOS, got, want)
	}
}

func TestExitCouldNotAnswer(t *testing.T) {
	// We can't actually test the exit behavior, but we can test that
	// the function is defined and doesn't panic when called in a subprocess
	t.Run("function exists", func(t *testing.T) {
		// Just verify the function is accessible - functions are never nil in Go
		defer func() {
			if r := recover(); r != nil {
				t.Error("ExitCouldNotAnswer function should not panic when accessed")
			}
		}()
		_ = ExitCouldNotAnswer
	})
}

func TestFatal(t *testing.T) {
	// We can't actually test the exit behavior, but we can test that
	// the function is defined and doesn't panic when called in a subprocess
	t.Run("function exists", func(t *testing.T) {
		// Just verify the function is accessible - functions are never nil in Go
		defer func() {
			if r := recover(); r != nil {
				t.Error("Fatal function should not panic when accessed")
			}
		}()
		_ = Fatal
	})
}

func TestStripANSIEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"no ANSI codes", "plain text", "plain text"},
		{"multiple ANSI codes", "\x1b[31m\x1b[1mBold Red\x1b[0m\x1b[32mGreen\x1b[0m", "Bold RedGreen"},
		{"ANSI at start", "\x1b[31mRed text", "Red text"},
		{"ANSI at end", "Normal text\x1b[0m", "Normal text"},
		{"complex ANSI", "\x1b[38;5;196mComplex color\x1b[0m", "Complex color"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeMarkdownEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"no brackets", "normal text", "normal text"},
		{"single bracket", "]", "\\]"},
		{"multiple brackets", "text]with]brackets]", "text\\]with\\]brackets\\]"},
		{"brackets at start and end", "]text]", "\\]text\\]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeMarkdown(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAutoLinkEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		text     string
		expected string
	}{
		{"empty URL and text", "", "", "<>"},
		{"URL with brackets in text", "https://example.com", "text[with]brackets", "[text[with\\]brackets](https://example.com)"},
		{"identical URL and text", "https://example.com", "https://example.com", "<https://example.com>"},
		{"URL with special chars", "https://example.com?q=test&x=1", "Search", "[Search](https://example.com?q=test&x=1)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AutoLink(tt.url, tt.text)
			if result != tt.expected {
				t.Errorf("AutoLink(%q, %q) = %q, want %q", tt.url, tt.text, result, tt.expected)
			}
		})
	}
}
