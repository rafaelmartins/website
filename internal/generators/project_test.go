package generators

import (
	"testing"
)

func TestFixSubPageHtmlLink(t *testing.T) {
	args := []struct {
		link     string
		subpage  string
		subpages []string
		rv       string
	}{
		{"", "", nil, ""},

		{"bola.md", "", nil, ""},
		{"./bola.md", "", nil, ""},
		{"../bola.md", "", nil, ""},
		{"/bola.md", "", nil, ""},
		{"guda/bola.md", "", nil, ""},
		{"./guda/bola.md", "", nil, ""},
		{"../guda/bola.md", "", nil, ""},
		{"/guda/bola.md", "", nil, ""},
		{"chunda/guda/bola.md", "", nil, ""},
		{"./chunda/guda/bola.md", "", nil, ""},
		{"../chunda/guda/bola.md", "", nil, ""},
		{"/chunda/guda/bola.md", "", nil, ""},

		{"bola.md", "", []string{"bola"}, "bola/"},
		{"./bola.md", "", []string{"bola"}, "bola/"},
		{"../bola.md", "", []string{"bola"}, ""},
		{"/bola.md", "", []string{"bola"}, "bola/"},
		{"guda/bola.md", "", []string{"bola"}, ""},
		{"./guda/bola.md", "", []string{"bola"}, ""},
		{"../guda/bola.md", "", []string{"bola"}, ""},
		{"/guda/bola.md", "", []string{"bola"}, ""},
		{"chunda/guda/bola.md", "", []string{"bola"}, ""},
		{"./chunda/guda/bola.md", "", []string{"bola"}, ""},
		{"../chunda/guda/bola.md", "", []string{"bola"}, ""},
		{"/chunda/guda/bola.md", "", []string{"bola"}, ""},

		{"bola.md", "", []string{"guda/bola"}, ""},
		{"./bola.md", "", []string{"guda/bola"}, ""},
		{"../bola.md", "", []string{"guda/bola"}, ""},
		{"/bola.md", "", []string{"guda/bola"}, ""},
		{"guda/bola.md", "", []string{"guda/bola"}, "guda/bola/"},
		{"./guda/bola.md", "", []string{"guda/bola"}, "guda/bola/"},
		{"../guda/bola.md", "", []string{"guda/bola"}, ""},
		{"/guda/bola.md", "", []string{"guda/bola"}, "guda/bola/"},
		{"chunda/guda/bola.md", "", []string{"guda/bola"}, ""},
		{"./chunda/guda/bola.md", "", []string{"guda/bola"}, ""},
		{"../chunda/guda/bola.md", "", []string{"guda/bola"}, ""},
		{"/chunda/guda/bola.md", "", []string{"guda/bola"}, ""},

		{"bola.md", "", []string{"chunda/guda/bola"}, ""},
		{"./bola.md", "", []string{"chunda/guda/bola"}, ""},
		{"../bola.md", "", []string{"chunda/guda/bola"}, ""},
		{"/bola.md", "", []string{"chunda/guda/bola"}, ""},
		{"guda/bola.md", "", []string{"chunda/guda/bola"}, ""},
		{"./guda/bola.md", "", []string{"chunda/guda/bola"}, ""},
		{"../guda/bola.md", "", []string{"chunda/guda/bola"}, ""},
		{"/guda/bola.md", "", []string{"chunda/guda/bola"}, ""},
		{"chunda/guda/bola.md", "", []string{"chunda/guda/bola"}, "chunda/guda/bola/"},
		{"./chunda/guda/bola.md", "", []string{"chunda/guda/bola"}, "chunda/guda/bola/"},
		{"../chunda/guda/bola.md", "", []string{"chunda/guda/bola"}, ""},
		{"/chunda/guda/bola.md", "", []string{"chunda/guda/bola"}, "chunda/guda/bola/"},

		{"bola.md", "guda", nil, ""},
		{"./bola.md", "guda", nil, ""},
		{"../bola.md", "guda", nil, ""},
		{"/bola.md", "guda", nil, ""},
		{"guda/bola.md", "guda", nil, ""},
		{"./guda/bola.md", "guda", nil, ""},
		{"../guda/bola.md", "guda", nil, ""},
		{"/guda/bola.md", "guda", nil, ""},
		{"chunda/guda/bola.md", "guda", nil, ""},
		{"./chunda/guda/bola.md", "guda", nil, ""},
		{"../chunda/guda/bola.md", "guda", nil, ""},
		{"/chunda/guda/bola.md", "guda", nil, ""},

		{"bola.md", "guda", []string{"bola"}, "../bola/"},
		{"./bola.md", "guda", []string{"bola"}, "../bola/"},
		{"../bola.md", "guda", []string{"bola"}, ""},
		{"/bola.md", "guda", []string{"bola"}, "../bola/"},
		{"guda/bola.md", "guda", []string{"bola"}, ""},
		{"./guda/bola.md", "guda", []string{"bola"}, ""},
		{"../guda/bola.md", "guda", []string{"bola"}, ""},
		{"/guda/bola.md", "guda", []string{"bola"}, ""},
		{"chunda/guda/bola.md", "guda", []string{"bola"}, ""},
		{"./chunda/guda/bola.md", "guda", []string{"bola"}, ""},
		{"../chunda/guda/bola.md", "guda", []string{"bola"}, ""},
		{"/chunda/guda/bola.md", "guda", []string{"bola"}, ""},

		{"bola.md", "guda", []string{"guda/bola"}, ""},
		{"./bola.md", "guda", []string{"guda/bola"}, ""},
		{"../bola.md", "guda", []string{"guda/bola"}, ""},
		{"/bola.md", "guda", []string{"guda/bola"}, ""},
		{"guda/bola.md", "guda", []string{"guda/bola"}, "bola/"},
		{"./guda/bola.md", "guda", []string{"guda/bola"}, "bola/"},
		{"../guda/bola.md", "guda", []string{"guda/bola"}, ""},
		{"/guda/bola.md", "guda", []string{"guda/bola"}, "bola/"},
		{"chunda/guda/bola.md", "guda", []string{"guda/bola"}, ""},
		{"./chunda/guda/bola.md", "guda", []string{"guda/bola"}, ""},
		{"../chunda/guda/bola.md", "guda", []string{"guda/bola"}, ""},
		{"/chunda/guda/bola.md", "guda", []string{"guda/bola"}, ""},

		{"bola.md", "guda", []string{"chunda/guda/bola"}, ""},
		{"./bola.md", "guda", []string{"chunda/guda/bola"}, ""},
		{"../bola.md", "guda", []string{"chunda/guda/bola"}, ""},
		{"/bola.md", "guda", []string{"chunda/guda/bola"}, ""},
		{"guda/bola.md", "guda", []string{"chunda/guda/bola"}, ""},
		{"./guda/bola.md", "guda", []string{"chunda/guda/bola"}, ""},
		{"../guda/bola.md", "guda", []string{"chunda/guda/bola"}, ""},
		{"/guda/bola.md", "guda", []string{"chunda/guda/bola"}, ""},
		{"chunda/guda/bola.md", "guda", []string{"chunda/guda/bola"}, "../chunda/guda/bola/"},
		{"./chunda/guda/bola.md", "guda", []string{"chunda/guda/bola"}, "../chunda/guda/bola/"},
		{"../chunda/guda/bola.md", "guda", []string{"chunda/guda/bola"}, ""},
		{"/chunda/guda/bola.md", "guda", []string{"chunda/guda/bola"}, "../chunda/guda/bola/"},

		{"bola.md", "chunda", nil, ""},
		{"./bola.md", "chunda", nil, ""},
		{"../bola.md", "chunda", nil, ""},
		{"/bola.md", "chunda", nil, ""},
		{"guda/bola.md", "chunda", nil, ""},
		{"./guda/bola.md", "chunda", nil, ""},
		{"../guda/bola.md", "chunda", nil, ""},
		{"/guda/bola.md", "chunda", nil, ""},
		{"chunda/guda/bola.md", "chunda", nil, ""},
		{"./chunda/guda/bola.md", "chunda", nil, ""},
		{"../chunda/guda/bola.md", "chunda", nil, ""},
		{"/chunda/guda/bola.md", "chunda", nil, ""},

		{"bola.md", "chunda", []string{"bola"}, "../bola/"},
		{"./bola.md", "chunda", []string{"bola"}, "../bola/"},
		{"../bola.md", "chunda", []string{"bola"}, ""},
		{"/bola.md", "chunda", []string{"bola"}, "../bola/"},
		{"guda/bola.md", "chunda", []string{"bola"}, ""},
		{"./guda/bola.md", "chunda", []string{"bola"}, ""},
		{"../guda/bola.md", "chunda", []string{"bola"}, ""},
		{"/guda/bola.md", "chunda", []string{"bola"}, ""},
		{"chunda/guda/bola.md", "chunda", []string{"bola"}, ""},
		{"./chunda/guda/bola.md", "chunda", []string{"bola"}, ""},
		{"../chunda/guda/bola.md", "chunda", []string{"bola"}, ""},
		{"/chunda/guda/bola.md", "chunda", []string{"bola"}, ""},

		{"bola.md", "chunda", []string{"guda/bola"}, ""},
		{"./bola.md", "chunda", []string{"guda/bola"}, ""},
		{"../bola.md", "chunda", []string{"guda/bola"}, ""},
		{"/bola.md", "chunda", []string{"guda/bola"}, ""},
		{"guda/bola.md", "chunda", []string{"guda/bola"}, "../guda/bola/"},
		{"./guda/bola.md", "chunda", []string{"guda/bola"}, "../guda/bola/"},
		{"../guda/bola.md", "chunda", []string{"guda/bola"}, ""},
		{"/guda/bola.md", "chunda", []string{"guda/bola"}, "../guda/bola/"},
		{"chunda/guda/bola.md", "chunda", []string{"guda/bola"}, ""},
		{"./chunda/guda/bola.md", "chunda", []string{"guda/bola"}, ""},
		{"../chunda/guda/bola.md", "chunda", []string{"guda/bola"}, ""},
		{"/chunda/guda/bola.md", "chunda", []string{"guda/bola"}, ""},

		{"bola.md", "chunda", []string{"chunda/guda/bola"}, ""},
		{"./bola.md", "chunda", []string{"chunda/guda/bola"}, ""},
		{"../bola.md", "chunda", []string{"chunda/guda/bola"}, ""},
		{"/bola.md", "chunda", []string{"chunda/guda/bola"}, ""},
		{"guda/bola.md", "chunda", []string{"chunda/guda/bola"}, ""},
		{"./guda/bola.md", "chunda", []string{"chunda/guda/bola"}, ""},
		{"../guda/bola.md", "chunda", []string{"chunda/guda/bola"}, ""},
		{"/guda/bola.md", "chunda", []string{"chunda/guda/bola"}, ""},
		{"chunda/guda/bola.md", "chunda", []string{"chunda/guda/bola"}, "guda/bola/"},
		{"./chunda/guda/bola.md", "chunda", []string{"chunda/guda/bola"}, "guda/bola/"},
		{"../chunda/guda/bola.md", "chunda", []string{"chunda/guda/bola"}, ""},
		{"/chunda/guda/bola.md", "chunda", []string{"chunda/guda/bola"}, "guda/bola/"},

		{"bola.md", "chunda/guda", nil, ""},
		{"./bola.md", "chunda/guda", nil, ""},
		{"../bola.md", "chunda/guda", nil, ""},
		{"/bola.md", "chunda/guda", nil, ""},
		{"guda/bola.md", "chunda/guda", nil, ""},
		{"./guda/bola.md", "chunda/guda", nil, ""},
		{"../guda/bola.md", "chunda/guda", nil, ""},
		{"/guda/bola.md", "chunda/guda", nil, ""},
		{"chunda/guda/bola.md", "chunda/guda", nil, ""},
		{"./chunda/guda/bola.md", "chunda/guda", nil, ""},
		{"../chunda/guda/bola.md", "chunda/guda", nil, ""},
		{"/chunda/guda/bola.md", "chunda/guda", nil, ""},

		{"bola.md", "chunda/guda", []string{"bola"}, ""},
		{"./bola.md", "chunda/guda", []string{"bola"}, ""},
		{"../bola.md", "chunda/guda", []string{"bola"}, "../../bola/"},
		{"/bola.md", "chunda/guda", []string{"bola"}, "../../bola/"},
		{"guda/bola.md", "chunda/guda", []string{"bola"}, ""},
		{"./guda/bola.md", "chunda/guda", []string{"bola"}, ""},
		{"../guda/bola.md", "chunda/guda", []string{"bola"}, ""},
		{"/guda/bola.md", "chunda/guda", []string{"bola"}, ""},
		{"chunda/guda/bola.md", "chunda/guda", []string{"bola"}, ""},
		{"./chunda/guda/bola.md", "chunda/guda", []string{"bola"}, ""},
		{"../chunda/guda/bola.md", "chunda/guda", []string{"bola"}, ""},
		{"/chunda/guda/bola.md", "chunda/guda", []string{"bola"}, ""},

		{"bola.md", "chunda/guda", []string{"guda/bola"}, ""},
		{"./bola.md", "chunda/guda", []string{"guda/bola"}, ""},
		{"../bola.md", "chunda/guda", []string{"guda/bola"}, ""},
		{"/bola.md", "chunda/guda", []string{"guda/bola"}, ""},
		{"guda/bola.md", "chunda/guda", []string{"guda/bola"}, ""},
		{"./guda/bola.md", "chunda/guda", []string{"guda/bola"}, ""},
		{"../guda/bola.md", "chunda/guda", []string{"guda/bola"}, "../../guda/bola/"},
		{"/guda/bola.md", "chunda/guda", []string{"guda/bola"}, "../../guda/bola/"},
		{"chunda/guda/bola.md", "chunda/guda", []string{"guda/bola"}, ""},
		{"./chunda/guda/bola.md", "chunda/guda", []string{"guda/bola"}, ""},
		{"../chunda/guda/bola.md", "chunda/guda", []string{"guda/bola"}, ""},
		{"/chunda/guda/bola.md", "chunda/guda", []string{"guda/bola"}, ""},

		{"bola.md", "chunda/guda", []string{"chunda/guda/bola"}, ""}, //
		{"./bola.md", "chunda/guda", []string{"chunda/guda/bola"}, ""},
		{"../bola.md", "chunda/guda", []string{"chunda/guda/bola"}, ""},
		{"/bola.md", "chunda/guda", []string{"chunda/guda/bola"}, ""},
		{"guda/bola.md", "chunda/guda", []string{"chunda/guda/bola"}, "bola/"},
		{"./guda/bola.md", "chunda/guda", []string{"chunda/guda/bola"}, "bola/"},
		{"../guda/bola.md", "chunda/guda", []string{"chunda/guda/bola"}, ""},
		{"/guda/bola.md", "chunda/guda", []string{"chunda/guda/bola"}, ""},
		{"chunda/guda/bola.md", "chunda/guda", []string{"chunda/guda/bola"}, ""},
		{"./chunda/guda/bola.md", "chunda/guda", []string{"chunda/guda/bola"}, ""},
		{"../chunda/guda/bola.md", "chunda/guda", []string{"chunda/guda/bola"}, "bola/"},
		{"/chunda/guda/bola.md", "chunda/guda", []string{"chunda/guda/bola"}, "bola/"},

		{"../test.md", "bola/foo", []string{"test", "bola/foo"}, "../../test/"},
		{"/test.md", "bola/foo", []string{"test", "bola/foo"}, "../../test/"},
	}

	for i, tt := range args {
		if rv := fixSubPageHtmlLink(tt.link, tt.subpage, tt.subpages); rv != tt.rv {
			t.Errorf("%d: bad rv: got %q, want %q", i, rv, tt.rv)
		}
	}
}

