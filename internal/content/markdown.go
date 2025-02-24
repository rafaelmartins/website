package content

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rafaelmartins/website/internal/content/frontmatter"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

func mkdRender(src []byte, style string, pc parser.Context, ext ...goldmark.Extender) (string, *frontmatter.FrontMatter, error) {
	meta, src, err := frontmatter.Parse(src)
	if err != nil {
		return "", nil, err
	}

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

	return buf.String(), meta, nil
}

type markdown struct{}

func (*markdown) IsSupported(f string) bool {
	e := filepath.Ext(f)
	return e == ".md" || e == ".markdown"
}

func (*markdown) Render(f string, style string, baseurl string) (string, *frontmatter.FrontMatter, error) {
	src, err := os.ReadFile(f)
	if err != nil {
		return "", nil, err
	}

	return mkdRender(src, style, nil)
}

func (*markdown) GetTimeStamps(f string) ([]time.Time, error) {
	st, err := os.Stat(f)
	if err != nil {
		return nil, err
	}
	return []time.Time{st.ModTime().UTC()}, nil
}

func (*markdown) ListAssets(f string) ([]string, error) {
	return nil, nil
}

func (*markdown) OpenAsset(f string, a string) (string, io.ReadCloser, error) {
	return "", nil, errors.New("content: markdown: assets not supported")
}
