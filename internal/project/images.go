package project

import (
	"path/filepath"

	"rafaelmartins.com/p/website/internal/generators"
	"rafaelmartins.com/p/website/internal/runner"
)

type imageTask struct {
	proj *Project
	path string
}

func (i *imageTask) GetDestination() string {
	return filepath.Join(i.proj.Repo, filepath.FromSlash(string(i.path)))
}

func (i *imageTask) GetGenerator() (runner.Generator, error) {
	if i.proj.LocalDirectory != nil {
		return generators.File(filepath.Join(*i.proj.LocalDirectory, i.path)), nil
	}

	return &generators.GithubFile{
		Owner:     i.proj.Owner,
		Repo:      i.proj.Repo,
		Ref:       i.proj.proj.Head,
		Path:      i.path,
		Immutable: i.proj.Immutable,
	}, nil
}
