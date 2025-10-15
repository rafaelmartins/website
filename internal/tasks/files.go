package tasks

import (
	"os"
	"path/filepath"

	"rafaelmartins.com/p/website/internal/generators"
	"rafaelmartins.com/p/website/internal/runner"
)

type fileTaskImpl string

func (t fileTaskImpl) GetDestination() string {
	return filepath.Base(string(t))
}

func (t fileTaskImpl) GetGenerator() (runner.Generator, error) {
	return generators.File(t), nil
}

type Files struct {
	Paths           []string
	BaseDestination string
}

func (f *Files) GetBaseDestination() string {
	return f.BaseDestination
}

func (f *Files) GetTasks() ([]*runner.Task, error) {
	rv := []*runner.Task{}
	for _, p := range f.Paths {
		ps, err := filepath.Glob(p)
		if err != nil {
			return nil, err
		}
		for _, pp := range ps {
			st, err := os.Stat(pp)
			if err != nil {
				return nil, err
			}
			if st.Mode().IsRegular() {
				rv = append(rv, runner.NewTask(f, fileTaskImpl(pp)))
			}
		}
	}
	return rv, nil
}
