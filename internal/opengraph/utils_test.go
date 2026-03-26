package opengraph

import (
	"image"
	"slices"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

func TestUrl(t *testing.T) {
	args := []struct {
		name    string
		baseurl string
		want    string
	}{
		{"empty", "", ""},
		{"root trailing slash", "/", "/opengraph.png"},
		{"blog post trailing slash", "/blog/hello-world/", "/blog/hello-world/opengraph.png"},
		{"project path", "/p/website/", "/p/website/opengraph.png"},
		{"series path", "/series/rcsid/", "/series/rcsid/opengraph.png"},
		{"index.html root", "/index.html", "/opengraph.png"},
		{"index.html subdir", "/blog/hello-world/index.html", "/blog/hello-world/opengraph.png"},
		{"bare file html", "/blog/hello-world.html", "/blog/hello-world/opengraph.png"},
		{"bare file no ext", "/blog/hello-world", "/blog/hello-world/opengraph.png"},
	}

	for _, tt := range args {
		t.Run(tt.name, func(t *testing.T) {
			got := url(tt.baseurl)
			if got != tt.want {
				t.Errorf("url(%q) = %q, want %q", tt.baseurl, got, tt.want)
			}
		})
	}
}

func TestImageTitleFaceHeight(t *testing.T) {
	fnt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    96,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer face.Close()

	h := imageTitleFaceHeight(face)
	if h <= 0 {
		t.Errorf("imageTitleFaceHeight() = %d, want > 0", h)
	}
}

func TestImageTitleFaceSpacing(t *testing.T) {
	fnt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    96,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer face.Close()

	s := imageTitleFaceSpacing(face)
	h := imageTitleFaceHeight(face)
	if s <= 0 {
		t.Errorf("imageTitleFaceSpacing() = %d, want > 0", s)
	}
	if s >= h {
		t.Errorf("imageTitleFaceSpacing() = %d, should be less than height %d", s, h)
	}
}

func TestImageTitleSplit(t *testing.T) {
	fnt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    96,
		DPI:     72,
		Hinting: font.HintingNone,
	})
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
		{"bola", "bola", []string{"bola"}, 91, nil},

		{"bola guda", "bola guda", []string{"bola guda"}, 91, nil},
		{"bola gudaaaaaaaaa aa", "bola gudaaaaaaaaa aa", []string{"bola gudaaaaaaaaa", "aa"}, 194, nil},
		{"bola gudaaaaaaaaaaa", "bola gudaaaaaaaaaaa", []string{"bola", "gudaaaaaaaaaaa"}, 194, nil},
		{"long multi-line", "bola gudaaaaaaaaa aaaaaaa aaaaaa 1234", []string{"bola gudaaaaaaaaa", "aaaaaaa aaaaaa", "1234"}, 297, nil},
		{"multi-line equal", "bola gudaaaaaaaaa aaaaaaaaaaaaaa aaaaaaaaaaaaaa", []string{"bola gudaaaaaaaaa", "aaaaaaaaaaaaaa", "aaaaaaaaaaaaaa"}, 297, nil},
		{"too long width", "bola gudaaaaaaaaa aaaaaaaaaaaaaaaaaa 1234", []string{}, 0, errTitleTooLongWidth},
		{"too long height", "bola gudaaaaaaaaa aaaaaaaaaaaaaa aaaaaaaaaaaaaa aaaaaaaaaaaaaa aaaaaaaaaaaaaa", []string{}, 0, errTitleTooLongHeight},
	}

	for _, tt := range args {
		t.Run(tt.name, func(t *testing.T) {
			lines, height, err := imageTitleSplit(tt.title, face, image.Rect(50, 20, 910, 510))
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