func TestFixHtmlImg(t *testing.T) {
	args := []struct {
		img     string
		subpage string
		rv      string
	}{
		{"", "", ""},

		{"bola.png", "", "images/bola.png"},
		{"./bola.png", "", "images/bola.png"},
		{"../bola.png", "", ""},
		{"/bola.png", "", "images/bola.png"},
		{"guda/bola.png", "", "images/guda/bola.png"},
		{"./guda/bola.png", "", "images/guda/bola.png"},
		{"../guda/bola.png", "", ""},
		{"/guda/bola.png", "", "images/guda/bola.png"},
		{"chunda/guda/bola.png", "", "images/chunda/guda/bola.png"},
		{"./chunda/guda/bola.png", "", "images/chunda/guda/bola.png"},
		{"../chunda/guda/bola.png", "", ""},
		{"/chunda/guda/bola.png", "", "images/chunda/guda/bola.png"},

		{"bola.png", "guda", "images/bola.png"},
		{"./bola.png", "guda", "images/bola.png"},
		{"../bola.png", "guda", ""},
		{"/bola.png", "guda", "images/bola.png"},
		{"guda/bola.png", "guda", "images/guda/bola.png"},
		{"./guda/bola.png", "guda", "images/guda/bola.png"},
		{"../guda/bola.png", "guda", ""},
		{"/guda/bola.png", "guda", "images/guda/bola.png"},
		{"chunda/guda/bola.png", "guda", "images/chunda/guda/bola.png"},
		{"./chunda/guda/bola.png", "guda", "images/chunda/guda/bola.png"},
		{"../chunda/guda/bola.png", "guda", ""},
		{"/chunda/guda/bola.png", "guda", "images/chunda/guda/bola.png"},

		{"bola.png", "chunda", "images/bola.png"},
		{"./bola.png", "chunda", "images/bola.png"},
		{"../bola.png", "chunda", ""},
		{"/bola.png", "chunda", "images/bola.png"},
		{"guda/bola.png", "chunda", "images/guda/bola.png"},
		{"./guda/bola.png", "chunda", "images/guda/bola.png"},
		{"../guda/bola.png", "chunda", ""},
		{"/guda/bola.png", "chunda", "images/guda/bola.png"},
		{"chunda/guda/bola.png", "chunda", "images/chunda/guda/bola.png"},
		{"./chunda/guda/bola.png", "chunda", "images/chunda/guda/bola.png"},
		{"../chunda/guda/bola.png", "chunda", ""},
		{"/chunda/guda/bola.png", "chunda", "images/chunda/guda/bola.png"},

		{"bola.png", "chunda/guda", "images/chunda/bola.png"},
		{"./bola.png", "chunda/guda", "images/chunda/bola.png"},
		{"../bola.png", "chunda/guda", "images/bola.png"},
		{"/bola.png", "chunda/guda", "images/bola.png"},
		{"guda/bola.png", "chunda/guda", "images/chunda/guda/bola.png"},
		{"./guda/bola.png", "chunda/guda", "images/chunda/guda/bola.png"},
		{"../guda/bola.png", "chunda/guda", "images/guda/bola.png"},
		{"/guda/bola.png", "chunda/guda", "images/guda/bola.png"},
		{"chunda/guda/bola.png", "chunda/guda", "images/chunda/chunda/guda/bola.png"},
		{"./chunda/guda/bola.png", "chunda/guda", "images/chunda/chunda/guda/bola.png"},
		{"../chunda/guda/bola.png", "chunda/guda", "images/chunda/guda/bola.png"},
		{"/chunda/guda/bola.png", "chunda/guda", "images/chunda/guda/bola.png"},

		{"../foo.png", "bola/foo", "images/foo.png"},
		{"/foo.png", "bola/foo", "images/foo.png"},
	}

	for i, tt := range args {
		if rv := fixSubPageHtmlImg(tt.img, tt.subpage); rv != tt.rv {
			t.Errorf("%d: bad rv: got %q, want %q", i, rv, tt.rv)
		}
	}
}

