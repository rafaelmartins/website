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
		name     string
		file     string
		expected bool
	}{
		{"markdown file", testdataPath("sample.md"), true},
		{"html file", testdataPath("sample.html"), true},
		{"textbundle file", testdataPath("sample.textbundle"), true},
		{"textpack file", testdataPath("sample.textpack"), true},
		{"unsupported txt file", "file.txt", false},
		{"unsupported unknown file", "file.unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSupported(tt.file)
			if result != tt.expected {
				t.Errorf("IsSupported(%s) = %v, want %v", tt.file, result, tt.expected)
			}
		})
	}
}

func TestRender(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		wantTitle    string
		wantDesc     string
		containsText string
	}{
		{"markdown", testdataPath("sample.md"), "Test Markdown", "A test markdown file", "Main Heading"},
		{"html", testdataPath("sample.html"), "Test HTML", "A test HTML file", "<h1>"},
		{"textbundle", testdataPath("sample.textbundle"), "Main Heading", "A test textbundle file", "This is a test textbundle file"},
		{"textpack", testdataPath("sample.textpack"), "Main Heading", "A test textpack file", "This is a test textpack file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, toc, html, err := Render(tt.file, "", nil)
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}

			if meta == nil {
				t.Fatal("render returned nil metadata")
			}

			if meta.Title != tt.wantTitle {
				t.Errorf("title=%q, want %q", meta.Title, tt.wantTitle)
			}

			if meta.Description != tt.wantDesc {
				t.Errorf("description=%q, want %q", meta.Description, tt.wantDesc)
			}

			if toc != "" {
				t.Errorf("toc should be empty with withToc=nil, got %q", toc)
			}

			if html == "" {
				t.Error("rendered HTML is empty")
			}

			if !bytes.Contains([]byte(html), []byte(tt.containsText)) {
				t.Errorf("rendered HTML missing expected content: %q", tt.containsText)
			}
		})
	}
}

func TestRenderWithToc(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		tocContains string
	}{
		{"markdown", testdataPath("sample.md"), `<a href="#main-heading">Main Heading</a>`},
		{"textbundle", testdataPath("sample.textbundle"), `<a href="#sub-heading">Sub Heading</a>`},
		{"textpack", testdataPath("sample.textpack"), `<a href="#sub-heading">Sub Heading</a>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, toc, _, err := Render(tt.file, "", new(true))
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}

			if !bytes.Contains([]byte(toc), []byte(tt.tocContains)) {
				t.Errorf("toc missing expected content %q, got %q", tt.tocContains, toc)
			}
		})
	}
}

func TestRenderTocFrontmatterOverride(t *testing.T) {
	_, toc, _, err := Render(testdataPath("sample.md"), "", new(false))
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if toc != "" {
		t.Errorf("toc should be empty when withToc=false and no frontmatter override, got %q", toc)
	}

	for _, tt := range []struct {
		file        string
		tocContains string
	}{
		{"sample-toc.md", `<a href="#heading">Heading</a>`},
		{"sample-toc.textbundle", `<a href="#sub-heading">Sub Heading</a>`},
	} {
		_, toc, _, err = Render(testdataPath(tt.file), "", new(false))
		if err != nil {
			t.Fatalf("render %s failed: %v", tt.file, err)
		}
		if !bytes.Contains([]byte(toc), []byte(tt.tocContains)) {
			t.Errorf("%s: frontmatter toc:true should override withToc=false, got toc=%q", tt.file, toc)
		}
	}

	for _, f := range []string{"sample-notoc.md", "sample-notoc.textbundle"} {
		_, toc, _, err = Render(testdataPath(f), "", new(true))
		if err != nil {
			t.Fatalf("render %s failed: %v", f, err)
		}
		if toc != "" {
			t.Errorf("%s: frontmatter toc:false should override withToc=true, got toc=%q", f, toc)
		}
	}

	for _, f := range []string{"sample-toc.md", "sample-toc.textbundle"} {
		_, toc, _, err = Render(testdataPath(f), "", nil)
		if err != nil {
			t.Fatalf("render %s failed: %v", f, err)
		}
		if toc != "" {
			t.Errorf("%s: withToc=nil should force toc off even with frontmatter toc:true, got toc=%q", f, toc)
		}
	}
}

func TestRenderWithBaseURL(t *testing.T) {
	tests := []struct {
		name            string
		file            string
		baseurl         string
		containsRewrite string
	}{
		{"textbundle with baseurl", testdataPath("sample.textbundle"), "/blog/test", "/blog/test/assets/image.png"},
		{"textpack with baseurl", testdataPath("sample.textpack"), "/files/test", "/files/test/assets/image.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, html, err := Render(tt.file, tt.baseurl, nil)
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}

			if !bytes.Contains([]byte(html), []byte(tt.containsRewrite)) {
				t.Error("rendered HTML missing rewritten asset path")
			}
		})
	}
}

func TestRenderUnsupported(t *testing.T) {
	_, _, _, err := Render("file.txt", "", nil)
	if err == nil {
		t.Error("render should return error for unsupported file")
	}

	if err.Error() != "content: no provider found: file.txt" {
		t.Errorf("error message=%q, want error about no provider found", err.Error())
	}
}
