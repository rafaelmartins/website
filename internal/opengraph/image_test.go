package opengraph

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"slices"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

func TestNewImageGenNil(t *testing.T) {
	ogi, err := NewImageGen(nil)
	if err != nil {
		t.Fatalf("NewImageGen(nil) returned error: %v", err)
	}
	if ogi != nil {
		t.Errorf("NewImageGen(nil) = %v, want nil", ogi)
	}
}

func TestNewImageGenEmptyTemplate(t *testing.T) {
	cfg := &ImageGenConfig{}
	ogi, err := NewImageGen(cfg)
	if err != nil {
		t.Fatalf("NewImageGen() returned error: %v", err)
	}
	if ogi == nil {
		t.Fatal("NewImageGen() = nil, want non-nil")
	}

	if ogi.template != "" {
		t.Errorf("template = %q, want empty", ogi.template)
	}
	if ogi.dcolor != color.Black {
		t.Errorf("dcolor = %v, want Black", ogi.dcolor)
	}
	if ogi.ddpi != 72 {
		t.Errorf("ddpi = %f, want 72", ogi.ddpi)
	}
	if ogi.dsize != 96 {
		t.Errorf("dsize = %f, want 96", ogi.dsize)
	}
}

func TestNewImageGenEmptyTemplateIgnoresCustomDefaults(t *testing.T) {
	cfg := &ImageGenConfig{
		DefaultColor: new("#336699"),
		DefaultDPI:   new(144.0),
		DefaultSize:  new(64.0),
	}

	ogi, err := NewImageGen(cfg)
	if err != nil {
		t.Fatalf("NewImageGen() returned error: %v", err)
	}

	if ogi.dcolor != color.Black {
		t.Errorf("dcolor = %v, want Black", ogi.dcolor)
	}
	if ogi.ddpi != 72 {
		t.Errorf("ddpi = %f, want 72", ogi.ddpi)
	}
	if ogi.dsize != 96 {
		t.Errorf("dsize = %f, want 96", ogi.dsize)
	}
}

func TestNewImageGenInvalidColor(t *testing.T) {
	cfg := &ImageGenConfig{
		Template:     "something",
		DefaultColor: new("invalid"),
	}

	if _, err := NewImageGen(cfg); err == nil {
		t.Error("expected error for invalid color, got nil")
	}
}

func TestGetPathsEmptyTemplate(t *testing.T) {
	ogi := &OpenGraphImageGen{template: ""}
	paths := ogi.GetPaths()
	if len(paths) != 0 {
		t.Errorf("GetPaths() = %v, want empty slice", paths)
	}
}

func TestGetPathsWithTemplate(t *testing.T) {
	ogi := &OpenGraphImageGen{template: "/some/path/template.xcf"}
	paths := ogi.GetPaths()
	want := []string{"/some/path/template.xcf"}
	if !slices.Equal(paths, want) {
		t.Errorf("GetPaths() = %v, want %v", paths, want)
	}
}

func TestGenerateNilOg(t *testing.T) {
	ogi := &OpenGraphImageGen{
		template: "test",
		image:    image.NewRGBA(image.Rect(0, 0, 1200, 630)),
	}
	if _, err := ogi.Generate(nil); err == nil {
		t.Error("expected error for nil og, got nil")
	}
}

func TestGenerateEmptyTemplate(t *testing.T) {
	ogi := &OpenGraphImageGen{
		template: "",
	}
	og := &OpenGraph{title: "Test"}
	if _, err := ogi.Generate(og); err == nil {
		t.Error("expected error for empty template, got nil")
	}
}

func TestGenerateNilImage(t *testing.T) {
	ogi := &OpenGraphImageGen{
		template: "test",
		image:    nil,
	}
	og := &OpenGraph{title: "Test"}
	if _, err := ogi.Generate(og); err == nil {
		t.Error("expected error for nil image, got nil")
	}
}

func TestGenerateInvalidWhitespace(t *testing.T) {
	fnt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 1200, 630))
	ogi := &OpenGraphImageGen{
		template: "test",
		font:     fnt,
		image:    img,
		mask:     image.Rect(50, 20, 910, 510),
	}

	args := []struct {
		name  string
		title string
	}{
		{"tab", "Hello\tWorld"},
		{"newline", "Hello\nWorld"},
		{"carriage return", "Hello\rWorld"},
	}

	for _, tt := range args {
		t.Run(tt.name, func(t *testing.T) {
			og := &OpenGraph{
				ogi:   ogi,
				title: tt.title,
				c:     color.Black,
				dpi:   72,
				size:  96,
			}
			if _, err := ogi.Generate(og); err == nil {
				t.Error("expected error for whitespace in title, got nil")
			}
		})
	}
}

