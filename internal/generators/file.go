package generators

import (
	"io"
	"os"

	"rafaelmartins.com/p/website/internal/runner"
)

type File string

func (File) GetID() string {
	return "FILE"
}

func (s File) GetReader() (io.ReadCloser, error) {
	return os.Open(string(s))
}

func (s File) GetPaths() ([]string, error) {
	return []string{string(s)}, nil
}

func (File) GetImmutable() bool {
	return false
}

func (File) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		close(ch)
	}
}
