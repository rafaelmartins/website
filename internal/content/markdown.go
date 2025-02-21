package content

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

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

func mkdRender(src []byte, style string, pc parser.Context, ext ...goldmark.Extender) (string, *Metadata, error) {
	opt := []highlighting.Option{}
	if style != "" {
		opt = append(opt, highlighting.WithStyle(style))
	}

	mkd := goldmark.New(
		goldmark.WithExtensions(
			append(
				[]goldmark.Extender{
					extension.GFM,
					emoji.Emoji,
					&frontmatter.Extender{},
					highlighting.NewHighlighting(opt...),
				},
				ext...,
			)...,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	if pc == nil {
		pc = parser.NewContext()
	}

	buf := &bytes.Buffer{}
	if err := mkd.Convert(src, buf, parser.WithContext(pc)); err != nil {
		return "", nil, err
	}

	md := &Metadata{}
	if m := frontmatter.Get(pc); m != nil {
		if err := m.Decode(md); err != nil {
			return "", nil, err
		}
	}
	return buf.String(), md, nil
}

type markdown struct{}

func (*markdown) IsSupported(f string) bool {
	e := filepath.Ext(f)
	return e == ".md" || e == ".markdown"
}

func (*markdown) Render(f string, style string, baseurl string) (string, *Metadata, error) {
	src, err := os.ReadFile(f)
	if err != nil {
		return "", nil, err
	}

	return mkdRender(src, style, nil)
}

func (*markdown) ListAssets(f string) ([]string, error) {
	return nil, nil
}

func (*markdown) ListAssetTimeStamps(f string) ([]time.Time, error) {
	return nil, nil
}

func (*markdown) OpenAsset(f string, a string) (string, io.ReadCloser, error) {
	return "", nil, errors.New("content: markdown: assets not supported")
}
