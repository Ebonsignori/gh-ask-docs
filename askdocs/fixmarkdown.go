package askdocs

import (
	"regexp"
	"strings"
)

// FixIncompleteMarkdown attempts to close common Markdown constructs
// when rendering a *partial* stream so that Glamour can colorise it safely.
// It follows the same logic as the original JavaScript helper you provided.
func fixIncompleteMarkdown(content string) string {
	content = fixCodeBlocks(content)
	content = fixInlineCode(content)
	content = fixLinks(content)
	content = fixImages(content)
	content = fixEmphasis(content)
	content = fixTables(content)
	return content
}

func fixCodeBlocks(s string) string {
	count := strings.Count(s, "```")
	if count%2 != 0 {
		s += "\n```"
	}
	return s
}

func fixInlineCode(s string) string {
	count := strings.Count(s, "`")
	if count%2 != 0 {
		s += "`"
	}
	return s
}

func fixLinks(s string) string {
	// unclosed link text '['
	if linkTextRe.MatchString(s) && !strings.Contains(s[strings.LastIndex(s, "["):], "]") {
		s += "]"
	}
	// unclosed link url '('
	if linkURLRe.MatchString(s) && !strings.HasSuffix(strings.TrimSpace(s), ")") {
		s += ")"
	}
	return s
}

func fixImages(s string) string {
	// unclosed alt text '!['
	if imgAltTextRe.MatchString(s) && !strings.Contains(s[strings.LastIndex(s, "!["):], "]") {
		s += "]"
	}
	// unclosed url
	if imgURLRe.MatchString(s) && !strings.HasSuffix(strings.TrimSpace(s), ")") {
		s += ")"
	}
	return s
}

func fixEmphasis(s string) string {
	tokens := []string{"***", "**", "__", "*", "_", "~~", "~"}

	type stackElem struct {
		token string
	}
	stack := []stackElem{}

	i := 0
	for i < len(s) {
		matched := false
		for _, tok := range tokens {
			if strings.HasPrefix(s[i:], tok) {
				if len(stack) > 0 && stack[len(stack)-1].token == tok {
					stack = stack[:len(stack)-1] // closing tag
				} else {
					stack = append(stack, stackElem{token: tok}) // opening tag
				}
				i += len(tok)
				matched = true
				break
			}
		}
		if !matched {
			i++
		}
	}

	for len(stack) > 0 {
		tok := stack[len(stack)-1].token
		stack = stack[:len(stack)-1]
		s += tok
	}

	return s
}

func fixTables(s string) string {
	lines := strings.Split(s, "\n")
	inTable := false
	headerPipes := 0

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if tableLineRe.MatchString(line) {
			if !inTable {
				// Potential header line – look ahead for separator
				if i+1 < len(lines) && tableSepRe.MatchString(lines[i+1]) {
					inTable = true
					headerPipes = strings.Count(line, "|")
					continue
				}
			} else {
				// Table body – pad missing columns
				diff := headerPipes - strings.Count(line, "|")
				if diff > 0 {
					lines[i] = strings.TrimRight(line, " ") + strings.Repeat(" |", diff)
				}
			}
		} else {
			inTable = false
			headerPipes = 0
		}
	}

	return strings.Join(lines, "\n")
}

// Pre‑compiled regexps
var (
	linkTextRe   = regexp.MustCompile(`\[[^\]]*$`)
	linkURLRe    = regexp.MustCompile(`\]\([^)]*$`)
	imgAltTextRe = regexp.MustCompile(`!\[[^\]]*$`)
	imgURLRe     = regexp.MustCompile(`!\[[^\]]*\([^)]*$`)

	tableLineRe = regexp.MustCompile(`^\s*\|.*$`)
	tableSepRe  = regexp.MustCompile(`^\s*\|[-:|\s]*$`)
)
