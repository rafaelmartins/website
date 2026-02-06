package markdown

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
)

func TestSpoiler(t *testing.T) {
	o := `<span class="spoiler">`
	c := "</span>"

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic spoiler",
			input:    `This is ||hidden text|| in a sentence.`,
			expected: "<p>This is " + o + "hidden text" + c + " in a sentence.</p>\n",
		},
		{
			name:     "spoiler with bold inside",
			input:    `||**bold secret**||`,
			expected: "<p>" + o + "<strong>bold secret</strong>" + c + "</p>\n",
		},
		{
			name:     "spoiler with italic inside",
			input:    `||*italic secret*||`,
			expected: "<p>" + o + "<em>italic secret</em>" + c + "</p>\n",
		},
		{
			name:     "multiple spoilers",
			input:    `||first|| and ||second||`,
			expected: "<p>" + o + "first" + c + " and " + o + "second" + c + "</p>\n",
		},
		{
			name:     "single pipe is not a spoiler",
			input:    `This |is not| a spoiler.`,
			expected: "<p>This |is not| a spoiler.</p>\n",
		},
		{
			name:     "unclosed spoiler",
			input:    `This ||is not closed.`,
			expected: "<p>This ||is not closed.</p>\n",
		},
		{
			name:     "empty spoiler is not parsed",
			input:    `||||`,
			expected: "<p>||||</p>\n",
		},
		{
			name:     "spoiler with link inside",
			input:    `||[click me](https://example.com)||`,
			expected: "<p>" + o + "<a href=\"https://example.com\">click me</a>" + c + "</p>\n",
		},
		{
			name:     "spoiler with code inside",
			input:    "||`secret code`||",
			expected: "<p>" + o + "<code>secret code</code>" + c + "</p>\n",
		},
		{
			name:     "spoiler in a larger paragraph",
			input:    `Hello, the answer is ||42|| and nothing else.`,
			expected: "<p>Hello, the answer is " + o + "42" + c + " and nothing else.</p>\n",
		},
	}

	md := goldmark.New(goldmark.WithExtensions(&spoilers{}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.input), &buf); err != nil {
				t.Fatalf("Convert failed: %v", err)
			}
			got := buf.String()
			if got != tt.expected {
				t.Errorf("got:\n%s\nexpected:\n%s", got, tt.expected)
			}
		})
	}
}
