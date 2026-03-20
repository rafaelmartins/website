package markdown

import (
	"bytes"

	figure "github.com/mangoumbrella/goldmark-figure"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func New(style string, ext ...goldmark.Extender) goldmark.Markdown {
	opt := []highlighting.Option{}
	if style != "" {
		opt = append(opt, highlighting.WithStyle(style))
	}

	return goldmark.New(
		goldmark.WithExtensions(
			append(
				[]goldmark.Extender{
					&admonitions{},
					&spoilers{},
					&toc{},
					extension.GFM,
					extension.DefinitionList,
					extension.Footnote,
					emoji.Emoji,
					figure.Figure.WithSkipNoCaption(),
					highlighting.NewHighlighting(opt...),
				},
				ext...,
			)...,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
}

func Render(mkd goldmark.Markdown, src []byte, pc parser.Context) (string, string, error) {
	buf := &bytes.Buffer{}
	if err := mkd.Convert(src, buf, parser.WithContext(pc)); err != nil {
		return "", "", err
	}

	t, err := tocRender(pc)
	if err != nil {
		return "", "", err
	}
	return t, buf.String(), nil
}
