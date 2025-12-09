package content

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"rafaelmartins.com/p/website/internal/frontmatter"
	"rafaelmartins.com/p/website/internal/markdown"
)

type mkd struct{}

func (*mkd) IsSupported(f string) bool {
	e := filepath.Ext(f)
	return e == ".md" || e == ".markdown"
}

func (*mkd) Render(f string, style string, baseurl string) (string, *frontmatter.FrontMatter, error) {
	fp, err := os.Open(f)
	if err != nil {
		return "", nil, err
	}
	defer fp.Close()

	meta, src, err := frontmatter.Parse(fp)
	if err != nil {
		return "", nil, err
	}

	m, err := markdown.Render(src, style, nil)
	if err != nil {
		return "", nil, err
	}
	return m, meta, nil
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
