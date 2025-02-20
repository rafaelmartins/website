package generators

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/rafaelmartins/website/internal/content"
	"github.com/rafaelmartins/website/internal/ogimage"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
)

type ContentSource struct {
	File string
	URL  string
}

type Content struct {
	Title             string
	Description       string
	URL               string
	Slug              string
	Sources           []*ContentSource
	IsPost            bool
	ExtraDependencies []string
	HighlightStyle    string
	Template          string
	TemplateCtx       map[string]interface{}
	Pagination        *templates.ContentPagination
	LayoutCtx         *templates.LayoutContext

	OpenGraphTitle         string
	OpenGraphDescription   string
	OpenGraphImage         string
	OpenGraphImageURL      string
	OpenGraphImageGenerate bool
	OpenGraphImageGenColor *uint32
	OpenGraphImageGenDPI   *float64
	OpenGraphImageGenSize  *float64

	ctx      *templates.ContentContext
	metadata *content.Metadata
}

func (*Content) GetID() string {
	return "CONTENT"
}

func (h *Content) GetReader() (io.ReadCloser, error) {
	if h.URL == "" {
		return nil, errors.New("markdown: missing url")
	}

	ctx := &templates.ContentContext{
		Title:       h.Title,
		Description: h.Description,
		URL:         h.URL,
		Slug:        h.Slug,
		OpenGraph: templates.OpenGraphEntry{
			Title:       h.OpenGraphTitle,
			Description: h.OpenGraphDescription,
			Image:       h.OpenGraphImageURL,
		},
		Atom:       &templates.AtomContentEntry{},
		Pagination: h.Pagination,
		Extra:      h.TemplateCtx,
	}

	if ctx.OpenGraph.Title == "" {
		ctx.OpenGraph.Title = h.Title
	}
	if ctx.OpenGraph.Description == "" {
		ctx.OpenGraph.Description = h.Description
	}
	if h.OpenGraphImageGenerate && ctx.OpenGraph.Image == "" {
		ctx.OpenGraph.Image = ogimage.URL(h.URL)
	}
	h.ctx = ctx

	atomUpdated := time.Time{}
	entries := []*templates.ContentEntry{}

	for _, src := range h.Sources {
		if src.File == "" {
			continue
		}

		body, metadata, err := content.Render(src.File, h.HighlightStyle, h.URL)
		if err != nil {
			return nil, err
		}

		entry := &templates.ContentEntry{
			File:  src.File,
			URL:   src.URL,
			Title: metadata.Title,
			Body:  body,
		}

		if ctx.OpenGraph.Title == "" {
			ctx.OpenGraph.Title = metadata.Title
		}
		if ctx.OpenGraph.Description == "" {
			ctx.OpenGraph.Description = metadata.Description
		}

		if h.IsPost {
			entry.Post = &templates.PostContentEntry{
				Date: metadata.Date.Time,
			}
			entry.Post.Author.Name = metadata.Author.Name
			entry.Post.Author.Email = metadata.Author.Email
			if atomUpdated.IsZero() {
				atomUpdated = entry.Post.Date
			}
		}

		entry.Extra = metadata.Extra

		if h.Pagination == nil {
			if h.Title == "" {
				ctx.Title = entry.Title
			}
			ctx.Entry = entry
			h.metadata = metadata
			break
		}

		entries = append(entries, entry)
	}

	if h.Pagination != nil {
		ctx.Entries = entries
		ctx.Atom.Updated = atomUpdated
		if ctx.Atom.Updated.IsZero() {
			ctx.Atom.Updated = time.Unix(0, 0)
		}
	}

	funcMap := template.FuncMap{
		"contentGetMetadata": content.GetMetadata,
	}

	buf := &bytes.Buffer{}
	if err := templates.Execute(buf, h.Template, funcMap, h.LayoutCtx, ctx); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (h *Content) GetTimeStamps() ([]time.Time, error) {
	rv, err := templates.GetTimestamps(h.Template, true)
	if err != nil {
		return nil, err
	}

	og, err := ogimage.GetTimeStamps()
	if err != nil {
		return nil, err
	}
	rv = append(rv, og...)

	for _, src := range h.Sources {
		if src.File == "" {
			continue
		}

		st, err := os.Stat(src.File)
		if err != nil {
			return nil, err
		}
		rv = append(rv, st.ModTime().UTC())

		assets, err := content.ListAssetTimeStamps(src.File)
		if err != nil {
			return nil, err
		}
		rv = append(rv, assets...)
	}

	for _, dep := range h.ExtraDependencies {
		st, err := os.Stat(dep)
		if err != nil {
			return nil, err
		}
		rv = append(rv, st.ModTime().UTC())
	}

	return rv, nil
}

func (*Content) GetImmutable() bool {
	return false
}

func (h *Content) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch == nil {
		return
	}

	if h.ctx.Entry != nil {
		assets, err := content.ListAssets(h.ctx.Entry.File)
		if err != nil {
			ch <- &runner.GeneratorByProduct{Err: err}
			return
		}
		for _, asset := range assets {
			fn, fp, err := content.OpenAsset(h.ctx.Entry.File, asset)
			if err != nil {
				ch <- &runner.GeneratorByProduct{Err: err}
				return
			}
			ch <- &runner.GeneratorByProduct{
				Filename: fn,
				Reader:   fp,
			}
		}
	}

	image := h.OpenGraphImage
	ccolor := h.OpenGraphImageGenColor
	dpi := h.OpenGraphImageGenDPI
	size := h.OpenGraphImageGenSize

	if h.metadata != nil {
		if h.metadata.OpenGraph.Image != "" {
			image = filepath.Join(filepath.Dir(h.Sources[0].File), h.metadata.OpenGraph.Image)
		}
		if h.metadata.OpenGraph.ImageGen.Color != nil {
			ccolor = h.metadata.OpenGraph.ImageGen.Color
		}
		if h.metadata.OpenGraph.ImageGen.DPI != nil {
			dpi = h.metadata.OpenGraph.ImageGen.DPI
		}
		if h.metadata.OpenGraph.ImageGen.Size != nil {
			size = h.metadata.OpenGraph.ImageGen.Size
		}
	}

	ogimage.GenerateByProduct(ch, h.ctx.OpenGraph.Title, h.OpenGraphImageGenerate, image, ccolor, dpi, size)
	close(ch)
}
