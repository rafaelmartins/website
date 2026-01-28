package project

import (
	"testing"
)

func TestResolveUrl(t *testing.T) {
	tests := []struct {
		name    string
		ppName  string
		current string
		want    string
	}{
		{"empty ppName with dot", "", ".", "./"},
		{"empty ppName with empty", "", "", "./"},
		{"dot ppName with dot", ".", ".", "./"},
		{"foo ppName with foo", "foo", "foo", "./"},
		{"about ppName with dot", "about", ".", "about/"},
		{"contact ppName with dot", "contact", ".", "contact/"},
		{"about ppName with contact", "about", "contact", "../about/"},
		{"dot ppName with about", ".", "about", "../"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := &ProjectPage{name: tt.ppName}
			got := pp.resolveUrl(tt.current)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
