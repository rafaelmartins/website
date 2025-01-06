package cdocs

import (
	"bytes"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

func highlight(code string) (string, error) {
	iter, err := chroma.Coalesce(lexers.Get("c")).Tokenise(nil, code)
	if err != nil {
		return "", err
	}

	formatter := html.New(
		html.WithClasses(false),
		html.WithLineNumbers(false),
	)

	// we use github style to match the project pages, that are rendered by github api
	b := bytes.Buffer{}
	if err := formatter.Format(&b, styles.Get("github"), iter); err != nil {
		return "", err
	}

	return b.String(), nil
}
