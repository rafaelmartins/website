package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func Render(src []byte, style string, pc parser.Context, ext ...goldmark.Extender) (string, error) {
	opt := []highlighting.Option{}
	if style != "" {
		opt = append(opt, highlighting.WithStyle(style))
	}

	mkd := goldmark.New(
		goldmark.WithExtensions(
			append(
				[]goldmark.Extender{
					extension.GFM,
					extension.DefinitionList,
					extension.Footnote,
					emoji.Emoji,
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

	if pc == nil {
		pc = parser.NewContext()
	}

	buf := &bytes.Buffer{}
	if err := mkd.Convert(src, buf, parser.WithContext(pc)); err != nil {
		return "", err
	}
	return buf.String(), nil
}
