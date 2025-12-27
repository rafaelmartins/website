package project

import (
	"testing"
)

func TestResolveUrl(t *testing.T) {
	tests := []struct {
		ppName  string
		current string
		want    string
	}{
		{"", ".", "./"},
		{"", "", "./"},
		{".", ".", "./"},
		{"foo", "foo", "./"},
		{"about", ".", "about/"},
		{"contact", ".", "contact/"},
		{"about", "contact", "../about/"},
		{".", "about", "../"},
	}

	for _, tt := range tests {
		pp := &ProjectPage{name: tt.ppName}
		got := pp.resolveUrl(tt.current)
		if got != tt.want {
			t.Errorf("resolveUrl(%q) with name=%q: got %q, want %q", tt.current, tt.ppName, got, tt.want)
		}
	}
}
