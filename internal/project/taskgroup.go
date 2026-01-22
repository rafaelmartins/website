package project

import (
	"path/filepath"
	"slices"

	"rafaelmartins.com/p/website/internal/runner"
)

func (p *Project) GetBaseDestination() string {
	if p.BaseDestination == "" {
		return "projects"
	}
	return p.BaseDestination
}

func (p *Project) GetSkipIfExists() *string {
	if p.LocalDirectory != nil {
		return nil
	}

	rv := filepath.Join(p.GetBaseDestination(), p.Repo, "index.html")
	return &rv
}

func (p *Project) GetTasks() ([]*runner.Task, error) {
	if err := p.init(); err != nil {
		return nil, err
	}

	rv := []*runner.Task{}
	images := []string{}
	for _, page := range p.pages {
		rv = append(rv, runner.NewTask(p, page))
		images = append(images, page.images...)
	}
	if len(p.CDocsHeaders) > 0 {
		rv = append(rv, runner.NewTask(p, &cDocs{proj: p}))
	}

	slices.Sort(images)
	for _, img := range slices.Compact(images) {
		rv = append(rv, runner.NewTask(p, &imageTask{
			proj: p,
			path: img,
		}))
	}
	return rv, nil
}
