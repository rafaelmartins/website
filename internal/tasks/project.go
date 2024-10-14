package tasks

import (
	"path"
	"path/filepath"

	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
)

type projectTaskImpl struct {
	baseDestination string
	owner           string
	repo            string
	template        string
	immutable       bool
	layoutCtx       *templates.LayoutContext
}

func (t *projectTaskImpl) GetDestination() string {
	return filepath.Join(t.repo, "index.html")
}

func (t *projectTaskImpl) GetGenerator() (runner.Generator, error) {
	return &generators.Project{
		Owner:     t.owner,
		Repo:      t.repo,
		URL:       path.Join("/", t.baseDestination, t.repo, "index.html"),
		Template:  t.template,
		Immutable: t.immutable,
		LayoutCtx: t.layoutCtx,
	}, nil
}

type Project struct {
	Owner           string
	Repo            string
	BaseDestination string
	Template        string
	Immutable       bool
	WithSidebar     bool
}

func (p *Project) GetBaseDestination() string {
	if p.BaseDestination == "" {
		return "project"
	}
	return p.BaseDestination
}

func (p *Project) GetTasks() ([]*runner.Task, error) {
	tmpl := p.Template
	if tmpl == "" {
		tmpl = "project.html"
	}

	return []*runner.Task{
		runner.NewTask(p,
			&projectTaskImpl{
				baseDestination: p.GetBaseDestination(),
				owner:           p.Owner,
				repo:            p.Repo,
				template:        tmpl,
				immutable:       p.Immutable,
				layoutCtx: &templates.LayoutContext{
					WithSidebar: p.WithSidebar,
				},
			},
		),
	}, nil
}
