package askdocs

import (
	"testing"
)

func TestFixIncompleteMarkdown(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"bold **", "This is **bold text", "This is **bold text**"},
		{"bold __", "This is __bold text", "This is __bold text__"},
		{"italic *", "This is *italic text", "This is *italic text*"},
		{"italic _", "This is _italic text", "This is _italic text_"},
		{"bold+italic ***", "This is ***bold and italic text", "This is ***bold and italic text***"},
		{"link text", "This is a [link text", "This is a [link text]"},
		{"link url", "This is a [link text](https://example.com", "This is a [link text](https://example.com)"},
		{"inline code", "This is `inline code", "This is `inline code`"},
		{"code block", "Here is some code:\n```\nconst x = 10;", "Here is some code:\n```\nconst x = 10;\n```"},
		{"nested bold/italic", "This is **bold and _italic text", "This is **bold and _italic text_**"},
		{"complete markdown unchanged", "This is **bold text** and *italic text*", "This is **bold text** and *italic text*"},
		{"multiline unclosed link", "Start of text [link text](https://example.com) and **bold text**\n\nI am a new paragraph with *italic text*\n\nThis is the end of the text, with a [link to the end", "Start of text [link text](https://example.com) and **bold text**\n\nI am a new paragraph with *italic text*\n\nThis is the end of the text, with a [link to the end]"},
		{"strikethrough", "This is ~~strikethrough text", "This is ~~strikethrough text~~"},
		{"image", "![Alt text](", "![Alt text]()"},
		{"nested emphasis", "Some _italic and **bold text", "Some _italic and **bold text**_"},
		{"code block with lang", "```javascript\nconsole.log(\"Hello, world!\");", "```javascript\nconsole.log(\"Hello, world!\");\n```"},
		{"heading unchanged", "### Heading level 3", "### Heading level 3"},
		{"horizontal rule unchanged", "Some text\n---", "Some text\n---"},
		{"table incomplete row", "| Header1 | Header2 |\n|---------|---------|\n| Row1Col1", "| Header1 | Header2 |\n|---------|---------|\n| Row1Col1 | |"},
		{"tilde emphasis", "This is ~tilde emphasis", "This is ~tilde emphasis~"},
		{"image alt text unclosed", "Here is an image ![alt text", "Here is an image ![alt text]"},
		{"image with url unclosed", "Image: ![alt](https://example.com/image.jpg", "Image: ![alt](https://example.com/image.jpg)"},
		{"image alt and url both unclosed", "Image: ![unclosed alt text](https://example.com", "Image: ![unclosed alt text](https://example.com)"},
		{"multiple images mixed", "![first image] and ![second unclosed", "![first image] and ![second unclosed]"},
		{"image at start", "![unclosed image", "![unclosed image]"},
		{"image at end", "Text before ![unclosed", "Text before ![unclosed]"},
		{"empty image alt", "![](complete.jpg) and ![", "![](complete.jpg) and ![]"},
		{"nested_image_syntax", "Text ![alt with [nested] brackets", "Text ![alt with [nested] brackets"},
		{"image with complex url", "![alt](https://example.com/path?param=value&other=test", "![alt](https://example.com/path?param=value&other=test)"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := fixIncompleteMarkdown(c.in)
			if got != c.want {
				t.Errorf("FixIncompleteMarkdown() = %q, want %q", got, c.want)
			}
		})
	}
}