func TestGenerateValidTitle(t *testing.T) {
	fnt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 1200, 630))
	ogi := &OpenGraphImageGen{
		template: "test",
		font:     fnt,
		image:    img,
		mask:     image.Rect(50, 20, 910, 510),
	}

	args := []struct {
		name  string
		title string
	}{
		{"simple", "Hello World"},
		{"single word", "Test"},
		{"blog post title", "Hello, World!"},
		{"project title", "blogc"},
		{"series title", "RCSID"},
		{"html entities", "&amp; Test &lt;"},
		{"two words", "Weekend Projects"},
	}

	for _, tt := range args {
		t.Run(tt.name, func(t *testing.T) {
			og := &OpenGraph{
				ogi:   ogi,
				title: tt.title,
				c:     color.Black,
				dpi:   72,
				size:  96,
			}

			rd, err := ogi.Generate(og)
			if err != nil {
				t.Fatalf("Generate() returned error: %v", err)
			}
			defer rd.Close()

			data, err := io.ReadAll(rd)
			if err != nil {
				t.Fatalf("ReadAll() returned error: %v", err)
			}

			if len(data) == 0 {
				t.Error("generated image is empty")
			}

			decoded, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("png.Decode() returned error: %v", err)
			}

			bounds := decoded.Bounds()
			if bounds.Dx() != 1200 || bounds.Dy() != 630 {
				t.Errorf("image size = %dx%d, want 1200x630", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestGenerateWithColor(t *testing.T) {
	fnt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 1200, 630))
	ogi := &OpenGraphImageGen{
		template: "test",
		font:     fnt,
		image:    img,
		mask:     image.Rect(50, 20, 910, 510),
	}

	og := &OpenGraph{
		ogi:   ogi,
		title: "Color Test",
		c:     color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff},
		dpi:   72,
		size:  96,
	}

	rd, err := ogi.Generate(og)
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}
	defer rd.Close()

	data, err := io.ReadAll(rd)
	if err != nil {
		t.Fatalf("ReadAll() returned error: %v", err)
	}

	ogBlack := &OpenGraph{
		ogi:   ogi,
		title: "Color Test",
		c:     color.Black,
		dpi:   72,
		size:  96,
	}

	rdBlack, err := ogi.Generate(ogBlack)
	if err != nil {
		t.Fatalf("Generate() black returned error: %v", err)
	}
	defer rdBlack.Close()

	dataBlack, err := io.ReadAll(rdBlack)
	if err != nil {
		t.Fatalf("ReadAll() black returned error: %v", err)
	}

	if bytes.Equal(data, dataBlack) {
		t.Error("red and black renders should produce different images")
	}
}

func TestGenerateWithDifferentSizes(t *testing.T) {
	fnt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 1200, 630))
	ogi := &OpenGraphImageGen{
		template: "test",
		font:     fnt,
		image:    img,
		mask:     image.Rect(50, 20, 910, 510),
	}

	rd48, err := ogi.Generate(&OpenGraph{
		ogi: ogi, title: "Size", c: color.Black, dpi: 72, size: 48,
	})
	if err != nil {
		t.Fatalf("Generate(size=48) returned error: %v", err)
	}
	data48, err := io.ReadAll(rd48)
	rd48.Close()
	if err != nil {
		t.Fatalf("ReadAll(size=48) returned error: %v", err)
	}

	rd96, err := ogi.Generate(&OpenGraph{
		ogi: ogi, title: "Size", c: color.Black, dpi: 72, size: 96,
	})
	if err != nil {
		t.Fatalf("Generate(size=96) returned error: %v", err)
	}
	data96, err := io.ReadAll(rd96)
	rd96.Close()
	if err != nil {
		t.Fatalf("ReadAll(size=96) returned error: %v", err)
	}

	if bytes.Equal(data48, data96) {
		t.Error("different font sizes should produce different images")
	}
}

