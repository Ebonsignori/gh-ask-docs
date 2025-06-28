package askdocs

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestExitCouldNotAnswerOutput(t *testing.T) {
	// Test the output of ExitCouldNotAnswer by running it in a subprocess
	if os.Getenv("TEST_EXIT_COULD_NOT_ANSWER") == "1" {
		ExitCouldNotAnswer()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExitCouldNotAnswerOutput")
	cmd.Env = append(os.Environ(), "TEST_EXIT_COULD_NOT_ANSWER=1")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should exit with status 1
	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
		}
	} else {
		t.Error("Expected exit error, but command succeeded")
	}

	// Check output
	output := stdout.String()
	if !strings.Contains(output, "⚠️  The AI could not answer your question.") {
		t.Errorf("Expected warning message in output, got: %q", output)
	}
}

func TestFatalOutput(t *testing.T) {
	// Test the output of Fatal by running it in a subprocess
	if os.Getenv("TEST_FATAL") == "1" {
		Fatal(errors.New("test error message"))
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatalOutput")
	cmd.Env = append(os.Environ(), "TEST_FATAL=1")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should exit with status 1
	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
		}
	} else {
		t.Error("Expected exit error, but command succeeded")
	}

	// Check stderr output
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "error:") {
		t.Errorf("Expected 'error:' in stderr, got: %q", stderrOutput)
	}
	if !strings.Contains(stderrOutput, "test error message") {
		t.Errorf("Expected error message in stderr, got: %q", stderrOutput)
	}
}

func TestFixImagesEdgeCases(t *testing.T) {
	// Test specific edge cases that might not be covered in the main test
	tests := []struct {
		name, in, want string
	}{
		{"image with spaces in alt", "![alt text with spaces", "![alt text with spaces]"},
		{"image at very end", "Text ![", "Text ![]"},
		{"image url with query params", "![alt](https://example.com/img.jpg?v=1&size=large", "![alt](https://example.com/img.jpg?v=1&size=large"},
		{"multiple unclosed images", "![first and ![second", "![first and ![second]"},
		{"image with no closing bracket after url", "![alt](url", "![alt](url"},
		{"image with whitespace before closing", "![alt](url   ", "![alt](url   "},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			got := fixImages(c.in)
			if got != c.want {
				t.Errorf("fixImages(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestLoadSupportedVersionsFileNotFound(t *testing.T) {
	// Save current working directory
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	// Create temporary directory without data file
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	_, err := LoadSupportedVersions()
	if err == nil {
		t.Error("Expected error when data file not found")
	}

	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("Expected file not found error, got: %v", err)
	}
}

func TestLoadSupportedVersionsInvalidJSON(t *testing.T) {
	// Save current working directory
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	// Create temporary directory with invalid JSON file
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Create data directory and invalid JSON file
	os.Mkdir("data", 0755)
	invalidJSON := `{"lastUpdated": "invalid json`
	err := os.WriteFile("data/supported-versions.json", []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = LoadSupportedVersions()
	if err == nil {
		t.Error("Expected error when JSON is invalid")
	}

	if !strings.Contains(err.Error(), "unexpected end of JSON input") && !strings.Contains(err.Error(), "invalid character") {
		t.Errorf("Expected JSON parsing error, got: %v", err)
	}
}

func TestFixTablesMoreCases(t *testing.T) {
	// Test additional table edge cases for better coverage
	tests := []struct {
		name, in, want string
	}{
		{
			"table with no separator",
			"| Header1 | Header2 |\n| Row1Col1 | Row1Col2 |",
			"| Header1 | Header2 |\n| Row1Col1 | Row1Col2 |",
		},
		{
			"incomplete table at end",
			"| Header1 | Header2 | Header3 |\n|---------|---------|----------|\n| Col1",
			"| Header1 | Header2 | Header3 |\n|---------|---------|----------|\n| Col1 | | |",
		},
		{
			"table with extra pipes",
			"| Header1 | Header2 |\n|---------|----------|\n| Col1 | Col2 | Extra |",
			"| Header1 | Header2 |\n|---------|----------|\n| Col1 | Col2 | Extra |",
		},
		{
			"non-table line resets state",
			"| Header1 | Header2 |\n|---------|----------|\nNot a table line\n| Col1",
			"| Header1 | Header2 |\n|---------|----------|\nNot a table line\n| Col1",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			got := fixTables(c.in)
			if got != c.want {
				t.Errorf("fixTables(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestCountVisualLinesWithTerminalError(t *testing.T) {
	// Test countVisualLines when terminal size cannot be determined
	// This is harder to test directly, but we can at least verify it handles
	// the fallback case gracefully

	// Test with content that would definitely wrap on a small terminal
	longContent := strings.Repeat("x", 1000)
	result := countVisualLines(longContent)

	// Should return a reasonable number of lines
	if result <= 0 {
		t.Errorf("countVisualLines should return positive number, got %d", result)
	}

	// Should be more than 1 line for very long content
	if result == 1 {
		t.Error("countVisualLines should wrap very long content to multiple lines")
	}
}

func TestRegexPatterns(t *testing.T) {
	// Test the compiled regex patterns used in fixmarkdown.go
	tests := []struct {
		name    string
		pattern string
		text    string
		should  bool
	}{
		{"linkTextRe matches", `\[[^\]]*$`, "This is a [link text", true},
		{"linkTextRe no match", `\[[^\]]*$`, "This is a [link text]", false},
		{"linkURLRe matches", `\]\([^)]*$`, "Text [link](https://example.com", true},
		{"linkURLRe no match", `\]\([^)]*$`, "Text [link](https://example.com)", false},
		{"imgAltTextRe matches", `!\[[^\]]*$`, "Here is ![alt text", true},
		{"imgAltTextRe no match", `!\[[^\]]*$`, "Here is ![alt text]", false},
		{"tableLineRe matches", `^\s*\|.*$`, "| Header | Value |", true},
		{"tableLineRe no match", `^\s*\|.*$`, "Not a table line", false},
		{"tableSepRe matches", `^\s*\|[-:|\s]*$`, "|-------|-------|", true},
		{"tableSepRe no match", `^\s*\|[-:|\s]*$`, "| Data | Value |", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't access the private compiled regexes directly,
			// but we can test the functions that use them
			switch {
			case strings.Contains(tt.name, "linkText"):
				result := fixLinks(tt.text)
				hasMatch := result != tt.text
				if hasMatch != tt.should {
					t.Errorf("linkTextRe pattern test failed for %q", tt.text)
				}
			case strings.Contains(tt.name, "linkURL"):
				result := fixLinks(tt.text)
				hasMatch := result != tt.text
				if hasMatch != tt.should {
					t.Errorf("linkURLRe pattern test failed for %q", tt.text)
				}
			case strings.Contains(tt.name, "imgAlt"):
				result := fixImages(tt.text)
				hasMatch := result != tt.text
				if hasMatch != tt.should {
					t.Errorf("imgAltTextRe pattern test failed for %q", tt.text)
				}
			case strings.Contains(tt.name, "table"):
				// For table patterns, we test the actual functions
				result := fixTables(tt.text)
				// Table fixing is more complex, so we just ensure it doesn't crash
				_ = result
			}
		})
	}
}
