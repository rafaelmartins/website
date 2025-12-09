package content

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"rafaelmartins.com/p/website/internal/frontmatter"
)

type textPack struct{}

func (*textPack) IsSupported(f string) bool {
	return filepath.Ext(f) == ".textpack"
}

func (*textPack) Render(f string, style string, baseurl string) (string, *frontmatter.FrontMatter, error) {
	r, err := zip.OpenReader(f)
	if err != nil {
		return "", nil, err
	}
	defer r.Close()

	info := (*zip.File)(nil)
	src := (*zip.File)(nil)
	for _, f := range r.File {
		if m, err := filepath.Match(path.Join("*.textbundle", "info.json"), f.Name); err != nil {
			return "", nil, err
		} else if m {
			info = f
		}

		if m, err := filepath.Match(path.Join("*.textbundle", "text.*"), f.Name); err != nil {
			return "", nil, err
		} else if m {
			if src != nil {
				return "", nil, errors.New("content: textpack: more than one text file found")
			}
			src = f
		}
	}

	if info == nil {
		return "", nil, errors.New("content: textpack: no info.json file found")
	}
	ifp, err := info.Open()
	if err != nil {
		return "", nil, err
	}
	defer ifp.Close()

	idata, err := io.ReadAll(ifp)
	if err != nil {
		return "", nil, err
	}
	if err := tbValidate(idata); err != nil {
		return "", nil, err
	}

	if src == nil {
		return "", nil, errors.New("content: textpack: no text file found")
	}
	fp, err := src.Open()
	if err != nil {
		return "", nil, err
	}
	defer fp.Close()

	return tbRender(fp, style, baseurl)
}

func (*textPack) GetTimeStamps(f string) ([]time.Time, error) {
	st, err := os.Stat(f)
	if err != nil {
		return nil, err
	}
	return []time.Time{st.ModTime().UTC()}, nil
}

func (*textPack) ListAssets(f string) ([]string, error) {
	r, err := zip.OpenReader(f)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	rv := []string{}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		p := strings.Split(f.Name, "/")
		if len(p) > 2 && strings.HasSuffix(p[0], ".textbundle") && p[1] == "assets" && p[2] != "" {
			rv = append(rv, path.Join(p[2:]...))
		}
	}
	return rv, nil
}

func (*textPack) OpenAsset(f string, a string) (string, io.ReadCloser, error) {
	r, err := zip.OpenReader(f)
	if err != nil {
		return "", nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		if m, err := filepath.Match(path.Join("*.textbundle", "assets", a), f.Name); err != nil {
			return "", nil, err
		} else if m {
			fp, err := f.Open()
			if err != nil {
				return "", nil, err
			}
			defer fp.Close()

			buf := &bytes.Buffer{}
			if _, err := io.Copy(buf, fp); err != nil {
				return "", nil, err
			}
			return filepath.Join("assets", a), io.NopCloser(buf), nil
		}
	}
	return "", nil, fmt.Errorf("content: textpack: not found: %s", a)
}
