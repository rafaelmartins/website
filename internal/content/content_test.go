package content

import (
	"bytes"
	"path/filepath"
	"testing"
)

func testdataPath(name string) string {
	return filepath.Join("testdata", name)
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		file     string
		expected bool
	}{
		{testdataPath("sample.md"), true},
		{testdataPath("sample.html"), true},
		{testdataPath("sample.textbundle"), true},
		{testdataPath("sample.textpack"), true},
		{"file.txt", false},
		{"file.unknown", false},
	}

	for _, tt := range tests {
		result := IsSupported(tt.file)
		if result != tt.expected {
			t.Errorf("IsSupported(%s) = %v, want %v", tt.file, result, tt.expected)
		}
	}
}

func TestRenderMarkdown(t *testing.T) {
	f := testdataPath("sample.md")
	html, meta, err := Render(f, "")
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if meta == nil {
		t.Fatal("render returned nil metadata")
	}

	if meta.Title != "Test Markdown" {
		t.Errorf("title=%q, want %q", meta.Title, "Test Markdown")
	}

	if meta.Description != "A test markdown file" {
		t.Errorf("description=%q, want %q", meta.Description, "A test markdown file")
	}

	if html == "" {
		t.Error("rendered HTML is empty")
	}

	if !bytes.Contains([]byte(html), []byte("Main Heading")) && !bytes.Contains([]byte(html), []byte("<h")) {
		t.Error("rendered HTML missing expected content")
	}
}

func TestRenderHTML(t *testing.T) {
	f := testdataPath("sample.html")
	html, meta, err := Render(f, "")
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if meta == nil {
		t.Fatal("render returned nil metadata")
	}

	if meta.Title != "Test HTML" {
		t.Errorf("title=%q, want %q", meta.Title, "Test HTML")
	}

	if meta.Description != "A test HTML file" {
		t.Errorf("description=%q, want %q", meta.Description, "A test HTML file")
	}

	if !bytes.Contains([]byte(html), []byte("<h1>")) {
		t.Error("rendered HTML missing expected <h1> tag")
	}
}

func TestRenderTextBundle(t *testing.T) {
	f := testdataPath("sample.textbundle")
	html, meta, err := Render(f, "")
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if meta == nil {
		t.Fatal("render returned nil metadata")
	}

	if meta.Title != "Main Heading" {
		t.Errorf("title=%q, want %q", meta.Title, "Main Heading")
	}

	if meta.Description != "A test textbundle file" {
		t.Errorf("description=%q, want %q", meta.Description, "A test textbundle file")
	}

	if html == "" {
		t.Error("rendered HTML is empty")
	}

	if !bytes.Contains([]byte(html), []byte("This is a test textbundle file")) {
		t.Error("rendered HTML missing expected content")
	}
}

func TestRenderTextBundleWithBaseURL(t *testing.T) {
	f := testdataPath("sample.textbundle")
	baseurl := "/blog/test"
	html, _, err := Render(f, baseurl)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !bytes.Contains([]byte(html), []byte("/blog/test/assets/image.png")) {
		t.Error("rendered HTML missing rewritten asset path")
	}
}

func TestRenderTextPack(t *testing.T) {
	f := testdataPath("sample.textpack")
	html, meta, err := Render(f, "")
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if meta == nil {
		t.Fatal("render returned nil metadata")
	}

	if meta.Title != "Main Heading" {
		t.Errorf("title=%q, want %q", meta.Title, "Main Heading")
	}

	if meta.Description != "A test textpack file" {
		t.Errorf("description=%q, want %q", meta.Description, "A test textpack file")
	}

	if html == "" {
		t.Error("rendered HTML is empty")
	}

	if !bytes.Contains([]byte(html), []byte("This is a test textpack file")) {
		t.Error("rendered HTML missing expected content")
	}
}

func TestRenderTextPackWithBaseURL(t *testing.T) {
	f := testdataPath("sample.textpack")
	baseurl := "/files/test"
	html, _, err := Render(f, baseurl)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !bytes.Contains([]byte(html), []byte("/files/test/assets/image.png")) {
		t.Error("rendered HTML missing rewritten asset path")
	}
}

func TestRenderUnsupported(t *testing.T) {
	_, _, err := Render("file.txt", "")
	if err == nil {
		t.Error("render should return error for unsupported file")
	}

	if err.Error() != "content: no provider found: file.txt" {
		t.Errorf("error message=%q, want error about no provider found", err.Error())
	}
}
