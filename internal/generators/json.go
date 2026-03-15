package generators

import (
	"bytes"
	"encoding/json"
	"io"

	"rafaelmartins.com/p/website/internal/runner"
)

type Json struct {
	Data any
}

func (Json) GetID() string {
	return "JSON"
}

func (j *Json) GetReader() (io.ReadCloser, error) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(j.Data); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (Json) GetPaths() ([]string, error) {
	return nil, nil
}

func (Json) GetImmutable() bool {
	return false
}

func (Json) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		close(ch)
	}
}
