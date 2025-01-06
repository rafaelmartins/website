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

type cdocsTaskImpl struct {
	baseDestination string
	owner           string
	repo            string
	destination     string
	headers         []string
	basedir         string
	template        string
	immutable       bool
	layoutCtx       *templates.LayoutContext
}

func (t *cdocsTaskImpl) GetDestination() string {
	dest := t.destination
	if dest == "" {
		dest = "api"
	}
	return filepath.Join(t.repo, dest, "index.html")
}

func (t *cdocsTaskImpl) GetGenerator() (runner.Generator, error) {
	dest := t.destination
	if dest == "" {
		dest = "api"
	}
	return &generators.CDocs{
		Owner:         t.owner,
		Repo:          t.repo,
		Headers:       t.headers,
		BaseDirectory: t.basedir,
		URL:           path.Join("/", t.baseDestination, t.repo, dest, "index.html"),
		Template:      t.template,
		LayoutCtx:     t.layoutCtx,
		Immutable:     t.immutable,
	}, nil
}

type Project struct {
	Owner string
	Repo  string

	CDocsDestination   string
	CDocsHeaders       []string
	CDocsBaseDirectory string
	CDocsTemplate      string
	CDocsWithSidebar   bool
	BaseDestination    string

	Template    string
	Immutable   bool
	WithSidebar bool
}

func (p *Project) GetBaseDestination() string {
	if p.BaseDestination == "" {
		return "projects"
	}
	return p.BaseDestination
}

func (p *Project) GetTasks() ([]*runner.Task, error) {
	tmpl := p.Template
	if tmpl == "" {
		tmpl = "project.html"
	}

	rv := []*runner.Task{
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
	}

	if len(p.CDocsHeaders) > 0 {
		dtmpl := p.CDocsTemplate
		if dtmpl == "" {
			dtmpl = "cdocs.html"
		}

		rv = append(rv,
			runner.NewTask(p,
				&cdocsTaskImpl{
					baseDestination: p.GetBaseDestination(),
					owner:           p.Owner,
					repo:            p.Repo,
					destination:     p.CDocsDestination,
					headers:         p.CDocsHeaders,
					basedir:         p.CDocsBaseDirectory,
					template:        dtmpl,
					immutable:       p.Immutable,
					layoutCtx: &templates.LayoutContext{
						WithSidebar: p.CDocsWithSidebar,
					},
				},
			),
		)
	}

	return rv, nil
}
