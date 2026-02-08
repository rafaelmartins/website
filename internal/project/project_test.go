package project

import (
	"path/filepath"
	"testing"
)

func TestHandleImageUrl(t *testing.T) {
	args := []struct {
		name        string
		img         string
		subdir      string
		currentPage string
		rv1         string
		rv2         string
		hasErr      bool
	}{
		{"empty all", "", "", "", "", "", false},
		{"empty img with subdir", "", "subdir", "", "", "", false},

		{"simple image.png", "image.png", "", "", "image.png", "image.png", false},
		{"./image.png", "./image.png", "", "", "image.png", "image.png", false},
		{"folder/image.png", "folder/image.png", "", "", "folder/image.png", "folder/image.png", false},
		{"./folder/image.png", "./folder/image.png", "", "", "folder/image.png", "folder/image.png", false},
		{"../image.png error", "../image.png", "", "", "", "", true},
		{"../../folder/image.png error", "../../folder/image.png", "", "", "", "", true},

		{"image.png with sub", "image.png", "sub", "", "sub/image.png", "sub/image.png", false},
		{"./image.png with sub", "./image.png", "sub", "", "sub/image.png", "sub/image.png", false},
		{"folder/image.png with sub", "folder/image.png", "sub", "", "sub/folder/image.png", "sub/folder/image.png", false},
		{"./folder/image.png with sub", "./folder/image.png", "sub", "", "sub/folder/image.png", "sub/folder/image.png", false},
		{"../image.png with sub", "../image.png", "sub", "", "image.png", "image.png", false},
		{"../../image.png with sub error", "../../image.png", "sub", "", "", "", true},
		{"../../../image.png with sub error", "../../../image.png", "sub", "", "", "", true},

		{"image.png with sub/nested", "image.png", "sub/nested", "", "sub/nested/image.png", "sub/nested/image.png", false},
		{"./image.png with sub/nested", "./image.png", "sub/nested", "", "sub/nested/image.png", "sub/nested/image.png", false},
		{"folder/image.png with sub/nested", "folder/image.png", "sub/nested", "", "sub/nested/folder/image.png", "sub/nested/folder/image.png", false},
		{"../image.png with sub/nested", "../image.png", "sub/nested", "", "sub/image.png", "sub/image.png", false},
		{"../../image.png with sub/nested", "../../image.png", "sub/nested", "", "image.png", "image.png", false},
		{"../../../image.png with sub/nested error", "../../../image.png", "sub/nested", "", "", "", true},

		{"/image.png", "/image.png", "", "", "image.png", "image.png", false},
		{"/folder/image.png", "/folder/image.png", "", "", "folder/image.png", "folder/image.png", false},
		{"/image.png with sub", "/image.png", "sub", "", "image.png", "image.png", false},
		{"/folder/image.png with sub", "/folder/image.png", "sub", "", "folder/image.png", "folder/image.png", false},
		{"/folder/image.png with sub/nested", "/folder/image.png", "sub/nested", "", "folder/image.png", "folder/image.png", false},

		{"@@ simple", "@@image.png", "", "", "", "image.png", false},
		{"@@ with path", "@@folder/image.png", "", "", "", "folder/image.png", false},
		{"@@ with sub", "@@image.png", "sub", "", "", "image.png", false},
		{"@@ with sub and currentPage", "@@image.png", "sub", "p", "", "image.png", false},
		{"@@ absolute", "@@/image.png", "", "", "", "/image.png", false},
		{"@@ empty after prefix", "@@", "", "", "", "", false},

		{"http url", "http://example.com/image.png", "", "", "", "", false},
		{"https url", "https://example.com/image.png", "", "", "", "", false},
		{"ftp url", "ftp://example.com/image.png", "", "", "", "", false},
		{"protocol-relative url", "//example.com/image.png", "", "", "", "", false},
		{"http url with sub", "http://example.com/image.png", "sub", "", "", "", false},
		{"https url with sub/nested", "https://example.com/folder/image.png", "sub/nested", "", "", "", false},

		{"image.png?v=1", "image.png?v=1", "", "", "image.png", "image.png", false},
		{"image.png#section", "image.png#section", "", "", "image.png", "image.png", false},
		{"image.png?v=1#section", "image.png?v=1#section", "", "", "image.png", "image.png", false},
		{"folder/image.png?v=1 with sub", "folder/image.png?v=1", "sub", "", "sub/folder/image.png", "sub/folder/image.png", false},
		{"/image.png?v=1 with sub", "/image.png?v=1", "sub", "", "image.png", "image.png", false},

		{"image.png currentPage p", "image.png", "", "p", "image.png", "../image.png", false},
		{"image.png with sub currentPage p", "image.png", "sub", "p", "sub/image.png", "../sub/image.png", false},
		{"/image.png currentPage p", "/image.png", "", "p", "image.png", "../image.png", false},
		{"image.png with p currentPage p", "image.png", "p", "p", "p/image.png", "image.png", false},
		{"../image.png with p/sub currentPage p", "../image.png", "p/sub", "p", "p/image.png", "image.png", false},

		{"image.png currentPage p/q", "image.png", "", "p/q", "image.png", "../../image.png", false},
		{"image.png with sub currentPage p/q", "image.png", "sub", "p/q", "sub/image.png", "../../sub/image.png", false},
		{"image.png with p currentPage p/q", "image.png", "p", "p/q", "p/image.png", "../image.png", false},
		{"image.png with p/q currentPage p/q", "image.png", "p/q", "p/q", "p/q/image.png", "image.png", false},
		{"image.png with p/q/r currentPage p/q", "image.png", "p/q/r", "p/q", "p/q/r/image.png", "r/image.png", false},

		{"image.png currentPage a/b/c", "image.png", "", "a/b/c", "image.png", "../../../image.png", false},
		{"image.png with a currentPage a/b/c", "image.png", "a", "a/b/c", "a/image.png", "../../image.png", false},
		{"image.png with a/b currentPage a/b/c", "image.png", "a/b", "a/b/c", "a/b/image.png", "../image.png", false},
		{"image.png with a/b/c currentPage a/b/c", "image.png", "a/b/c", "a/b/c", "a/b/c/image.png", "image.png", false},

		{"../image.png with sub currentPage p", "../image.png", "sub", "p", "image.png", "../image.png", false},
		{"../../image.png with p/sub currentPage p", "../../image.png", "p/sub", "p", "image.png", "../image.png", false},
		{"../folder/image.png with sub/nested currentPage p/q", "../folder/image.png", "sub/nested", "p/q", "sub/folder/image.png", "../../sub/folder/image.png", false},

		{"/image.png currentPage a/b", "/image.png", "", "a/b", "image.png", "../../image.png", false},
		{"/folder/image.png currentPage a/b/c", "/folder/image.png", "", "a/b/c", "folder/image.png", "../../../folder/image.png", false},
		{"/folder/image.png with sub currentPage a/b/c", "/folder/image.png", "sub", "a/b/c", "folder/image.png", "../../../folder/image.png", false},
	}

	for _, tt := range args {
		t.Run(tt.name, func(t *testing.T) {
			proj := &Project{subdir: tt.subdir}
			rv1, rv2, err := proj.handleImageUrl(tt.img, tt.currentPage)
			if tt.hasErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if rv1 != tt.rv1 {
				t.Errorf("bad rv1: got %q, want %q", rv1, tt.rv1)
			}
			if rv2n := filepath.ToSlash(rv2); rv2n != tt.rv2 {
				t.Errorf("bad rv2: got %q, want %q", rv2n, tt.rv2)
			}
		})
	}
}

