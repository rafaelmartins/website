package markdown

import (
	"bytes"
	"errors"
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	gmTitle = goldmark.New(goldmark.WithExtensions(&titleExtension{}))

	pcTitleKey = parser.NewContextKey()
)

type titleExtension struct{}

func (te *titleExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(te, 0)))
}

func (*titleExtension) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	if fc := node.FirstChild(); pc != nil && fc != nil && fc.Kind() == ast.KindHeading {
		pc.Set(pcTitleKey, string(fc.(*ast.Heading).Lines().Value(reader.Source())))
	}
}

func GetTitle(src []byte) (string, error) {
	line := []byte{}
	for l := range bytes.Lines(src) {
		line = l
		break
	}

	pc := parser.NewContext()
	if err := gmTitle.Convert(line, io.Discard, parser.WithContext(pc)); err != nil {
		return "", err
	}

	if title := pc.Get(pcTitleKey); title != nil {
		return title.(string), nil
	}
	return "", errors.New("markdown: title: not found")
}
