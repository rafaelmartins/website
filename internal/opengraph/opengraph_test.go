package opengraph

import (
	"image/color"
	"testing"

	"rafaelmartins.com/p/website/internal/hexcolor"
)

func TestNewNilOpenGraphImageGen(t *testing.T) {
	og, err := New(nil, false, "/", "About me", "Rafael Martins' Website", nil, "", "", nil)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if og.ogi != nil {
		t.Error("ogi should be nil")
	}
	if og.title != "About me" {
		t.Errorf("title = %q, want %q", og.title, "About me")
	}
	if !og.generate {
		t.Error("generate should be true")
	}
}

func TestNewDefaults(t *testing.T) {
	ogi := &OpenGraphImageGen{
		dcolor: color.Black,
		ddpi:   72,
		dsize:  96,
	}

	og, err := New(ogi, false, "/", "About me", "Rafael Martins' Website", nil, "", "", nil)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if og.title != "About me" {
		t.Errorf("title = %q, want %q", og.title, "About me")
	}
	if og.description != "Rafael Martins' Website" {
		t.Errorf("description = %q, want %q", og.description, "Rafael Martins' Website")
	}
	if og.baseurl != "/" {
		t.Errorf("baseurl = %q, want %q", og.baseurl, "/")
	}
	if og.ogi != ogi {
		t.Error("ogi should be set")
	}
	if !og.generate {
		t.Error("generate should be true by default")
	}
}

func TestNewWithOpenGraphImageGen(t *testing.T) {
	ogi := &OpenGraphImageGen{
		dcolor: color.RGBA{R: 0x33, G: 0x66, B: 0x99, A: 0xff},
		ddpi:   144,
		dsize:  48,
	}

	og, err := New(ogi, false, "/blog/hello-world/", "Hello, World!", "Yep, I'm starting a blog again!", nil, "", "", nil)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if og.c != ogi.dcolor {
		t.Errorf("color = %v, want %v", og.c, ogi.dcolor)
	}
	if og.dpi != 144 {
		t.Errorf("dpi = %f, want 144", og.dpi)
	}
	if og.size != 48 {
		t.Errorf("size = %f, want 48", og.size)
	}
	if og.title != "Hello, World!" {
		t.Errorf("title = %q, want %q", og.title, "Hello, World!")
	}
}

func TestNewWithConfig(t *testing.T) {
	cfg := &Config{
		Title:       "Custom Title",
		Description: "Custom Description",
		Image:       "/custom/image.png",
	}
	cfg.ImageGen.Generate = new(false)
	cfg.ImageGen.Color = new("#ff0000")
	cfg.ImageGen.DPI = new(144.0)
	cfg.ImageGen.Size = new(64.0)

	ogi := &OpenGraphImageGen{dcolor: color.Black, ddpi: 72, dsize: 96}
	og, err := New(ogi, false, "/", "Default Title", "Default Desc", cfg, "", "", nil)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if og.title != "Custom Title" {
		t.Errorf("title = %q, want %q", og.title, "Custom Title")
	}
	if og.description != "Custom Description" {
		t.Errorf("description = %q, want %q", og.description, "Custom Description")
	}
	if og.image != "/custom/image.png" {
		t.Errorf("image = %q, want %q", og.image, "/custom/image.png")
	}
	if og.generate {
		t.Error("generate should be false")
	}
	wantColor := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}
	if og.c != wantColor {
		t.Errorf("color = %v, want %v", og.c, wantColor)
	}
	if og.dpi != 144 {
		t.Errorf("dpi = %f, want 144", og.dpi)
	}
	if og.size != 64 {
		t.Errorf("size = %f, want 64", og.size)
	}
}