func TestGenerateEmptyTitle(t *testing.T) {
	fnt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 1200, 630))
	ogi := &OpenGraphImageGen{
		template: "test",
		font:     fnt,
		image:    img,
		mask:     image.Rect(50, 20, 910, 510),
	}

	og := &OpenGraph{
		ogi:   ogi,
		title: "",
		c:     color.Black,
		dpi:   72,
		size:  96,
	}

	rd, err := ogi.Generate(og)
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}
	defer rd.Close()

	decoded, err := png.Decode(rd)
	if err != nil {
		t.Fatalf("png.Decode() returned error: %v", err)
	}

	bounds := decoded.Bounds()
	if bounds.Dx() != 1200 || bounds.Dy() != 630 {
		t.Errorf("image size = %dx%d, want 1200x630", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateWithDifferentDPI(t *testing.T) {
	fnt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 1200, 630))
	ogi := &OpenGraphImageGen{
		template: "test",
		font:     fnt,
		image:    img,
		mask:     image.Rect(50, 20, 910, 510),
	}

	rd72, err := ogi.Generate(&OpenGraph{
		ogi: ogi, title: "DPI", c: color.Black, dpi: 72, size: 48,
	})
	if err != nil {
		t.Fatalf("Generate(dpi=72) returned error: %v", err)
	}
	data72, err := io.ReadAll(rd72)
	rd72.Close()
	if err != nil {
		t.Fatalf("ReadAll(dpi=72) returned error: %v", err)
	}

	rd144, err := ogi.Generate(&OpenGraph{
		ogi: ogi, title: "DPI", c: color.Black, dpi: 144, size: 48,
	})
	if err != nil {
		t.Fatalf("Generate(dpi=144) returned error: %v", err)
	}
	data144, err := io.ReadAll(rd144)
	rd144.Close()
	if err != nil {
		t.Fatalf("ReadAll(dpi=144) returned error: %v", err)
	}

	if bytes.Equal(data72, data144) {
		t.Error("different DPI values should produce different images")
	}
}

func TestGeneratePreservesImageDimensions(t *testing.T) {
	fnt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	args := []struct {
		name string
		w, h int
	}{
		{"1200x630", 1200, 630},
		{"800x400", 800, 400},
		{"1920x1080", 1920, 1080},
	}

	for _, tt := range args {
		t.Run(tt.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, tt.w, tt.h))
			ogi := &OpenGraphImageGen{
				template: "test",
				font:     fnt,
				image:    img,
				mask:     image.Rect(10, 10, tt.w-10, tt.h-10),
			}

			og := &OpenGraph{
				ogi: ogi, title: "Test", c: color.Black, dpi: 72, size: 32,
			}

			rd, err := ogi.Generate(og)
			if err != nil {
				t.Fatalf("Generate() returned error: %v", err)
			}
			defer rd.Close()

			decoded, err := png.Decode(rd)
			if err != nil {
				t.Fatalf("png.Decode() returned error: %v", err)
			}

			bounds := decoded.Bounds()
			if bounds.Dx() != tt.w || bounds.Dy() != tt.h {
				t.Errorf("size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tt.w, tt.h)
			}
		})
	}
}

func TestGenerateDoesNotMutateSource(t *testing.T) {
	fnt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{R: 0x42, G: 0x42, B: 0x42, A: 0xff}), image.Point{}, draw.Src)

	origPix := make([]byte, len(img.Pix))
	copy(origPix, img.Pix)

	ogi := &OpenGraphImageGen{
		template: "test",
		font:     fnt,
		image:    img,
		mask:     image.Rect(5, 5, 95, 95),
	}

	og := &OpenGraph{
		ogi: ogi, title: "Mut", c: color.Black, dpi: 72, size: 20,
	}

	rd, err := ogi.Generate(og)
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}
	rd.Close()

	if !bytes.Equal(img.Pix, origPix) {
		t.Error("Generate() mutated the source image")
	}
}

func TestGenerateWithFaceCloseBehavior(t *testing.T) {
	fnt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 1200, 630))
	ogi := &OpenGraphImageGen{
		template: "test",
		font:     fnt,
		image:    img,
		mask:     image.Rect(50, 20, 910, 510),
	}

	for i := range 3 {
		og := &OpenGraph{
			ogi: ogi, title: "Repeat", c: color.Black, dpi: 72, size: 48,
		}
		rd, err := ogi.Generate(og)
		if err != nil {
			t.Fatalf("Generate() call %d returned error: %v", i, err)
		}
		rd.Close()
	}
}
