package askdocs

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestNewRenderer(t *testing.T) {
	tests := []struct {
		name  string
		theme string
		wrap  int
	}{
		{"dark theme no wrap", "dark", 0},
		{"light theme with wrap", "light", 80},
		{"auto theme with wrap", "auto", 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewRenderer(tt.theme, tt.wrap)
			if renderer == nil {
				t.Error("NewRenderer returned nil")
			}

			// Test that we can render some basic markdown
			markdown := "# Test Header\n\nThis is a test."
			output, err := renderer.Render(markdown)
			if err != nil {
				t.Errorf("Failed to render markdown: %v", err)
			}
			if output == "" {
				t.Error("Renderer produced empty output")
			}
		})
	}
}

func TestRenderFrame(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer("dark", 80)
	prevLines := 0

	// Test basic rendering
	RenderFrame(renderer, "# Test\nSome content", '|', &prevLines)

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "|") {
		t.Error("RenderFrame should include spinner character")
	}

	if prevLines == 0 {
		t.Error("prevLines should be updated after rendering")
	}
}

func TestRenderSpinner(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Test spinner rendering
	RenderSpinner('/')

	// Close writer and restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "/") {
		t.Error("RenderSpinner should output the spinner character")
	}
	if !strings.Contains(output, "\r") {
		t.Error("RenderSpinner should include carriage return")
	}
}

func TestClearLines(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test clearing lines
	clearLines(3)

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should contain ANSI escape sequences for clearing lines
	if !strings.Contains(output, "\033[2K") {
		t.Error("clearLines should contain line clear escape sequence")
	}
	if !strings.Contains(output, "\033[1A") {
		t.Error("clearLines should contain cursor up escape sequence")
	}
	if !strings.Contains(output, "\r") {
		t.Error("clearLines should end with carriage return")
	}
}

func TestClearLinesZero(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test clearing zero lines
	clearLines(0)

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should only contain carriage return
	if output != "\r" {
		t.Errorf("clearLines(0) should only output carriage return, got %q", output)
	}
}

func TestCountVisualLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			"empty string",
			"",
			1,
		},
		{
			"single line",
			"Hello world",
			1,
		},
		{
			"multiple lines",
			"Line 1\nLine 2\nLine 3",
			3,
		},
		{
			"empty lines",
			"Line 1\n\nLine 3",
			3,
		},
		{
			"with ANSI codes",
			"\x1b[31mRed text\x1b[0m",
			1,
		},
		{
			"mixed content",
			"Normal\n\x1b[32mGreen\x1b[0m\nMore text",
			3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countVisualLines(tt.input)
			if result != tt.expected {
				t.Errorf("countVisualLines(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCountVisualLinesLongLines(t *testing.T) {
	// Test with a very long line that should wrap
	longLine := strings.Repeat("a", 200)
	result := countVisualLines(longLine)

	// Should be more than 1 line due to wrapping
	// Exact value depends on terminal width, but should be > 1
	if result <= 1 {
		t.Errorf("countVisualLines with long line should return > 1, got %d", result)
	}
}

func TestRenderFrameWithIncompleteMarkdown(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer("dark", 80)
	prevLines := 0

	// Test with incomplete markdown that should be fixed
	incompleteMarkdown := "This is **bold text\nAnd this is `inline code"
	RenderFrame(renderer, incompleteMarkdown, '-', &prevLines)

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "-") {
		t.Error("RenderFrame should include the provided spinner character")
	}
}

func TestSpinnerFrames(t *testing.T) {
	// Test that SpinnerFrames is properly defined
	if len(SpinnerFrames) == 0 {
		t.Error("SpinnerFrames should not be empty")
	}

	expectedFrames := []rune{'|', '/', '-', '\\'}
	if len(SpinnerFrames) != len(expectedFrames) {
		t.Errorf("Expected %d spinner frames, got %d", len(expectedFrames), len(SpinnerFrames))
	}

	for i, expected := range expectedFrames {
		if i < len(SpinnerFrames) && SpinnerFrames[i] != expected {
			t.Errorf("SpinnerFrames[%d] = %c, want %c", i, SpinnerFrames[i], expected)
		}
	}
}
