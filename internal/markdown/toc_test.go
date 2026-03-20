package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func newTocMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(&toc{}),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
}

func renderWithToc(t *testing.T, md goldmark.Markdown, input string) (body string, tocHTML string) {
	t.Helper()
	pc := parser.NewContext()
	pc.Set(PcTocEnable, new(true))
	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf, parser.WithContext(pc)); err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	toc, err := tocRender(pc)
	if err != nil {
		t.Fatalf("tocRender failed: %v", err)
	}
	return buf.String(), toc
}

func TestTocRender(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedToc  string
		containsBody []string
	}{
		{
			name:        "no headings",
			input:       "Just a paragraph.",
			expectedToc: "",
		},
		{
			name:  "single heading",
			input: "## Hello",
			expectedToc: `<details id="__toc__">
<summary>Table of contents</summary>
<ul>
<li><a href="#hello">Hello</a></li>
</ul>
</details>
`,
			containsBody: []string{
				`<h2 id="hello">Hello`,
				`<a href="#hello" class="toc-anchor has-text-link-light fa-solid fa-paragraph"></a>`,
				`<a href="#__toc__" class="toc-anchor has-text-link-light fa-solid fa-arrow-turn-up"></a>`,
			},
		},
		{
			name: "multiple same-level headings",
			input: `## First
## Second
## Third`,
			expectedToc: `<details id="__toc__">
<summary>Table of contents</summary>
<ul>
<li><a href="#first">First</a></li>
<li><a href="#second">Second</a></li>
<li><a href="#third">Third</a></li>
</ul>
</details>
`,
			containsBody: []string{
				`<h2 id="first">First`,
				`<h2 id="second">Second`,
				`<h2 id="third">Third`,
			},
		},
		{
			name: "nested headings",
			input: `## Parent
### Child`,
			expectedToc: `<details id="__toc__">
<summary>Table of contents</summary>
<ul>
<li><a href="#parent">Parent</a>
<ul>
<li><a href="#child">Child</a></li>
</ul>
</li>
</ul>
</details>
`,
			containsBody: []string{
				`<h2 id="parent">Parent`,
				`<h3 id="child">Child`,
			},
		},
		{
			name: "deeply nested headings",
			input: `## Level 2
### Level 3
#### Level 4`,
			expectedToc: `<details id="__toc__">
<summary>Table of contents</summary>
<ul>
<li><a href="#level-2">Level 2</a>
<ul>
<li><a href="#level-3">Level 3</a>
<ul>
<li><a href="#level-4">Level 4</a></li>
</ul>
</li>
</ul>
</li>
</ul>
</details>
`,
			containsBody: []string{
				`<h2 id="level-2">`,
				`<h3 id="level-3">`,
				`<h4 id="level-4">`,
			},
		},
		{
			name: "heading level goes back up",
			input: `## First
### Nested
## Second`,
			expectedToc: `<details id="__toc__">
<summary>Table of contents</summary>
<ul>
<li><a href="#first">First</a>
<ul>
<li><a href="#nested">Nested</a></li>
</ul>
</li>
<li><a href="#second">Second</a></li>
</ul>
</details>
`,
		},
		{
			name: "h1 with nested h2",
			input: `# Title
## Section`,
			expectedToc: `<details id="__toc__">
<summary>Table of contents</summary>
<ul>
<li><a href="#title">Title</a>
<ul>
<li><a href="#section">Section</a></li>
</ul>
</li>
</ul>
</details>
`,
			containsBody: []string{
				`<h1 id="title">Title`,
				`<h2 id="section">Section`,
			},
		},
		{
			name: "skipped heading level",
			input: `## H2
#### H4`,
			expectedToc: `<details id="__toc__">
<summary>Table of contents</summary>
<ul>
<li><a href="#h2">H2</a>
<ul>
<li>
<ul>
<li><a href="#h4">H4</a></li>
</ul>
</li>
</ul>
</li>
</ul>
</details>
`,
		},
		{
			name:  "heading with inline code",
			input: "## The `fmt` package",
			expectedToc: `<details id="__toc__">
<summary>Table of contents</summary>
<ul>
<li><a href="#the-fmt-package">The <code>fmt</code> package</a></li>
</ul>
</details>
`,
			containsBody: []string{
				`<h2 id="the-fmt-package">The <code>fmt</code> package`,
			},
		},
		{
			name:  "heading with bold text",
			input: "## A **bold** heading",
			expectedToc: `<details id="__toc__">
<summary>Table of contents</summary>
<ul>
<li><a href="#a-bold-heading">A <strong>bold</strong> heading</a></li>
</ul>
</details>
`,
			containsBody: []string{
				`<h2 id="a-bold-heading">A <strong>bold</strong> heading`,
			},
		},
	}

	md := newTocMarkdown()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, tocHTML := renderWithToc(t, md, tt.input)
			if tocHTML != tt.expectedToc {
				t.Errorf("toc got:\n%s\nexpected:\n%s", tocHTML, tt.expectedToc)
			}
			for _, s := range tt.containsBody {
				if !strings.Contains(body, s) {
					t.Errorf("body missing %q\ngot:\n%s", s, body)
				}
			}
		})
	}
}

