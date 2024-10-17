package tasks

import (
	"fmt"
	"path/filepath"

	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/runner"
)

type npmTask struct {
	name    string
	version string
	file    string
}

func (t *npmTask) GetDestination() string {
	return filepath.FromSlash(t.file)
}

func (t npmTask) GetGenerator() (runner.Generator, error) {
	return &generators.HTTP{
		Url:       fmt.Sprintf("https://cdn.jsdelivr.net/npm/%s@%s/%s", t.name, t.version, t.file),
		Immutable: true,
	}, nil
}

type NpmPackage struct {
	Name            string
	Version         string
	Files           []string
	BaseDestination string
}

func (p *NpmPackage) GetBaseDestination() string {
	return filepath.Join(p.BaseDestination, p.Name)
}

func (p *NpmPackage) GetTasks() ([]*runner.Task, error) {
	rv := []*runner.Task{}
	for _, f := range p.Files {
		rv = append(rv,
			runner.NewTask(p,
				&npmTask{
					name:    p.Name,
					version: p.Version,
					file:    f,
				},
			),
		)
	}
	return rv, nil
}
