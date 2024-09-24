package generators

import (
	"embed"
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/utils"
)

type Embed struct {
	FS   *embed.FS
	Name string
}

func (*Embed) GetID() string {
	return "EMBED"
}

func (s *Embed) GetReader() (io.ReadCloser, error) {
	if s.FS == nil {
		return nil, errors.New("generators: embed: FS is nil")
	}
	return s.FS.Open(s.Name)
}

func (s *Embed) GetTimeStamps() ([]time.Time, error) {
	return utils.ExecutableTimestamps()
}

func (*Embed) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		close(ch)
	}
}

func EmbedListFiles(efs *embed.FS) ([]string, error) {
	var process func(ent []fs.DirEntry, dir string) error

	rv := []string{}
	process = func(ent []fs.DirEntry, dir string) error {
		for _, entry := range ent {
			if entry.IsDir() {
				d := filepath.Join(dir, entry.Name())
				e, err := efs.ReadDir(d)
				if err != nil {
					return err
				}
				return process(e, d)
			}
			rv = append(rv, filepath.Join(dir, entry.Name()))
		}
		return nil
	}

	entries, err := efs.ReadDir(".")
	if err != nil {
		return nil, err
	}
	if err := process(entries, ""); err != nil {
		return nil, err
	}
	return rv, nil
}
