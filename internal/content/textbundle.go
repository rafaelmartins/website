package content

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type imageTransformer struct {
	baseurl string
}

func (it *imageTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindImage {
			img := n.(*ast.Image)
			if s := string(img.Destination); it.baseurl != "" && strings.HasPrefix(s, "assets/") {
				img.Destination = []byte(filepath.Join(it.baseurl, s))
			}
		}
		return ast.WalkContinue, nil
	})
}

type imageExtension struct {
	baseurl string
}

func (ie *imageExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&imageTransformer{ie.baseurl}, 0),
	))
}

type textBundle struct{}

func (*textBundle) IsSupported(f string) bool {
	if filepath.Ext(f) != ".textbundle" {
		return false
	}

	st, err := os.Stat(f)
	if err != nil {
		return false
	}
	return st.IsDir()
}

func (*textBundle) Render(f string, style string, baseurl string) (string, *Metadata, error) {
	srcs, err := filepath.Glob(filepath.Join(f, "text.*"))
	if err != nil {
		return "", nil, err
	}

	if l := len(srcs); l == 0 {
		return "", nil, errors.New("content: textbundle: no text file found")
	} else if l > 1 {
		return "", nil, errors.New("content: textbundle: more than one text file found")
	}

	src, err := os.ReadFile(srcs[0])
	if err != nil {
		return "", nil, err
	}
	return mkdRender(src, style, &imageExtension{baseurl})
}

func (*textBundle) ListAssets(f string) ([]string, error) {
	dir := filepath.Join(f, "assets")

	rv := []string{}
	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		r, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rv = append(rv, r)
		return nil
	}); err != nil {
		return nil, err
	}
	return rv, nil
}

func (tb *textBundle) ListAssetTimeStamps(f string) ([]time.Time, error) {
	assets, err := tb.ListAssets(f)
	if err != nil {
		return nil, err
	}

	rv := []time.Time{}
	for _, asset := range assets {
		st, err := os.Stat(filepath.Join(f, "assets", asset))
		if err != nil {
			return nil, err
		}
		rv = append(rv, st.ModTime().UTC())
	}
	return rv, nil
}

func (*textBundle) OpenAsset(f string, a string) (string, io.ReadCloser, error) {
	fp, err := os.Open(filepath.Join(f, "assets", a))
	if err != nil {
		return "", nil, err
	}
	return filepath.Join("assets", a), fp, nil
}
