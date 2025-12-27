package project

import (
	"path/filepath"
	"testing"
)

func TestHandleImageUrl(t *testing.T) {
	args := []struct {
		img         string
		subdir      string
		currentPage string
		rv1         string
		rv2         string
		hasErr      bool
	}{
		{"", "", "", "", ".", false},
		{"", "subdir", "", "subdir", "subdir", false},

		{"image.png", "", "", "image.png", "image.png", false},
		{"./image.png", "", "", "image.png", "image.png", false},
		{"folder/image.png", "", "", "folder/image.png", "folder/image.png", false},
		{"./folder/image.png", "", "", "folder/image.png", "folder/image.png", false},
		{"../image.png", "", "", "", "", true},
		{"../../folder/image.png", "", "", "", "", true},

		{"image.png", "sub", "", "sub/image.png", "sub/image.png", false},
		{"./image.png", "sub", "", "sub/image.png", "sub/image.png", false},
		{"folder/image.png", "sub", "", "sub/folder/image.png", "sub/folder/image.png", false},
		{"./folder/image.png", "sub", "", "sub/folder/image.png", "sub/folder/image.png", false},
		{"../image.png", "sub", "", "image.png", "image.png", false},
		{"../../image.png", "sub", "", "", "", true},
		{"../../../image.png", "sub", "", "", "", true},

		{"image.png", "sub/nested", "", "sub/nested/image.png", "sub/nested/image.png", false},
		{"./image.png", "sub/nested", "", "sub/nested/image.png", "sub/nested/image.png", false},
		{"folder/image.png", "sub/nested", "", "sub/nested/folder/image.png", "sub/nested/folder/image.png", false},
		{"../image.png", "sub/nested", "", "sub/image.png", "sub/image.png", false},
		{"../../image.png", "sub/nested", "", "image.png", "image.png", false},
		{"../../../image.png", "sub/nested", "", "", "", true},

		{"/image.png", "", "", "image.png", "image.png", false},
		{"/folder/image.png", "", "", "folder/image.png", "folder/image.png", false},
		{"/image.png", "sub", "", "image.png", "image.png", false},
		{"/folder/image.png", "sub", "", "folder/image.png", "folder/image.png", false},
		{"/folder/image.png", "sub/nested", "", "folder/image.png", "folder/image.png", false},

		{"http://example.com/image.png", "", "", "", "", false},
		{"https://example.com/image.png", "", "", "", "", false},
		{"ftp://example.com/image.png", "", "", "", "", false},
		{"//example.com/image.png", "", "", "", "", false},
		{"http://example.com/image.png", "sub", "", "", "", false},
		{"https://example.com/folder/image.png", "sub/nested", "", "", "", false},

		{"image.png?v=1", "", "", "image.png", "image.png", false},
		{"image.png#section", "", "", "image.png", "image.png", false},
		{"image.png?v=1#section", "", "", "image.png", "image.png", false},
		{"folder/image.png?v=1", "sub", "", "sub/folder/image.png", "sub/folder/image.png", false},
		{"/image.png?v=1", "sub", "", "image.png", "image.png", false},

		{"image.png", "", "p", "image.png", "../image.png", false},
		{"image.png", "sub", "p", "sub/image.png", "../sub/image.png", false},
		{"/image.png", "", "p", "image.png", "../image.png", false},
		{"image.png", "p", "p", "p/image.png", "image.png", false},
		{"../image.png", "p/sub", "p", "p/image.png", "image.png", false},

		{"image.png", "", "p/q", "image.png", "../../image.png", false},
		{"image.png", "sub", "p/q", "sub/image.png", "../../sub/image.png", false},
		{"image.png", "p", "p/q", "p/image.png", "../image.png", false},
		{"image.png", "p/q", "p/q", "p/q/image.png", "image.png", false},
		{"image.png", "p/q/r", "p/q", "p/q/r/image.png", "r/image.png", false},

		{"image.png", "", "a/b/c", "image.png", "../../../image.png", false},
		{"image.png", "a", "a/b/c", "a/image.png", "../../image.png", false},
		{"image.png", "a/b", "a/b/c", "a/b/image.png", "../image.png", false},
		{"image.png", "a/b/c", "a/b/c", "a/b/c/image.png", "image.png", false},

		{"../image.png", "sub", "p", "image.png", "../image.png", false},
		{"../../image.png", "p/sub", "p", "image.png", "../image.png", false},
		{"../folder/image.png", "sub/nested", "p/q", "sub/folder/image.png", "../../sub/folder/image.png", false},

		{"/image.png", "", "a/b", "image.png", "../../image.png", false},
		{"/folder/image.png", "", "a/b/c", "folder/image.png", "../../../folder/image.png", false},
		{"/folder/image.png", "sub", "a/b/c", "folder/image.png", "../../../folder/image.png", false},
	}

	for i, tt := range args {
		proj := &Project{subdir: tt.subdir}
		rv1, rv2, err := proj.handleImageUrl(tt.img, tt.currentPage)
		if tt.hasErr {
			if err == nil {
				t.Errorf("%d: expected error for img=%q, subdir=%q, currentPage=%q, but got none", i, tt.img, tt.subdir, tt.currentPage)
			}
			continue
		}
		if err != nil {
			t.Errorf("%d: unexpected error for img=%q, subdir=%q, currentPage=%q: %v", i, tt.img, tt.subdir, tt.currentPage, err)
		}
		if rv1 != tt.rv1 {
			t.Errorf("%d: bad rv1 for img=%q, subdir=%q, currentPage=%q: got %q, want %q", i, tt.img, tt.subdir, tt.currentPage, rv1, tt.rv1)
		}
		if rv2n := filepath.ToSlash(rv2); rv2n != tt.rv2 {
			t.Errorf("%d: bad rv2 for img=%q, subdir=%q, currentPage=%q: got %q, want %q", i, tt.img, tt.subdir, tt.currentPage, rv2n, tt.rv2)
		}
	}
}

