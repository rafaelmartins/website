package generators

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/rafaelmartins/website/internal/ogimage"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

func mdGetGoldmark(style string) goldmark.Markdown {
	opt := []highlighting.Option{}
	if style != "" {
		opt = append(opt, highlighting.WithStyle(style))
	}

	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			emoji.Emoji,
			&frontmatter.Extender{},
			highlighting.NewHighlighting(opt...),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
}

type MetadataDate struct {
	time.Time
}

func (d *MetadataDate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	s := ""
	if err := unmarshal(&s); err != nil {
		return err
	}

	dt, err1 := time.Parse(time.DateTime, s)
	if err1 == nil {
		d.Time = dt
		return nil
	}

	dt, err := time.Parse(time.DateOnly, s)
	if err == nil {
		d.Time = dt
		return nil
	}
	return err1
}

type Metadata struct {
	Title       string       `yaml:"title"`
	Description string       `yaml:"description"`
	Date        MetadataDate `yaml:"date"`
	Author      struct {
		Name  string `yaml:"name"`
		Email string `yaml:"email"`
	} `yaml:"author"`
	OpenGraph struct {
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
		Image       string `yaml:"image"`
		ImageGen    struct {
			Color *uint32  `yaml:"color"`
			DPI   *float64 `yaml:"dpi"`
			Size  *float64 `yaml:"size"`
		} `yaml:"image-gen"`
	} `yaml:"opengraph"`
	Extra map[string]any `yaml:"extra"`
}

func mdGetMetadataFromContext(ctx parser.Context) (*Metadata, error) {
	rv := &Metadata{}
	m := frontmatter.Get(ctx)
	if m == nil {
		return nil, errors.New("markdown: missing frontmatter")
	}
	if err := m.Decode(rv); err != nil {
		return nil, err
	}
	return rv, nil
}

func MarkdownGetMetadata(f string) (*Metadata, error) {
	src, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}

	ctx := parser.NewContext()
	mdGetGoldmark("").Parser().Parse(text.NewReader(src), parser.WithContext(ctx))
	return mdGetMetadataFromContext(ctx)
}

type MarkdownSource struct {
	File string
	URL  string
}

type Markdown struct {
	Title             string
	Description       string
	URL               string
	Sources           []*MarkdownSource
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
	metadata *Metadata
}

func (*Markdown) GetID() string {
	return "MARKDOWN"
}

func (h *Markdown) GetReader() (io.ReadCloser, error) {
	if h.URL == "" {
		return nil, errors.New("markdown: missing url")
	}

	ctx := &templates.ContentContext{
		Title:       h.Title,
		Description: h.Description,
		URL:         h.URL,
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

		fc, err := os.ReadFile(src.File)
		if err != nil {
			return nil, err
		}

		buf := &bytes.Buffer{}
		context := parser.NewContext()
		if err := mdGetGoldmark(h.HighlightStyle).Convert(fc, buf, parser.WithContext(context)); err != nil {
			return nil, err
		}
		body := buf.String()
		metadata, err := mdGetMetadataFromContext(context)
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
		"markdownGetMetadata": func(f string) interface{} {
			rv, err := MarkdownGetMetadata(f)
			if err != nil {
				panic(err)
			}
			return rv
		},
	}

	if h.OpenGraphImageGenerate {
		if err := ctx.OpenGraph.Validate(); err != nil {
			return nil, err
		}
	}

	buf := &bytes.Buffer{}
	if err := templates.Execute(buf, h.Template, funcMap, h.LayoutCtx, ctx); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (h *Markdown) GetTimeStamps() ([]time.Time, error) {
	rv, err := templates.GetTimestamps(h.Template, true)
	if err != nil {
		return nil, err
	}

	for _, src := range h.Sources {
		if src.File == "" {
			continue
		}

		st, err := os.Stat(src.File)
		if err != nil {
			return nil, err
		}
		rv = append(rv, st.ModTime().UTC())
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

func (*Markdown) GetImmutable() bool {
	return false
}

func (h *Markdown) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch == nil {
		return
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
