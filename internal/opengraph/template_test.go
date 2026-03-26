package opengraph

import (
	"image/color"
	"testing"
)

func TestGetTemplateContext(t *testing.T) {
	ogi := &OpenGraphImageGen{template: "test"}

	args := []struct {
		name      string
		og        *OpenGraph
		wantTitle string
		wantDesc  string
		wantImage string
	}{
		{
			"generate with blog path",
			&OpenGraph{
				ogi:         ogi,
				baseurl:     "/blog/hello-world/",
				title:       "Hello, World!",
				description: "Yep, I'm starting a blog again!",
				generate:    true,
				c:           color.Black,
			},
			"Hello, World!",
			"Yep, I'm starting a blog again!",
			"/blog/hello-world/opengraph.png",
		},
		{
			"generate false no image",
			&OpenGraph{
				baseurl:     "/",
				title:       "About me",
				description: "Rafael Martins' Website",
				generate:    false,
			},
			"About me",
			"Rafael Martins' Website",
			"",
		},
		{
			"generate true no ogi no image",
			&OpenGraph{
				baseurl:     "/blog/test/",
				title:       "Test",
				description: "Desc",
				generate:    true,
			},
			"Test",
			"Desc",
			"",
		},
		{
			"generate true ogi empty template",
			&OpenGraph{
				ogi:         &OpenGraphImageGen{template: ""},
				baseurl:     "/blog/test/",
				title:       "Test",
				description: "Desc",
				generate:    true,
			},
			"Test",
			"Desc",
			"",
		},
		{
			"generate false with explicit image",
			&OpenGraph{
				baseurl:     "/p/website/",
				title:       "website",
				description: "Yet another NIH static website framework.",
				generate:    false,
				image:       "/some/image.png",
			},
			"website",
			"Yet another NIH static website framework.",
			"/p/website/opengraph.png",
		},
		{
			"empty baseurl",
			&OpenGraph{
				ogi:      ogi,
				baseurl:  "",
				title:    "Test",
				generate: true,
			},
			"Test",
			"",
			"",
		},
		{
			"root path",
			&OpenGraph{
				ogi:         ogi,
				baseurl:     "/",
				title:       "About me",
				description: "Rafael Martins' Website",
				generate:    true,
			},
			"About me",
			"Rafael Martins' Website",
			"/opengraph.png",
		},
		{
			"series path",
			&OpenGraph{
				ogi:         ogi,
				baseurl:     "/series/rcsid/",
				title:       "RCSID",
				description: "Series about RCSID",
				generate:    true,
			},
			"RCSID",
			"Series about RCSID",
			"/series/rcsid/opengraph.png",
		},
		{
			"project path",
			&OpenGraph{
				ogi:         ogi,
				baseurl:     "/p/blogc/",
				title:       "blogc",
				description: "A blog compiler",
				generate:    true,
			},
			"blogc",
			"A blog compiler",
			"/p/blogc/opengraph.png",
		},
		{
			"empty description",
			&OpenGraph{
				ogi:      ogi,
				baseurl:  "/blog/test/",
				title:    "Test Post",
				generate: true,
			},
			"Test Post",
			"",
			"/blog/test/opengraph.png",
		},
		{
			"pregenerated",
			&OpenGraph{
				pregenerated: true,
				baseurl:      "/blog/",
				title:        "Blog",
				description:  "My Blog",
				generate:     false,
			},
			"Blog",
			"My Blog",
			"/blog/opengraph.png",
		},
		{
			"pregenerated empty baseurl",
			&OpenGraph{
				pregenerated: true,
				baseurl:      "",
				title:        "Test",
				generate:     false,
			},
			"Test",
			"",
			"",
		},
	}

	for _, tt := range args {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.og.GetTemplateContext()
			if ctx.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", ctx.Title, tt.wantTitle)
			}
			if ctx.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", ctx.Description, tt.wantDesc)
			}
			if ctx.Image != tt.wantImage {
				t.Errorf("Image = %q, want %q", ctx.Image, tt.wantImage)
			}
		})
	}
}