func TestHandleLinkUrl(t *testing.T) {
	pages := []*ProjectPage{
		{src: "docs/index.md", name: "."},
		{src: "docs/asd.md", name: "asd"},
		{src: "docs/guide.md", name: "guide"},
		{src: "docs/page.md", name: "page"},
	}

	args := []struct {
		link        string
		subdir      string
		currentPage string
		pages       []*ProjectPage
		expGH       bool
		expRV       string
	}{
		{"https://example.com/foo", "sub", ".", pages, false, ""},
		{"http://example.com/foo", "", "", pages, false, ""},
		{"//example.com/a/b", "", "", pages, false, ""},
		{"ftp://example.com/file", "docs", ".", pages, false, ""},

		{"/assets/image.png", "sub", ".", pages, true, "assets/image.png"},
		{"/assets/image.png", "", ".", pages, true, "assets/image.png"},
		{"/assets/image.png", "ignored", ".", pages, true, "assets/image.png"},

		{"/docs/index.md", "", ".", pages, false, "./"},
		{"/docs/guide.md", "", ".", pages, false, "guide/"},
		{"/docs/asd.md", "", "asd", pages, false, "./"},
		{"/docs/asd.md", "", "guide", pages, false, "../asd/"},
		{"/docs/page.md", "", ".", pages, false, "page/"},
		{"/docs/nested/deep.md", "", ".", pages, true, "docs/nested/deep.md"},
		{"/docs/nested/deep.md", "", "asd", pages, true, "docs/nested/deep.md"},

		{"guide.md", "docs", ".", pages, false, "guide/"},
		{"asd.md", "docs", ".", pages, false, "asd/"},
		{"asd.md", "docs", "guide", pages, false, "../asd/"},
		{"page.md", "docs", ".", pages, false, "page/"},
		{"index.md", "docs", ".", pages, false, "./"},
		{"index.md", "docs", "guide", pages, false, "../"},
		{"index.md", "docs", "asd", pages, false, "../"},
		{"nested/deep.md", "docs", ".", pages, true, "docs/nested/deep.md"},
		{"nested/deep.md", "docs", "guide", pages, true, "docs/nested/deep.md"},

		{"assets/pic.png", "docs", ".", pages, true, "docs/assets/pic.png"},
		{"images/logo.svg", "", ".", pages, true, "images/logo.svg"},
		{"file.txt", "subdir", ".", pages, true, "subdir/file.txt"},
		{"../file.txt", "docs/nested", ".", pages, true, "docs/file.txt"},

		{"/assets/style.css", "", ".", pages, true, "assets/style.css"},
		{"/images/pic.png", "ignored", ".", pages, true, "images/pic.png"},

		{"/docs/guide.md?v=1", "", ".", pages, false, "guide/"},
		{"/docs/guide.md#section", "", ".", pages, false, "guide/"},
		{"/docs/guide.md?v=1#section", "", ".", pages, false, "guide/"},
		{"asd.md?query=value", "docs", ".", pages, false, "asd/"},
		{"asd.md#anchor", "docs", "guide", pages, false, "../asd/"},
		{"assets/file.png?v=2", "docs", ".", pages, true, "docs/assets/file.png"},

		{"", "", ".", pages, true, "."},
		{"", "docs", ".", pages, true, "docs"},
		{".", "", ".", pages, true, "."},
		{".", "docs", ".", pages, true, "docs"},
		{"./", "docs", ".", pages, true, "docs"},

		{"./guide.md", "docs", ".", pages, false, "guide/"},
		{"./asd.md", "docs", "guide", pages, false, "../asd/"},
		{"docs/./index.md", "", ".", pages, false, "./"},
		{"docs//guide.md", "", ".", pages, false, "guide/"},

		{"guide.md", "docs", "page", pages, false, "../guide/"},
		{"page.md", "docs", "guide", pages, false, "../page/"},
		{"index.md", "docs", "nested/deep", pages, false, "../"},
		{"nested/deep.md", "docs", "asd", pages, true, "docs/nested/deep.md"},
		{"nested/deep.md", "docs", "nested/deep", pages, true, "docs/nested/deep.md"},

		{"any/link.md", "docs", ".", nil, true, "docs/any/link.md"},
		{"/absolute/link.md", "", ".", nil, true, "absolute/link.md"},

		{"any/link.md", "docs", ".", []*ProjectPage{}, true, "docs/any/link.md"},
	}

	for i, tt := range args {
		proj := &Project{subdir: tt.subdir, pages: tt.pages}
		gh, rv, err := proj.handleLinkUrl(tt.link, tt.currentPage)
		if err != nil {
			t.Fatalf("%d: unexpected error for link=%q, subdir=%q, currentPage=%q: %v", i, tt.link, tt.subdir, tt.currentPage, err)
		}
		if gh != tt.expGH {
			t.Errorf("%d: gh mismatch for link=%q, subdir=%q, currentPage=%q: got %v, want %v", i, tt.link, tt.subdir, tt.currentPage, gh, tt.expGH)
		}
		if got := filepath.ToSlash(rv); got != tt.expRV {
			t.Errorf("%d: bad rv for link=%q, subdir=%q, currentPage=%q: got %q, want %q", i, tt.link, tt.subdir, tt.currentPage, got, tt.expRV)
		}
	}
}
