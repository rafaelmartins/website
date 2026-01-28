package ogimage

import (
	"image"
	"slices"
	"testing"

	"rafaelmartins.com/p/website/internal/ogfont"
)

func TestTitleSplit(t *testing.T) {
	fnt, err := ogfont.New()
	if err != nil {
		t.Fatal(err)
	}

	face, err := fnt.GetFace(96, 72)
	if err != nil {
		t.Fatal(err)
	}
	defer face.Close()

	args := []struct {
		name   string
		title  string
		lines  []string
		height int
		err    error
	}{
		{"empty", "", []string{}, 0, nil},
		{"bola", "bola", []string{"bola"}, 95, nil},

		{"bola guda", "bola guda", []string{"bola guda"}, 95, nil},
		{"bola gudaaaaaaaaa aa", "bola gudaaaaaaaaa aa", []string{"bola", "gudaaaaaaaaa aa"}, 202, nil},
		{"bola gudaaaaaaaaaaa", "bola gudaaaaaaaaaaa", []string{"bola", "gudaaaaaaaaaaa"}, 202, nil},
		{"long multi-line", "bola gudaaaaaaaaa aaaaaaa aaaaaa 1234", []string{"bola", "gudaaaaaaaaa", "aaaaaaa aaaaaa", "1234"}, 416, nil},
		{"multi-line equal", "bola gudaaaaaaaaa aaaaaaaaaaaaaa aaaaaaaaaaaaaa", []string{"bola", "gudaaaaaaaaa", "aaaaaaaaaaaaaa", "aaaaaaaaaaaaaa"}, 416, nil},
		{"too long width", "bola gudaaaaaaaaa aaaaaaaaaaaaaaaaaa 1234", []string{}, 0, errTitleTooLongWidth},
		{"too long height", "bola gudaaaaaaaaa aaaaaaaaaaaaaa aaaaaaaaaaaaaa aaaaaaaaaaaaaa aaaaaaaaaaaaaa", []string{}, 0, errTitleTooLongHeight},
	}

	for _, tt := range args {
		t.Run(tt.name, func(t *testing.T) {
			lines, height, err := titleSplit(tt.title, face, image.Rect(50, 20, 910, 510))
			if err != tt.err {
				t.Errorf("bad error: got %q, want %q", err, tt.err)
			}
			if slices.Compare(lines, tt.lines) != 0 {
				t.Errorf("bad lines: got %q, want %q", lines, tt.lines)
			}
			if height != tt.height {
				t.Errorf("bad height: got %d, want %d", height, tt.height)
			}
		})
	}
}
