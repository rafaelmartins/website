package tasks

import (
	"embed"
	"path/filepath"

	"rafaelmartins.com/p/website/internal/generators"
	"rafaelmartins.com/p/website/internal/runner"
)

type embedTaskImpl struct {
	fs   embed.FS
	name string
}

func (t *embedTaskImpl) GetDestination() string {
	return filepath.Base(t.name)
}

func (t *embedTaskImpl) GetGenerator() (runner.Generator, error) {
	return &generators.Embed{
		FS:   t.fs,
		Name: t.name,
	}, nil
}

type Embed struct {
	FS              embed.FS
	Directory       string
	BaseDestination string
}

func (e *Embed) GetBaseDestination() string {
	return e.BaseDestination
}

func (e *Embed) GetTasks() ([]*runner.Task, error) {
	dir := e.Directory
	if dir == "" {
		dir = "."
	}

	entries, err := e.FS.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	if dir == "." && len(entries) == 1 && entries[0].IsDir() {
		dir = entries[0].Name()
		entries, err = e.FS.ReadDir(dir)
		if err != nil {
			return nil, err
		}
	}

	rv := []*runner.Task{}
	for _, ee := range entries {
		if !ee.IsDir() {
			rv = append(rv, runner.NewTask(e, &embedTaskImpl{
				fs:   e.FS,
				name: filepath.Join(dir, ee.Name()),
			}))
		}
	}
	return rv, nil
}
