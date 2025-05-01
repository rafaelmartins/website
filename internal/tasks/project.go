package tasks

import (
	"path"
	"path/filepath"

	"rafaelmartins.com/p/website/internal/generators"
	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/templates"
)

type projectTaskImpl struct {
	baseDestination string
	owner           string
	repo            string
	subPage         string
	template        string
	immutable       bool
	layoutCtx       *templates.LayoutContext

	subPages []string
	docLinks []*templates.ProjectContentDocLink

	goImport string
	goRepo   string

	cdocsEnabled     bool
	cdocsDestination string

	openGraphTitle         string
	openGraphDescription   string
	openGraphImage         string
	openGraphImageGenColor *uint32
	openGraphImageGenDPI   *float64
	openGraphImageGenSize  *float64
}

func (t *projectTaskImpl) GetDestination() string {
	if t.subPage != "" {
		return filepath.Join(t.repo, t.subPage, "index.html")
	}
	return filepath.Join(t.repo, "index.html")
}

func (t *projectTaskImpl) GetGenerator() (runner.Generator, error) {
	url := path.Join("/", t.baseDestination, t.repo)
	if t.subPage != "" {
		url = path.Join(url, t.subPage)
	}
	if url != "/" {
		url += "/"
	}
	cdocsUrl := ""
	if t.cdocsEnabled {
		cdocsUrl = path.Join("/", t.baseDestination, t.repo, t.cdocsDestination)
		if cdocsUrl != "/" {
			cdocsUrl += "/"
		}
	}

	return &generators.Project{
		Owner:   t.owner,
		Repo:    t.repo,
		SubPage: t.subPage,

		SubPages: t.subPages,
		DocLinks: t.docLinks,

		GoImport: t.goImport,
		GoRepo:   t.goRepo,

		CDocsURL: cdocsUrl,

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
	localdir        string
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
	return filepath.Join(t.repo, t.destination, "index.html")
}

func (t *cdocsTaskImpl) GetGenerator() (runner.Generator, error) {
	dest := t.destination
	if dest == "" {
		dest = "api"
	}

	return &generators.CDocs{
		Owner:          t.owner,
		Repo:           t.repo,
		Headers:        t.headers,
		BaseDirectory:  t.basedir,
		LocalDirectory: t.localdir,
		URL:            path.Join("/", t.baseDestination, t.repo, dest) + "/",
		Template:       t.template,
		LayoutCtx:      t.layoutCtx,
		Immutable:      t.immutable,

		OpenGraphTitle:         t.openGraphTitle,
		OpenGraphDescription:   t.openGraphDescription,
		OpenGraphImage:         t.openGraphImage,
		OpenGraphImageGenColor: t.openGraphImageGenColor,
		OpenGraphImageGenDPI:   t.openGraphImageGenDPI,
		OpenGraphImageGenSize:  t.openGraphImageGenSize,
	}, nil
}

type ProjectSubPage struct {
	SubPage                string
	Template               string
	WithSidebar            bool
	OpenGraphTitle         string
	OpenGraphDescription   string
	OpenGraphImage         string
	OpenGraphImageGenColor *uint32
	OpenGraphImageGenDPI   *float64
	OpenGraphImageGenSize  *float64
}

type Project struct {
	Owner string
	Repo  string

	SubPages []*ProjectSubPage
	DocLinks []*templates.ProjectContentDocLink

	GoImport string
	GoRepo   string

	CDocsDestination            string
	CDocsHeaders                []string
	CDocsBaseDirectory          string
	CDocsLocalDirectory         string
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

	cdocsDestination := p.CDocsDestination
	if cdocsDestination == "" {
		cdocsDestination = "api"
	}

	subpages := []string{}
	for _, subpage := range p.SubPages {
		subpages = append(subpages, subpage.SubPage)
	}

	rv := []*runner.Task{
		runner.NewTask(p,
			&projectTaskImpl{
				baseDestination:  p.GetBaseDestination(),
				owner:            p.Owner,
				repo:             p.Repo,
				subPages:         subpages,
				docLinks:         p.DocLinks,
				goImport:         p.GoImport,
				goRepo:           p.GoRepo,
				cdocsDestination: cdocsDestination,
				cdocsEnabled:     len(p.CDocsHeaders) > 0,
				template:         tmpl,
				immutable:        p.Immutable,
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
					destination:     cdocsDestination,
					headers:         p.CDocsHeaders,
					basedir:         p.CDocsBaseDirectory,
					localdir:        p.CDocsLocalDirectory,
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

	for _, subpage := range p.SubPages {
		tmpl := subpage.Template
		if tmpl == "" {
			tmpl = "project.html"
		}

		rv = append(rv,
			runner.NewTask(p,
				&projectTaskImpl{
					baseDestination:  p.GetBaseDestination(),
					owner:            p.Owner,
					repo:             p.Repo,
					subPage:          subpage.SubPage,
					subPages:         subpages,
					docLinks:         p.DocLinks,
					cdocsDestination: cdocsDestination,
					cdocsEnabled:     len(p.CDocsHeaders) > 0,
					template:         tmpl,
					immutable:        p.Immutable,
					layoutCtx: &templates.LayoutContext{
						WithSidebar: subpage.WithSidebar,
					},

					openGraphTitle:         subpage.OpenGraphTitle,
					openGraphDescription:   subpage.OpenGraphDescription,
					openGraphImage:         subpage.OpenGraphImage,
					openGraphImageGenColor: subpage.OpenGraphImageGenColor,
					openGraphImageGenDPI:   subpage.OpenGraphImageGenDPI,
					openGraphImageGenSize:  subpage.OpenGraphImageGenSize,
				},
			),
		)
	}

	return rv, nil
}
