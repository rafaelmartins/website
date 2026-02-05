package generators

import (
	"embed"
	"io"

	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/utils"
)

type Embed struct {
	FS   embed.FS
	Name string
}

func (*Embed) GetID() string {
	return "EMBED"
}

func (s *Embed) GetReader() (io.ReadCloser, error) {
	return s.FS.Open(s.Name)
}

func (s *Embed) GetPaths() ([]string, error) {
	return utils.Executables()
}

func (*Embed) GetImmutable() bool {
	return false
}

func (*Embed) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		close(ch)
	}
}