func TestNewFrontmatterOverridesConfig(t *testing.T) {
	cfg := &Config{
		Title:       "Config Title",
		Description: "Config Description",
	}

	ogi := &OpenGraphImageGen{dcolor: color.Black, ddpi: 72, dsize: 96}
	og, err := New(ogi, false, "/", "Default", "Default", cfg, "Frontmatter Title", "Frontmatter Description", nil)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if og.title != "Frontmatter Title" {
		t.Errorf("title = %q, want %q", og.title, "Frontmatter Title")
	}
	if og.description != "Frontmatter Description" {
		t.Errorf("description = %q, want %q", og.description, "Frontmatter Description")
	}
}

func TestNewMetadataOverridesAll(t *testing.T) {
	cfg := &Config{
		Title: "Config Title",
	}
	metadata := &Config{
		Title:       "Metadata Title",
		Description: "Metadata Description",
	}

	ogi := &OpenGraphImageGen{dcolor: color.Black, ddpi: 72, dsize: 96}
	og, err := New(ogi, false, "/", "Default", "Default", cfg, "Frontmatter Title", "", metadata)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if og.title != "Metadata Title" {
		t.Errorf("title = %q, want %q", og.title, "Metadata Title")
	}
	if og.description != "Metadata Description" {
		t.Errorf("description = %q, want %q", og.description, "Metadata Description")
	}
}

func TestNewInvalidColor(t *testing.T) {
	ogi := &OpenGraphImageGen{dcolor: color.Black, ddpi: 72, dsize: 96}
	cfg := &Config{}
	cfg.ImageGen.Color = new("invalid")

	if _, err := New(ogi, false, "/", "Title", "Desc", cfg, "", "", nil); err == nil {
		t.Error("expected error for invalid color, got nil")
	}
}

func TestNewInvalidMetadataColor(t *testing.T) {
	ogi := &OpenGraphImageGen{dcolor: color.Black, ddpi: 72, dsize: 96}
	metadata := &Config{}
	metadata.ImageGen.Color = new("invalid")

	if _, err := New(ogi, false, "/", "Title", "Desc", nil, "", "", metadata); err == nil {
		t.Error("expected error for invalid metadata color, got nil")
	}
}

func TestApplyNilConfig(t *testing.T) {
	og := &OpenGraph{
		title:       "Original",
		description: "Original Desc",
		generate:    true,
	}

	err := og.apply(nil)
	if err != nil {
		t.Fatalf("apply(nil) returned error: %v", err)
	}

	if og.title != "Original" {
		t.Errorf("title changed: got %q, want %q", og.title, "Original")
	}
	if og.description != "Original Desc" {
		t.Errorf("description changed: got %q, want %q", og.description, "Original Desc")
	}
}

func TestApplyPartialConfig(t *testing.T) {
	og := &OpenGraph{
		title:       "Original Title",
		description: "Original Description",
		generate:    true,
		dpi:         72,
		size:        96,
	}

	cfg := &Config{
		Title: "New Title",
	}
	cfg.ImageGen.DPI = new(144.0)

	err := og.apply(cfg)
	if err != nil {
		t.Fatalf("apply() returned error: %v", err)
	}

	if og.title != "New Title" {
		t.Errorf("title = %q, want %q", og.title, "New Title")
	}
	if og.description != "Original Description" {
		t.Errorf("description changed: got %q, want %q", og.description, "Original Description")
	}
	if og.dpi != 144 {
		t.Errorf("dpi = %f, want 144", og.dpi)
	}
	if og.size != 96 {
		t.Errorf("size changed: got %f, want 96", og.size)
	}
}

