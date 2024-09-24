package markdown

import (
	"bytes"
	"os"

	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func getGoldmark(style string) goldmark.Markdown {
	opt := []highlighting.Option{}
	if style != "" {
		opt = append(opt, highlighting.WithStyle(style))
	}

	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			emoji.Emoji,
			meta.Meta,
			highlighting.NewHighlighting(opt...),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
}

func ParseFile(style string, f string) (string, map[string]interface{}, error) {
	src, err := os.ReadFile(f)
	if err != nil {
		return "", nil, err
	}

	buf := &bytes.Buffer{}
	context := parser.NewContext()
	if err := getGoldmark(style).Convert(src, buf, parser.WithContext(context)); err != nil {
		return "", nil, err
	}
	return buf.String(), meta.Get(context), nil
}

func GetMetadataProperty(f string, prop string, dflt interface{}) (interface{}, error) {
	src, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}

	context := parser.NewContext()
	getGoldmark("").Parser().Parse(text.NewReader(src), parser.WithContext(context))

	if m := meta.Get(context); m != nil {
		if v, ok := m[prop]; ok {
			return v, nil
		}
	}
	return dflt, nil
}
