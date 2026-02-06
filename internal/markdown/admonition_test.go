package markdown

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
)

func TestAdmonition(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "NOTE admonition",
			input: `> [!NOTE]
> This is a note.`,
			expected: `<div class="admonition admonition-note">
<p class="admonition-title"><i class="fas fa-circle-info"></i>Note</p>
<p>This is a note.</p>
</div>
`,
		},
		{
			name: "TIP admonition",
			input: `> [!TIP]
> This is a tip.`,
			expected: `<div class="admonition admonition-tip">
<p class="admonition-title"><i class="fas fa-lightbulb"></i>Tip</p>
<p>This is a tip.</p>
</div>
`,
		},
		{
			name: "IMPORTANT admonition",
			input: `> [!IMPORTANT]
> This is important.`,
			expected: `<div class="admonition admonition-important">
<p class="admonition-title"><i class="fas fa-circle-exclamation"></i>Important</p>
<p>This is important.</p>
</div>
`,
		},
		{
			name: "WARNING admonition",
			input: `> [!WARNING]
> This is a warning.`,
			expected: `<div class="admonition admonition-warning">
<p class="admonition-title"><i class="fas fa-triangle-exclamation"></i>Warning</p>
<p>This is a warning.</p>
</div>
`,
		},
		{
			name: "CAUTION admonition",
			input: `> [!CAUTION]
> This is a caution.`,
			expected: `<div class="admonition admonition-caution">
<p class="admonition-title"><i class="fas fa-circle-xmark"></i>Caution</p>
<p>This is a caution.</p>
</div>
`,
		},
		{
			name: "admonition with multiple lines",
			input: `> [!NOTE]
> Line one.
> Line two.`,
			expected: `<div class="admonition admonition-note">
<p class="admonition-title"><i class="fas fa-circle-info"></i>Note</p>
<p>Line one.
Line two.</p>
</div>
`,
		},
		{
			name:  "regular blockquote (not admonition)",
			input: `> This is a regular blockquote.`,
			expected: `<blockquote>
<p>This is a regular blockquote.</p>
</blockquote>
`,
		},
		{
			name: "lowercase level should not match",
			input: `> [!note]
> This should be a blockquote.`,
			expected: `<blockquote>
<p>[!note]
This should be a blockquote.</p>
</blockquote>
`,
		},
		{
			name: "admonition with multiple paragraphs",
			input: `> [!NOTE]
> First paragraph.
>
> Second paragraph.`,
			expected: `<div class="admonition admonition-note">
<p class="admonition-title"><i class="fas fa-circle-info"></i>Note</p>
<p>First paragraph.</p>
<p>Second paragraph.</p>
</div>
`,
		},
		{
			name: "admonition with code block",
			input: `> [!TIP]
> Here is some code:
>
> ` + "```" + `go
> func main() {}
> ` + "```",
			expected: `<div class="admonition admonition-tip">
<p class="admonition-title"><i class="fas fa-lightbulb"></i>Tip</p>
<p>Here is some code:</p>
<pre><code class="language-go">func main() {}
</code></pre>
</div>
`,
		},
		{
			name: "admonition with inline code",
			input: `> [!NOTE]
> Use the ` + "`" + `fmt.Println` + "`" + ` function.`,
			expected: `<div class="admonition admonition-note">
<p class="admonition-title"><i class="fas fa-circle-info"></i>Note</p>
<p>Use the <code>fmt.Println</code> function.</p>
</div>
`,
		},
		{
			name: "admonition with unordered list",
			input: `> [!IMPORTANT]
> Remember these points:
>
> - First item
> - Second item
> - Third item`,
			expected: `<div class="admonition admonition-important">
<p class="admonition-title"><i class="fas fa-circle-exclamation"></i>Important</p>
<p>Remember these points:</p>
<ul>
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ul>
</div>
`,
		},
		{
			name: "admonition with ordered list",
			input: `> [!WARNING]
> Follow these steps:
>
> 1. Step one
> 2. Step two
> 3. Step three`,
			expected: `<div class="admonition admonition-warning">
<p class="admonition-title"><i class="fas fa-triangle-exclamation"></i>Warning</p>
<p>Follow these steps:</p>
<ol>
<li>Step one</li>
<li>Step two</li>
<li>Step three</li>
</ol>
</div>
`,
		},
		{
			name: "admonition with link",
			input: `> [!NOTE]
> See [the documentation](https://example.com) for more info.`,
			expected: `<div class="admonition admonition-note">
<p class="admonition-title"><i class="fas fa-circle-info"></i>Note</p>
<p>See <a href="https://example.com">the documentation</a> for more info.</p>
</div>
`,
		},
		{
			name: "admonition with bold and italic",
			input: `> [!CAUTION]
> This is **very important** and *should not* be ignored.`,
			expected: `<div class="admonition admonition-caution">
<p class="admonition-title"><i class="fas fa-circle-xmark"></i>Caution</p>
<p>This is <strong>very important</strong> and <em>should not</em> be ignored.</p>
</div>
`,
		},
		{
			name: "admonition with nested blockquote",
			input: `> [!NOTE]
> Someone said:
>
> > This is a nested quote.`,
			expected: `<div class="admonition admonition-note">
<p class="admonition-title"><i class="fas fa-circle-info"></i>Note</p>
<p>Someone said:</p>
<blockquote>
<p>This is a nested quote.</p>
</blockquote>
</div>
`,
		},
	}

	md := goldmark.New(goldmark.WithExtensions(&admonitions{}))

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