func TestApplyColorVariants(t *testing.T) {
	args := []struct {
		name  string
		color string
		want  color.RGBA
	}{
		{"hex 6 digit", "#336699", color.RGBA{R: 0x33, G: 0x66, B: 0x99, A: 0xff}},
		{"hex 3 digit", "#f00", color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}},
		{"hex 8 digit", "#336699ff", color.RGBA{R: 0x33, G: 0x66, B: 0x99, A: 0xff}},
		{"white", "#ffffff", color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}},
	}

	for _, tt := range args {
		t.Run(tt.name, func(t *testing.T) {
			og := &OpenGraph{}
			cfg := &Config{}
			cfg.ImageGen.Color = new(tt.color)

			err := og.apply(cfg)
			if err != nil {
				t.Fatalf("apply() returned error: %v", err)
			}

			got, ok := og.c.(color.RGBA)
			if !ok {
				t.Fatalf("color type = %T, want color.RGBA", og.c)
			}
			if got != tt.want {
				t.Errorf("color = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateByProductNilChannel(t *testing.T) {
	og := &OpenGraph{
		ogi:      &OpenGraphImageGen{template: "test"},
		generate: true,
	}
	og.GenerateByProduct(nil, "/tmp")
}

func TestGenerateByProductNilOpenGraphImageGen(t *testing.T) {
	og := &OpenGraph{
		generate: true,
	}
	og.GenerateByProduct(nil, "/tmp")
}

func TestGenerateByProductNotGenerate(t *testing.T) {
	og := &OpenGraph{
		ogi:      &OpenGraphImageGen{template: "test"},
		generate: false,
	}
	og.GenerateByProduct(nil, "/tmp")
}

func TestGenerateByProductEmptyTemplate(t *testing.T) {
	og := &OpenGraph{
		ogi:      &OpenGraphImageGen{template: ""},
		generate: true,
	}
	og.GenerateByProduct(nil, "/tmp")
}

func TestNewPregenerated(t *testing.T) {
	ogi := &OpenGraphImageGen{
		template: "test",
		dcolor:   color.Black,
		ddpi:     72,
		dsize:    96,
	}

	og, err := New(ogi, true, "/blog/", "Blog", "My Blog", nil, "", "", nil)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if og.generate {
		t.Error("generate should be false when pregenerated is true")
	}
	if !og.pregenerated {
		t.Error("pregenerated should be true")
	}

	ctx := og.GetTemplateContext()
	if ctx.Image != "/blog/opengraph.png" {
		t.Errorf("Image = %q, want %q", ctx.Image, "/blog/opengraph.png")
	}
}

func TestNewPregeneratedOverridesConfig(t *testing.T) {
	ogi := &OpenGraphImageGen{
		template: "test",
		dcolor:   color.Black,
		ddpi:     72,
		dsize:    96,
	}

	cfg := &Config{}
	cfg.ImageGen.Generate = new(true)

	og, err := New(ogi, true, "/blog/", "Blog", "My Blog", cfg, "", "", nil)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if og.generate {
		t.Error("generate should be false: pregenerated must override config")
	}
}

func TestGenerateByProductPregenerated(t *testing.T) {
	og := &OpenGraph{
		ogi:          &OpenGraphImageGen{template: "test"},
		generate:     true,
		pregenerated: true,
	}
	og.GenerateByProduct(nil, "/tmp")
}

func TestNewPriorityChain(t *testing.T) {
	ogi := &OpenGraphImageGen{
		dcolor: color.Black,
		ddpi:   72,
		dsize:  96,
	}

	cfg := &Config{
		Description: "Config-level description",
	}
	cfg.ImageGen.Color = new("#336699")

	metadata := &Config{
		Description: "Metadata description wins",
	}
	metadata.ImageGen.Color = new("#ff0000")

	og, err := New(ogi, false, "/blog/hello-world/", "Default Title", "Default Description", cfg, "Frontmatter Title", "Frontmatter Description", metadata)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if og.title != "Frontmatter Title" {
		t.Errorf("title = %q, want %q", og.title, "Frontmatter Title")
	}
	if og.description != "Metadata description wins" {
		t.Errorf("description = %q, want %q", og.description, "Metadata description wins")
	}

	wantColor, _ := hexcolor.ToRGBA("#ff0000")
	got, ok := og.c.(color.RGBA)
	if !ok {
		t.Fatalf("color type = %T, want color.RGBA", og.c)
	}
	if got != wantColor {
		t.Errorf("color = %v, want %v", got, wantColor)
	}
}
