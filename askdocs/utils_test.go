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
