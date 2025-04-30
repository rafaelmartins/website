package postproc

import (
	"io"
	"os"
	"path/filepath"
)

type Processor interface {
	Supported(ext string) bool
	Run(dstFn string, dst io.Writer, src io.Reader) error
}

var registry = []Processor{
	&Minify{},
	&JPEG{},
	&PNG{},
}

func PostProc(dst string, src io.ReadCloser) error {
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return err
	}

	fp, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fp.Close()

	pproc := (Processor)(nil)
	if ext := filepath.Ext(dst); ext != "" {
		for _, proc := range registry {
			if proc.Supported(ext) {
				pproc = proc
				break
			}
		}
	}

	if pproc != nil {
		return pproc.Run(dst, fp, src)
	}
	_, err = io.Copy(fp, src)
	return err
}
