package content

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/yuin/goldmark/parser"
	"rafaelmartins.com/p/website/internal/frontmatter"
	"rafaelmartins.com/p/website/internal/markdown"
)

var gmMarkdown = markdown.New("github")

type mkd struct{}

func (*mkd) IsSupported(f string) bool {
	e := filepath.Ext(f)
	return e == ".md" || e == ".markdown"
}

func (*mkd) Render(f string, baseurl string, withToc *bool) (*frontmatter.FrontMatter, string, string, error) {
	fp, err := os.Open(f)
	if err != nil {
		return nil, "", "", err
	}
	defer fp.Close()

	src, err := io.ReadAll(fp)
	if err != nil {
		return nil, "", "", err
	}

	meta, src, err := frontmatter.Parse(src)
	if err != nil {
		return nil, "", "", err
	}
	if withToc != nil && meta.Toc != nil {
		withToc = meta.Toc
	}

	pc := parser.NewContext()
	pc.Set(markdown.PcTocEnable, withToc)
	t, m, err := markdown.Render(gmMarkdown, src, pc)
	if err != nil {
		return nil, "", "", err
	}
	return meta, t, m, nil
}

func (*mkd) GetTimeStamps(f string) ([]time.Time, error) {
	st, err := os.Stat(f)
	if err != nil {
		return nil, err
	}
	return []time.Time{st.ModTime().UTC()}, nil
}

func (*mkd) ListAssets(f string) ([]string, error) {
	return nil, nil
}

func (*mkd) OpenAsset(f string, a string) (string, io.ReadCloser, error) {
	return "", nil, errors.New("content: markdown: assets not supported")
}