func TestHandleHideComments(t *testing.T) {
	args := []struct {
		mkd string
		rv  string
	}{
		{"# Foo\n", "# Foo\n"},
		{"# Foo\n\n<!-- website-hide -->qweqwe<!-- /website-hide -->\n", "# Foo\n\n\n"},

		{"<!-- website-hide -->hidden content<!-- /website-hide -->\n# Foo\n", "\n# Foo\n"},
		{"<!-- website-hide -->multiple\nlines\nhidden<!-- /website-hide --># Visible", "# Visible"},

		{"# Foo\n\nVisible content\n<!-- website-hide -->hidden at end<!-- /website-hide -->", "# Foo\n\nVisible content\n"},

		{"# Foo\n<!-- website-hide -->hide1<!-- /website-hide -->\nVisible\n<!-- website-hide -->hide2<!-- /website-hide -->\nEnd", "# Foo\n\nVisible\n\nEnd"},
		{"<!-- website-hide -->start<!-- /website-hide -->middle<!-- website-hide -->end<!-- /website-hide -->", "middle"},

		{"# Title\n\n<!-- website-hide -->\n## Hidden Section\n\nHidden paragraph\n\n<!-- /website-hide -->\n\n## Visible Section", "# Title\n\n\n\n## Visible Section"},

		{"# Foo\n<!-- website-hide --><!-- /website-hide -->\n# Bar", "# Foo\n\n# Bar"},
		{"<!-- website-hide -->\n\n<!-- /website-hide -->Content", "Content"},

		{"# Foo\n<!--   website-hide   -->content<!--   /website-hide   -->\n", "# Foo\n\n"},
		{"Pre<!--website-hide\n-->\n  spaced content  \n<!-- /website-hide -->Post", "PrePost"},

		{"Start<!-- website-hide -->hide1<!-- /website-hide --><!-- website-hide -->hide2<!-- /website-hide -->End", "StartEnd"},

		{"# Test\n<!-- website-hide -->Special: !@#$%^&*()<!-- /website-hide -->\nNormal", "# Test\n\nNormal"},

		{"<!-- website-hide -->Everything is hidden<!-- /website-hide -->", ""},

		{"Start\n<!-- website-hide -->hide1<!-- /website-hide -->\nMiddle\n<!-- website-hide -->hide2<!-- /website-hide -->\nEnd", "Start\n\nMiddle\n\nEnd"},

		{"# Foo\n<!-- website-hide -->unclosed\n# Bar", "# Foo\n"},
		{"# Foo\nunopened<!-- /website-hide -->\n# Bar", "# Foo\nunopened\n# Bar"},

		{"# Foo\n<!-- website-hide -->\n<!-- website-hide -->nested<!-- /website-hide -->\n<!-- /website-hide -->\n# Bar", "# Foo\n\n\n# Bar"},
	}

	for i, tt := range args {
		if rv := handleHideComments(tt.mkd); rv != tt.rv {
			t.Errorf("%d: bad rv: got %q, want %q", i, rv, tt.rv)
		}
	}
}
