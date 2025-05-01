package content

import (
	"encoding/json"
	"errors"
	"fmt"
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
	"rafaelmartins.com/p/website/internal/content/frontmatter"
)

var pcTitleKey = parser.NewContextKey()

type tbTransformer struct {
	baseurl string
}

func (tt *tbTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	if fc := node.FirstChild(); pc != nil && fc != nil && fc.Kind() == ast.KindHeading {
		pc.Set(pcTitleKey, string(fc.(*ast.Heading).Lines().Value(reader.Source())))
		node.RemoveChild(node, fc)
	}

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindImage {
			img := n.(*ast.Image)
			if s := string(img.Destination); tt.baseurl != "" && strings.HasPrefix(s, "assets/") {
				img.Destination = []byte(filepath.Join(tt.baseurl, s))
				return ast.WalkStop, nil
			}
		}
		return ast.WalkContinue, nil
	})
}

type tbExtension struct {
	baseurl string
}

func (te *tbExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&tbTransformer{te.baseurl}, 0),
	))
}

func tbRender(src []byte, style string, baseurl string) (string, *frontmatter.FrontMatter, error) {
	pc := parser.NewContext()

	rendered, meta, err := mkdRender(src, style, pc, &tbExtension{baseurl})
	if err != nil {
		return "", nil, err
	}

	if title, ok := pc.Get(pcTitleKey).(string); ok && title != "" {
		meta.Title = title
	}

	return rendered, meta, nil
}

func tbValidate(info []byte) error {
	v := &struct {
		Version int    `json:"version"`
		Type    string `json:"type"`
	}{}

	if err := json.Unmarshal(info, v); err != nil {
		return err
	}

	if v.Version >= 2 {
		if v.Type != "net.daringfireball.markdown" {
			return fmt.Errorf("content: textbundle: invalid type: %s", v.Type)
		}
	}

	return nil
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

func (*textBundle) getSource(f string) (string, error) {
	srcs, err := filepath.Glob(filepath.Join(f, "text.*"))
	if err != nil {
		return "", err
	}

	if l := len(srcs); l == 0 {
		return "", errors.New("content: textbundle: no text file found")
	} else if l > 1 {
		return "", errors.New("content: textbundle: more than one text file found")
	}
	return srcs[0], nil
}

func (tb *textBundle) Render(f string, style string, baseurl string) (string, *frontmatter.FrontMatter, error) {
	info, err := os.ReadFile(filepath.Join(f, "info.json"))
	if err != nil {
		return "", nil, err
	}
	if err := tbValidate(info); err != nil {
		return "", nil, err
	}

	srcf, err := tb.getSource(f)
	if err != nil {
		return "", nil, err
	}

	src, err := os.ReadFile(srcf)
	if err != nil {
		return "", nil, err
	}
	return tbRender(src, style, baseurl)
}

func (tb *textBundle) GetTimeStamps(f string) ([]time.Time, error) {
	src, err := tb.getSource(f)
	if err != nil {
		return nil, err
	}
	st, err := os.Stat(src)
	if err != nil {
		return nil, err
	}
	rv := []time.Time{st.ModTime().UTC()}

	ist, err := os.Stat(filepath.Join(f, "info.json"))
	if err != nil {
		return nil, err
	}
	rv = append(rv, ist.ModTime().UTC())

	assets, err := tb.ListAssets(f)
	if err != nil {
		return nil, err
	}
	for _, asset := range assets {
		st, err := os.Stat(filepath.Join(f, "assets", asset))
		if err != nil {
			return nil, err
		}
		rv = append(rv, st.ModTime().UTC())
	}
	return rv, nil
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

func (*textBundle) OpenAsset(f string, a string) (string, io.ReadCloser, error) {
	fp, err := os.Open(filepath.Join(f, "assets", a))
	if err != nil {
		return "", nil, err
	}
	return filepath.Join("assets", a), fp, nil
}
