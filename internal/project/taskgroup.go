package project

import (
	"path/filepath"

	"rafaelmartins.com/p/website/internal/runner"
)

func (p *Project) GetBaseDestination() string {
	if p.BaseDestination == "" {
		return "projects"
	}
	return p.BaseDestination
}

func (p *Project) GetSkipIfExists() string {
	return filepath.Join(p.GetBaseDestination(), p.Repo, "index.html")
}

func (p *Project) GetTasks() ([]*runner.Task, error) {
	if err := p.init(); err != nil {
		return nil, err
	}

	rv := []*runner.Task{}
	for _, page := range p.pages {
		rv = append(rv, runner.NewTask(p, page))
	}
	if len(p.CDocsHeaders) > 0 {
		rv = append(rv, runner.NewTask(p, &cDocs{proj: p}))
	}
	return rv, nil
}
