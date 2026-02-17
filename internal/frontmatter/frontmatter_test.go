package frontmatter

import (
	"testing"
	"time"
)

func TestParseFrontMatterWithValidYAML(t *testing.T) {
	src := []byte(`---
title: Test Article
description: A test article
published: 2025-01-28
updated: 2025-01-29
menu: blog
license: MIT
author:
  name: John Doe
  email: john@example.com
opengraph:
  title: OG Title
  description: OG Description
  image: /path/to/image.png
extra:
  custom_field: custom_value
---
# Content

This is the content.
`)

	metadata, rest, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if metadata == nil {
		t.Fatal("metadata should not be nil")
	}

	if metadata.Title != "Test Article" {
		t.Errorf("title=%q, want %q", metadata.Title, "Test Article")
	}

	if metadata.Description != "A test article" {
		t.Errorf("description=%q, want %q", metadata.Description, "A test article")
	}

	if metadata.Menu != "blog" {
		t.Errorf("menu=%q, want %q", metadata.Menu, "blog")
	}

	if metadata.License != "MIT" {
		t.Errorf("license=%q, want %q", metadata.License, "MIT")
	}

	if metadata.Author.Name != "John Doe" {
		t.Errorf("author.name=%q, want %q", metadata.Author.Name, "John Doe")
	}

	if metadata.Author.Email != "john@example.com" {
		t.Errorf("author.email=%q, want %q", metadata.Author.Email, "john@example.com")
	}

	if metadata.OpenGraph.Title != "OG Title" {
		t.Errorf("opengraph.title=%q, want %q", metadata.OpenGraph.Title, "OG Title")
	}

	if metadata.OpenGraph.Image != "/path/to/image.png" {
		t.Errorf("opengraph.image=%q, want %q", metadata.OpenGraph.Image, "/path/to/image.png")
	}

	if metadata.Extra["custom_field"] != "custom_value" {
		t.Errorf("extra.custom_field=%q, want %q", metadata.Extra["custom_field"], "custom_value")
	}

	if metadata.Published.Time != time.Date(2025, 1, 28, 0, 0, 0, 0, time.UTC) {
		t.Errorf("published=%v, want %v", metadata.Published.Time, time.Date(2025, 1, 28, 0, 0, 0, 0, time.UTC))
	}

	if metadata.Updated.Time != time.Date(2025, 1, 29, 0, 0, 0, 0, time.UTC) {
		t.Errorf("updated=%v, want %v", metadata.Updated.Time, time.Date(2025, 1, 29, 0, 0, 0, 0, time.UTC))
	}

	if string(rest) != "# Content\n\nThis is the content.\n" {
		t.Errorf("rest content mismatch: got %q", string(rest))
	}
}