func TestHandleLinkUrl(t *testing.T) {
	pageResolvers := []*projectPageResolver{
		{src: "docs/index.md", name: "."},
		{src: "docs/asd.md", name: "asd"},
		{src: "docs/guide.md", name: "guide"},
		{src: "docs/page.md", name: "page"},
	}

	args := []struct {
		name        string
		link        string
		subdir      string
		currentPage string
		pages       []*projectPageResolver
		expGH       bool
		expRV       string
	}{
		{"@@ simple", "@@page", "docs", ".", pageResolvers, false, "page"},
		{"@@ with path", "@@some/path", "docs", ".", pageResolvers, false, "some/path"},
		{"@@ with sub", "@@page", "sub", ".", pageResolvers, false, "page"},
		{"@@ absolute", "@@/absolute/path", "", ".", pageResolvers, false, "/absolute/path"},
		{"@@ empty after prefix", "@@", "docs", ".", pageResolvers, false, ""},

		{"https url", "https://example.com/foo", "sub", ".", pageResolvers, false, ""},
		{"http url", "http://example.com/foo", "", "", pageResolvers, false, ""},
		{"protocol-relative url", "//example.com/a/b", "", "", pageResolvers, false, ""},
		{"ftp url", "ftp://example.com/file", "docs", ".", pageResolvers, false, ""},

		{"/assets/image.png", "/assets/image.png", "sub", ".", pageResolvers, true, "assets/image.png"},
		{"/assets/image.png no subdir", "/assets/image.png", "", ".", pageResolvers, true, "assets/image.png"},
		{"/assets/image.png ignored subdir", "/assets/image.png", "ignored", ".", pageResolvers, true, "assets/image.png"},

		{"/docs/index.md", "/docs/index.md", "", ".", pageResolvers, false, "./"},
		{"/docs/guide.md", "/docs/guide.md", "", ".", pageResolvers, false, "guide/"},
		{"/docs/asd.md from asd", "/docs/asd.md", "", "asd", pageResolvers, false, "./"},
		{"/docs/asd.md from guide", "/docs/asd.md", "", "guide", pageResolvers, false, "../asd/"},
		{"/docs/page.md", "/docs/page.md", "", ".", pageResolvers, false, "page/"},
		{"/docs/nested/deep.md", "/docs/nested/deep.md", "", ".", pageResolvers, true, "docs/nested/deep.md"},
		{"/docs/nested/deep.md from asd", "/docs/nested/deep.md", "", "asd", pageResolvers, true, "docs/nested/deep.md"},

		{"guide.md", "guide.md", "docs", ".", pageResolvers, false, "guide/"},
		{"asd.md", "asd.md", "docs", ".", pageResolvers, false, "asd/"},
		{"asd.md from guide", "asd.md", "docs", "guide", pageResolvers, false, "../asd/"},
		{"page.md", "page.md", "docs", ".", pageResolvers, false, "page/"},
		{"index.md", "index.md", "docs", ".", pageResolvers, false, "./"},
		{"index.md from guide", "index.md", "docs", "guide", pageResolvers, false, "../"},
		{"index.md from asd", "index.md", "docs", "asd", pageResolvers, false, "../"},
		{"nested/deep.md", "nested/deep.md", "docs", ".", pageResolvers, true, "docs/nested/deep.md"},
		{"nested/deep.md from guide", "nested/deep.md", "docs", "guide", pageResolvers, true, "docs/nested/deep.md"},

		{"assets/pic.png", "assets/pic.png", "docs", ".", pageResolvers, true, "docs/assets/pic.png"},
		{"images/logo.svg", "images/logo.svg", "", ".", pageResolvers, true, "images/logo.svg"},
		{"file.txt", "file.txt", "subdir", ".", pageResolvers, true, "subdir/file.txt"},
		{"../file.txt", "../file.txt", "docs/nested", ".", pageResolvers, true, "docs/file.txt"},

		{"/assets/style.css", "/assets/style.css", "", ".", pageResolvers, true, "assets/style.css"},
		{"/images/pic.png", "/images/pic.png", "ignored", ".", pageResolvers, true, "images/pic.png"},

		{"/docs/guide.md?v=1", "/docs/guide.md?v=1", "", ".", pageResolvers, false, "guide/"},
		{"/docs/guide.md#section", "/docs/guide.md#section", "", ".", pageResolvers, false, "guide/"},
		{"/docs/guide.md?v=1#section", "/docs/guide.md?v=1#section", "", ".", pageResolvers, false, "guide/"},
		{"asd.md?query=value", "asd.md?query=value", "docs", ".", pageResolvers, false, "asd/"},
		{"asd.md#anchor from guide", "asd.md#anchor", "docs", "guide", pageResolvers, false, "../asd/"},
		{"assets/file.png?v=2", "assets/file.png?v=2", "docs", ".", pageResolvers, true, "docs/assets/file.png"},

		{"empty", "", "", ".", pageResolvers, false, ""},
		{"empty with docs", "", "docs", ".", pageResolvers, false, ""},
		{".", ".", "", ".", pageResolvers, true, "."},
		{".with docs", ".", "docs", ".", pageResolvers, true, "docs"},
		{"./ with docs", "./", "docs", ".", pageResolvers, true, "docs"},

		{"./guide.md", "./guide.md", "docs", ".", pageResolvers, false, "guide/"},
		{"./asd.md from guide", "./asd.md", "docs", "guide", pageResolvers, false, "../asd/"},
		{"docs/./index.md", "docs/./index.md", "", ".", pageResolvers, false, "./"},
		{"docs//guide.md", "docs//guide.md", "", ".", pageResolvers, false, "guide/"},

		{"guide.md from page", "guide.md", "docs", "page", pageResolvers, false, "../guide/"},
		{"page.md from guide", "page.md", "docs", "guide", pageResolvers, false, "../page/"},
		{"index.md from nested/deep", "index.md", "docs", "nested/deep", pageResolvers, false, "../"},
		{"nested/deep.md from asd", "nested/deep.md", "docs", "asd", pageResolvers, true, "docs/nested/deep.md"},
		{"nested/deep.md from nested/deep", "nested/deep.md", "docs", "nested/deep", pageResolvers, true, "docs/nested/deep.md"},

		{"any/link.md with nil pages", "any/link.md", "docs", ".", nil, true, "docs/any/link.md"},
		{"/absolute/link.md with nil pages", "/absolute/link.md", "", ".", nil, true, "absolute/link.md"},

		{"any/link.md with empty pages", "any/link.md", "docs", ".", []*projectPageResolver{}, true, "docs/any/link.md"},
	}

	for _, tt := range args {
		t.Run(tt.name, func(t *testing.T) {
			proj := &Project{subdir: tt.subdir, pageResolvers: tt.pages}
			gh, rv, err := proj.handleLinkUrl(tt.link, tt.currentPage)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gh != tt.expGH {
				t.Errorf("gh mismatch: got %v, want %v", gh, tt.expGH)
			}
			if got := filepath.ToSlash(rv); got != tt.expRV {
				t.Errorf("bad rv: got %q, want %q", got, tt.expRV)
			}
		})
	}
}
