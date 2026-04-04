package dfuse

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/ulikunitz/xz"
)

func decompress(f string, r io.ReadCloser) (io.Reader, error) {
	defer r.Close()

	compress := ""
	archive := ""

	switch {
	case strings.HasSuffix(f, ".tar.xz"):
		compress = "xz"
		archive = "tar"

	case strings.HasSuffix(f, ".tar.bz2"):
		compress = "bzip2"
		archive = "tar"

	case strings.HasSuffix(f, ".tar.gz"):
		compress = "gzip"
		archive = "tar"

	case strings.HasSuffix(f, ".zip"):
		compress = ""
		archive = "zip"

	default:
		return nil, fmt.Errorf("dfuse: unknown archive format: %s", f)
	}

	rd := io.Reader(r)

	switch compress {
	case "xz":
		rr, err := xz.NewReader(r)
		if err != nil {
			return nil, err
		}
		rd = rr

	case "bzip2":
		rd = bzip2.NewReader(r)

	case "gzip":
		rr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		rd = rr
	}

	var df io.Reader

	switch archive {
	case "tar":
		t := tar.NewReader(rd)
		for {
			hdr, err := t.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			if strings.HasSuffix(hdr.Name, ".dfu") {
				df = t
				break
			}
		}

	case "zip":
		data, err := io.ReadAll(rd)
		if err != nil {
			return nil, err
		}

		z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, err
		}

		for _, f := range z.File {
			if !strings.HasSuffix(f.Name, ".dfu") {
				continue
			}

			fp, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer fp.Close()
			df = fp
		}
	}

	if df == nil {
		return nil, fmt.Errorf("dfuse: no .dfu found in archive: %s", f)
	}

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, df); err != nil {
		return nil, err
	}
	return buf, nil
}
