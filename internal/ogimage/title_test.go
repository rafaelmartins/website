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
		title  string
		lines  []string
		height int
		err    error
	}{
		{"", []string{}, 0, nil},
		{"bola", []string{"bola"}, 95, nil},

		{"bola guda", []string{"bola guda"}, 95, nil},
		{"bola gudaaaaaaaaa aa", []string{"bola", "gudaaaaaaaaa aa"}, 202, nil},
		{"bola gudaaaaaaaaaaa", []string{"bola", "gudaaaaaaaaaaa"}, 202, nil},
		{"bola gudaaaaaaaaa aaaaaaa aaaaaa 1234", []string{"bola", "gudaaaaaaaaa", "aaaaaaa aaaaaa", "1234"}, 416, nil},
		{"bola gudaaaaaaaaa aaaaaaaaaaaaaa aaaaaaaaaaaaaa", []string{"bola", "gudaaaaaaaaa", "aaaaaaaaaaaaaa", "aaaaaaaaaaaaaa"}, 416, nil},
		{"bola gudaaaaaaaaa aaaaaaaaaaaaaaaaaa 1234", []string{}, 0, errTitleTooLongWidth},
		{"bola gudaaaaaaaaa aaaaaaaaaaaaaa aaaaaaaaaaaaaa aaaaaaaaaaaaaa aaaaaaaaaaaaaa", []string{}, 0, errTitleTooLongHeight},
	}

	for i, tt := range args {
		lines, height, err := titleSplit(tt.title, face, image.Rect(50, 20, 910, 510))
		if err != tt.err {
			t.Errorf("%d: bad error: got %q, want %q", i, err, tt.err)
		}
		if slices.Compare(lines, tt.lines) != 0 {
			t.Errorf("%d: bad lines: got %q, want %q", i, lines, tt.lines)
		}
		if height != tt.height {
			t.Errorf("%d: bad height: got %d, want %d", i, height, tt.height)
		}
	}
}
