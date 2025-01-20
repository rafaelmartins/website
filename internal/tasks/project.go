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

	openGraphTitle         string
	openGraphDescription   string
	openGraphImage         string
	openGraphImageGenColor *uint32
	openGraphImageGenDPI   *float64
	openGraphImageGenSize  *float64
}

func (t *projectTaskImpl) GetDestination() string {
	return filepath.Join(t.repo, "index.html")
}

func (t *projectTaskImpl) GetGenerator() (runner.Generator, error) {
	url := path.Join("/", t.baseDestination, t.repo)
	if url != "/" {
		url += "/"
	}

	return &generators.Project{
		Owner:     t.owner,
		Repo:      t.repo,
		URL:       url,
		Template:  t.template,
		Immutable: t.immutable,
		LayoutCtx: t.layoutCtx,

		OpenGraphTitle:         t.openGraphTitle,
		OpenGraphDescription:   t.openGraphDescription,
		OpenGraphImage:         t.openGraphImage,
		OpenGraphImageGenColor: t.openGraphImageGenColor,
		OpenGraphImageGenDPI:   t.openGraphImageGenDPI,
		OpenGraphImageGenSize:  t.openGraphImageGenSize,
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

	openGraphTitle         string
	openGraphDescription   string
	openGraphImage         string
	openGraphImageGenColor *uint32
	openGraphImageGenDPI   *float64
	openGraphImageGenSize  *float64
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
		URL:           path.Join("/", t.baseDestination, t.repo, dest) + "/",
		Template:      t.template,
		LayoutCtx:     t.layoutCtx,
		Immutable:     t.immutable,

		OpenGraphTitle:         t.openGraphTitle,
		OpenGraphDescription:   t.openGraphDescription,
		OpenGraphImage:         t.openGraphImage,
		OpenGraphImageGenColor: t.openGraphImageGenColor,
		OpenGraphImageGenDPI:   t.openGraphImageGenDPI,
		OpenGraphImageGenSize:  t.openGraphImageGenSize,
	}, nil
}

type Project struct {
	Owner string
	Repo  string

	CDocsDestination            string
	CDocsHeaders                []string
	CDocsBaseDirectory          string
	CDocsTemplate               string
	CDocsWithSidebar            bool
	CDocsOpenGraphTitle         string
	CDocsOpenGraphDescription   string
	CDocsOpenGraphImage         string
	CDocsOpenGraphImageGenColor *uint32
	CDocsOpenGraphImageGenDPI   *float64
	CDocsOpenGraphImageGenSize  *float64

	BaseDestination        string
	Template               string
	Immutable              bool
	WithSidebar            bool
	OpenGraphTitle         string
	OpenGraphDescription   string
	OpenGraphImage         string
	OpenGraphImageGenColor *uint32
	OpenGraphImageGenDPI   *float64
	OpenGraphImageGenSize  *float64
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

				openGraphTitle:         p.OpenGraphTitle,
				openGraphDescription:   p.OpenGraphDescription,
				openGraphImage:         p.OpenGraphImage,
				openGraphImageGenColor: p.OpenGraphImageGenColor,
				openGraphImageGenDPI:   p.OpenGraphImageGenDPI,
				openGraphImageGenSize:  p.OpenGraphImageGenSize,
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

					openGraphTitle:         p.CDocsOpenGraphTitle,
					openGraphDescription:   p.CDocsOpenGraphDescription,
					openGraphImage:         p.CDocsOpenGraphImage,
					openGraphImageGenColor: p.CDocsOpenGraphImageGenColor,
					openGraphImageGenDPI:   p.CDocsOpenGraphImageGenDPI,
					openGraphImageGenSize:  p.CDocsOpenGraphImageGenSize,
				},
			),
		)
	}

	return rv, nil
}