func TestTocMultipleHeadingsAnchorCount(t *testing.T) {
	md := newTocMarkdown()
	body, _ := renderWithToc(t, md, "## A\n## B\n## C")

	if c := strings.Count(body, "fa-solid fa-paragraph"); c != 3 {
		t.Errorf("expected 3 paragraph anchors, got %d\nbody:\n%s", c, body)
	}
	if c := strings.Count(body, "fa-solid fa-arrow-turn-up"); c != 3 {
		t.Errorf("expected 3 back-to-top anchors, got %d\nbody:\n%s", c, body)
	}
}

func TestTocDisabled(t *testing.T) {
	md := newTocMarkdown()

	t.Run("not set", func(t *testing.T) {
		pc := parser.NewContext()
		var buf bytes.Buffer
		if err := md.Convert([]byte("## Heading"), &buf, parser.WithContext(pc)); err != nil {
			t.Fatalf("Convert failed: %v", err)
		}
		tocHTML, err := tocRender(pc)
		if err != nil {
			t.Fatalf("tocRender failed: %v", err)
		}
		if tocHTML != "" {
			t.Errorf("expected empty toc when not set, got:\n%s", tocHTML)
		}
		if !strings.Contains(buf.String(), `<h2 id="heading">Heading</h2>`) {
			t.Errorf("heading should render normally when toc disabled, got:\n%s", buf.String())
		}
		if strings.Contains(buf.String(), "toc-anchor") {
			t.Errorf("body should not contain toc anchors when disabled:\n%s", buf.String())
		}
	})

	t.Run("nil", func(t *testing.T) {
		pc := parser.NewContext()
		pc.Set(PcTocEnable, (*bool)(nil))
		var buf bytes.Buffer
		if err := md.Convert([]byte("## Heading"), &buf, parser.WithContext(pc)); err != nil {
			t.Fatalf("Convert failed: %v", err)
		}
		tocHTML, err := tocRender(pc)
		if err != nil {
			t.Fatalf("tocRender failed: %v", err)
		}
		if tocHTML != "" {
			t.Errorf("expected empty toc when nil, got:\n%s", tocHTML)
		}
		if strings.Contains(buf.String(), "toc-anchor") {
			t.Errorf("body should not contain toc anchors when nil:\n%s", buf.String())
		}
	})

	t.Run("false", func(t *testing.T) {
		pc := parser.NewContext()
		pc.Set(PcTocEnable, new(false))
		var buf bytes.Buffer
		if err := md.Convert([]byte("## Heading"), &buf, parser.WithContext(pc)); err != nil {
			t.Fatalf("Convert failed: %v", err)
		}
		tocHTML, err := tocRender(pc)
		if err != nil {
			t.Fatalf("tocRender failed: %v", err)
		}
		if tocHTML != "" {
			t.Errorf("expected empty toc when false, got:\n%s", tocHTML)
		}
		if strings.Contains(buf.String(), "toc-anchor") {
			t.Errorf("body should not contain toc anchors when false:\n%s", buf.String())
		}
	})
}
