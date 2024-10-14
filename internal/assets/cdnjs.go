package assets

import (
	"fmt"
	"path/filepath"

	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/runner"
)

type cdnjsTask struct {
	name    string
	version string
	file    string
}

func (t *cdnjsTask) GetDestination() string {
	return filepath.FromSlash(t.file)
}

func (t cdnjsTask) GetGenerator() (runner.Generator, error) {
	return &generators.HTTP{
		Url:       fmt.Sprintf("https://cdnjs.cloudflare.com/ajax/libs/%s/%s/%s", t.name, t.version, t.file),
		Immutable: true,
	}, nil
}

type CdnjsLibrary struct {
	Name            string
	Version         string
	Files           []string
	BaseDestination string
}

func (l *CdnjsLibrary) GetBaseDestination() string {
	return filepath.Join(l.BaseDestination, l.Name)
}

func (l *CdnjsLibrary) GetTasks() ([]*runner.Task, error) {
	rv := []*runner.Task{}
	for _, f := range l.Files {
		rv = append(rv,
			runner.NewTask(l,
				&cdnjsTask{
					name:    l.Name,
					version: l.Version,
					file:    f,
				},
			),
		)
	}
	return rv, nil
}