func TestParseFrontMatterWithDateFormats(t *testing.T) {
	tests := []struct {
		name          string
		src           []byte
		wantPublished time.Time
		wantUpdated   time.Time
		wantErr       bool
	}{
		{
			name: "full datetime format",
			src: []byte(`---
published: 2025-01-28 15:30:45
updated: 2025-01-29 16:45:30
---
content
`),
			wantPublished: time.Date(2025, 1, 28, 15, 30, 45, 0, time.UTC),
			wantUpdated:   time.Date(2025, 1, 29, 16, 45, 30, 0, time.UTC),
			wantErr:       false,
		},
		{
			name: "date only format",
			src: []byte(`---
published: 2025-01-28
updated: 2025-01-29
---
content
`),
			wantPublished: time.Date(2025, 1, 28, 0, 0, 0, 0, time.UTC),
			wantUpdated:   time.Date(2025, 1, 29, 0, 0, 0, 0, time.UTC),
			wantErr:       false,
		},
		{
			name: "invalid date format",
			src: []byte(`---
published: invalid-date
updated: 2025-01-29
---
content
`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, _, err := Parse(tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse error=%v, wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && metadata.Published.Time != tt.wantPublished {
				t.Errorf("published=%v, want %v", metadata.Published.Time, tt.wantPublished)
			}
			if !tt.wantErr && metadata.Updated.Time != tt.wantUpdated {
				t.Errorf("updated=%v, want %v", metadata.Updated.Time, tt.wantUpdated)
			}
		})
	}
}

func TestParseNoFrontMatter(t *testing.T) {
	src := []byte(`# No frontmatter
This is content without frontmatter.
`)

	metadata, rest, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if metadata == nil {
		t.Fatal("metadata should not be nil")
	}

	if string(rest) != string(src) {
		t.Errorf("rest should contain entire source when no frontmatter present")
	}

	if metadata.Title != "" {
		t.Errorf("title should be empty, got %q", metadata.Title)
	}
}

func TestParseEmptyFrontMatter(t *testing.T) {
	src := []byte(`---
---
Some content here
`)

	metadata, rest, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if metadata == nil {
		t.Fatal("metadata should not be nil")
	}

	if metadata.Title != "" {
		t.Errorf("title should be empty for empty frontmatter")
	}

	if string(rest) != "Some content here\n" {
		t.Errorf("rest=%q, want %q", string(rest), "Some content here\n")
	}
}

func TestParseMinimalFrontMatter(t *testing.T) {
	src := []byte(`---
title: Just Title
---
Rest of content
`)

	metadata, rest, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if metadata.Title != "Just Title" {
		t.Errorf("title=%q, want %q", metadata.Title, "Just Title")
	}

	if string(rest) != "Rest of content\n" {
		t.Errorf("rest=%q, want %q", string(rest), "Rest of content\n")
	}
}

func TestParseWithMultilineContent(t *testing.T) {
	src := []byte(`---
title: Article
description: Test
---
# Heading

Paragraph 1

Paragraph 2

List:
- Item 1
- Item 2
`)

	metadata, rest, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if metadata.Title != "Article" {
		t.Errorf("title=%q, want %q", metadata.Title, "Article")
	}

	expectedContent := `# Heading

Paragraph 1

Paragraph 2

List:
- Item 1
- Item 2
`
	if string(rest) != expectedContent {
		t.Errorf("rest content mismatch")
	}
}

func TestParseInvalidYAML(t *testing.T) {
	src := []byte(`---
title: Valid
invalid: [unclosed
---
content
`)

	_, _, err := Parse(src)
	if err == nil {
		t.Error("Parse should return error for invalid YAML")
	}
}

func TestParseWithOpenGraphImageGen(t *testing.T) {
	src := []byte(`---
title: Test
opengraph:
  image: /image.png
  image-gen:
    color: 255
    dpi: 72.0
    size: 1200.0
---
content
`)

	metadata, _, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if metadata.OpenGraph.Image != "/image.png" {
		t.Errorf("image=%q", metadata.OpenGraph.Image)
	}

	if metadata.OpenGraph.ImageGen.Color == nil || *metadata.OpenGraph.ImageGen.Color != 255 {
		t.Errorf("color=%v, want 255", metadata.OpenGraph.ImageGen.Color)
	}

	if metadata.OpenGraph.ImageGen.DPI == nil || *metadata.OpenGraph.ImageGen.DPI != 72.0 {
		t.Errorf("dpi=%v, want 72.0", metadata.OpenGraph.ImageGen.DPI)
	}

	if metadata.OpenGraph.ImageGen.Size == nil || *metadata.OpenGraph.ImageGen.Size != 1200.0 {
		t.Errorf("size=%v, want 1200.0", metadata.OpenGraph.ImageGen.Size)
	}
}

func TestParseWithExtraFields(t *testing.T) {
	src := []byte(`---
title: Test
extra:
  tags:
    - golang
    - testing
  featured: true
  count: 42
  ratio: 3.14
---
content
`)

	metadata, _, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if metadata.Extra == nil {
		t.Fatal("extra should not be nil")
	}

	tags, ok := metadata.Extra["tags"]
	if !ok {
		t.Error("tags not found in extra")
	}

	tagsList, ok := tags.([]any)
	if !ok || len(tagsList) != 2 {
		t.Errorf("tags=%v, expected array of 2 elements", tags)
	}

	if featured, ok := metadata.Extra["featured"].(bool); !ok || !featured {
		t.Errorf("featured=%v, want true", metadata.Extra["featured"])
	}

	if count, ok := metadata.Extra["count"].(int); !ok || count != 42 {
		t.Errorf("count=%v, want 42", metadata.Extra["count"])
	}
}

func TestParseWithWindowsLineEndings(t *testing.T) {
	src := []byte("---\r\ntitle: Test\r\n---\r\nContent\r\n")

	metadata, rest, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if metadata.Title != "Test" {
		t.Errorf("title=%q, want %q", metadata.Title, "Test")
	}

	if string(rest) != "Content\r\n" {
		t.Errorf("rest=%q, want %q", string(rest), "Content\r\n")
	}
}

func TestFrontMatterDateUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "datetime format",
			yaml:    "2025-01-28 15:30:45",
			want:    time.Date(2025, 1, 28, 15, 30, 45, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "date format",
			yaml:    "2025-01-28",
			want:    time.Date(2025, 1, 28, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "invalid format",
			yaml:    "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := []byte("---\npublished: " + tt.yaml + "\n---\ncontent\n")
			metadata, _, err := Parse(src)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse error=%v, wantErr=%v", err, tt.wantErr)
			}

			if !tt.wantErr && metadata.Published.Time != tt.want {
				t.Errorf("published=%v, want %v", metadata.Published.Time, tt.want)
			}
		})
	}
}

func TestParseEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		src       []byte
		wantErr   bool
		checkRest func([]byte) bool
	}{
		{
			name:    "only opening delimiter",
			src:     []byte("---\n"),
			wantErr: false,
			checkRest: func(rest []byte) bool {
				return len(rest) == 0
			},
		},
		{
			name:    "empty file",
			src:     []byte(""),
			wantErr: false,
			checkRest: func(rest []byte) bool {
				return len(rest) == 0
			},
		},
		{
			name:    "delimiter at end with newline",
			src:     []byte("---\ntitle: Test\n---\n"),
			wantErr: false,
			checkRest: func(rest []byte) bool {
				return len(rest) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, rest, err := Parse(tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse error=%v, wantErr=%v", err, tt.wantErr)
			}
			if !tt.checkRest(rest) {
				t.Errorf("rest check failed for %s: %q", tt.name, string(rest))
			}
		})
	}
}
