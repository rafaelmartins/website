package cdocs

import (
	"bytes"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

var (
	chFormatter = html.New(
		html.WithClasses(false),
		html.WithLineNumbers(false),
	)
	chLexer = chroma.Coalesce(lexers.Get("c"))
	chStyle = styles.Get("github")
)

func highlight(code string) (string, error) {
	iter, err := chLexer.Tokenise(nil, code)
	if err != nil {
		return "", err
	}

	// we use github style to match the project pages, that are rendered by github api
	b := bytes.Buffer{}
	if err := chFormatter.Format(&b, chStyle, iter); err != nil {
		return "", err
	}
	return b.String(), nil
}
