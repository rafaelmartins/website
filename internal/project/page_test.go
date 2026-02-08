package project

import (
	"testing"
)

func TestSplitFileName(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		isReadme bool
		isRoot   bool
		wantIdx  int
		wantName string
		wantErr  bool
	}{
		{"readme returns zero values", "README.md", true, false, 0, "", false},
		{"readme with isRoot", "README.md", true, true, 0, "", false},
		{"root page", "10_about.md", false, true, 10, "", false},
		{"non-root page", "10_about.md", false, false, 10, "about", false},
		{"non-root page with extension stripped", "20_contact.html", false, false, 20, "contact", false},
		{"index zero non-root errors", "0_index.md", false, false, 0, "", true},
		{"index zero root", "0_index.md", false, true, 0, "", false},
		{"no underscore", "badname.md", false, false, 0, "", true},
		{"non-numeric prefix", "abc_page.md", false, false, 0, "", true},
		{"nested path non-root", "docs/10_getting-started.md", false, false, 10, "getting-started", false},
		{"nested path root", "docs/10_getting-started.md", false, true, 10, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, name, err := splitFileName(tt.fileName, tt.isReadme, tt.isRoot)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if idx != tt.wantIdx {
				t.Errorf("idx = %d, want %d", idx, tt.wantIdx)
			}
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
		})
	}
}

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
			ppr := &projectPageResolver{name: tt.ppName}
			got := ppr.resolveUrl(tt.current)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
