package generators

import (
	"io"
	"os"
	"time"

	"rafaelmartins.com/p/website/internal/runner"
)

type File string

func (s File) GetID() string {
	return "FILE"
}

func (s File) GetReader() (io.ReadCloser, error) {
	return os.Open(string(s))
}

func (s File) GetTimeStamps() ([]time.Time, error) {
	st, err := os.Stat(string(s))
	if err != nil {
		return nil, err
	}
	return []time.Time{st.ModTime().UTC()}, nil
}

func (File) GetImmutable() bool {
	return false
}

func (File) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		close(ch)
	}
}
