package askdocs

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

var SpinnerFrames = []rune{'|', '/', '-', '\\'}

// newRenderer returns a Glamour renderer with the provided theme and wrap width.
func NewRenderer(theme string, wrap int) *glamour.TermRenderer {
	opts := []glamour.TermRendererOption{
		glamour.WithStandardStyle(theme),
		glamour.WithWordWrap(wrap),
	}
	r, _ := glamour.NewTermRenderer(opts...)
	return r
}

// renderFrame renders the buffer plus a spinner, clearing the previous frame first.
func RenderFrame(r *glamour.TermRenderer, raw string, spin rune, prevLines *int) {
	safe := fixIncompleteMarkdown(raw)
	base, _ := r.Render(safe)
	out := base + string(spin) + "\n"
	clearLines(*prevLines)
	fmt.Print(out)
	*prevLines = countVisualLines(out)
}

// renderSpinner prints a single rune to stderr, keeping stdout clean.
func RenderSpinner(spin rune) {
	fmt.Fprintf(os.Stderr, "\r%c", spin)
}

// clearLines erases the given number of terminal lines above the cursor.
func clearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Print("\033[2K") // clear entire line
		if i < n-1 {
			fmt.Print("\033[1A") // move cursor up
		}
	}
	fmt.Print("\r")
}

// countVisualLines estimates how many terminal lines the string will occupy,
// taking ANSI escapes and terminal width into account.
func countVisualLines(s string) int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		width = 80
	}
	plain := StripANSI(s)

	lines := 0
	for _, l := range strings.Split(plain, "\n") {
		if l == "" {
			lines++
			continue
		}
		runes := len([]rune(l))
		lines += (runes + width - 1) / width
	}
	return lines
}
