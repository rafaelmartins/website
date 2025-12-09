package content

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"rafaelmartins.com/p/website/internal/frontmatter"
)

type html struct{}

func (*html) IsSupported(f string) bool {
	e := filepath.Ext(f)
	return e == ".htm" || e == ".html"
}

func (*html) Render(f string, style string, baseurl string) (string, *frontmatter.FrontMatter, error) {
	fp, err := os.Open(f)
	if err != nil {
		return "", nil, err
	}
	defer fp.Close()

	meta, src, err := frontmatter.Parse(fp)
	if err != nil {
		return "", nil, err
	}
	return string(src), meta, nil
}

func (*html) GetTimeStamps(f string) ([]time.Time, error) {
	st, err := os.Stat(f)
	if err != nil {
		return nil, err
	}
	return []time.Time{st.ModTime().UTC()}, nil
}

func (*html) ListAssets(f string) ([]string, error) {
	return nil, nil
}

func (*html) OpenAsset(f string, a string) (string, io.ReadCloser, error) {
	return "", nil, errors.New("content: html: assets not supported")
}
