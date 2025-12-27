package tasks

import (
	"path"
	"path/filepath"

	"rafaelmartins.com/p/website/internal/generators"
	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/templates"
)

type pageTaskImpl struct {
	baseDestination   string
	slug              string
	source            string
	extraDependencies []string
	prettyURL         bool
	template          string
	templateCtx       map[string]any
	layoutCtx         *templates.LayoutContext

	openGraphTitle         string
	openGraphDescription   string
	openGraphImage         string
	openGraphImageGenerate bool
	openGraphImageGenColor *uint32
	openGraphImageGenDPI   *float64
	openGraphImageGenSize  *float64
}

func (t *pageTaskImpl) GetDestination() string {
	if t.prettyURL {
		return filepath.Join(t.slug, "index.html")
	}
	return t.slug + ".html"
}

func (t *pageTaskImpl) GetGenerator() (runner.Generator, error) {
	url := path.Join("/", t.baseDestination, t.slug) + ".html"
	if t.prettyURL {
		url = path.Join("/", t.baseDestination, t.slug)
		if url != "/" {
			url += "/"
		}
	}

	return &generators.Content{
		URL:  url,
		Slug: t.slug,
		Sources: []*generators.ContentSource{
			{
				File: t.source,
				URL:  url,
			},
		},
		ExtraDependencies: t.extraDependencies,
		Template:          t.template,
		TemplateCtx:       t.templateCtx,
		LayoutCtx:         t.layoutCtx,

		OpenGraphTitle:         t.openGraphTitle,
		OpenGraphDescription:   t.openGraphDescription,
		OpenGraphImage:         t.openGraphImage,
		OpenGraphImageGenerate: t.openGraphImageGenerate,
		OpenGraphImageGenColor: t.openGraphImageGenColor,
		OpenGraphImageGenDPI:   t.openGraphImageGenDPI,
		OpenGraphImageGenSize:  t.openGraphImageGenSize,
	}, nil
}

type PageSource struct {
	Slug string
	File string

	OpenGraphTitle         string
	OpenGraphDescription   string
	OpenGraphImage         string
	OpenGraphImageGenerate bool
	OpenGraphImageGenColor *uint32
	OpenGraphImageGenDPI   *float64
	OpenGraphImageGenSize  *float64
}

type Pages struct {
	Sources           []*PageSource
	ExtraDependencies []string
	PrettyURL         bool
	BaseDestination   string
	Template          string
	TemplateCtx       map[string]any
	WithSidebar       bool
}

func (p *Pages) GetBaseDestination() string {
	return p.BaseDestination
}

func (p *Pages) GetTasks() ([]*runner.Task, error) {
	tmpl := p.Template
	if tmpl == "" {
		tmpl = "entry.html"
	}

	deps := []string{}
	for _, dep := range p.ExtraDependencies {
		gdeps, err := filepath.Glob(dep)
		if err != nil {
			return nil, err
		}
		deps = append(deps, gdeps...)
	}

	rv := []*runner.Task{}
	for _, v := range p.Sources {
		rv = append(rv,
			runner.NewTask(p,
				&pageTaskImpl{
					baseDestination:   p.BaseDestination,
					slug:              v.Slug,
					source:            v.File,
					extraDependencies: deps,
					prettyURL:         p.PrettyURL,
					template:          tmpl,
					templateCtx:       p.TemplateCtx,
					layoutCtx: &templates.LayoutContext{
						WithSidebar: p.WithSidebar,
					},

					openGraphTitle:         v.OpenGraphTitle,
					openGraphDescription:   v.OpenGraphDescription,
					openGraphImage:         v.OpenGraphImage,
					openGraphImageGenerate: v.OpenGraphImageGenerate,
					openGraphImageGenColor: v.OpenGraphImageGenColor,
					openGraphImageGenDPI:   v.OpenGraphImageGenDPI,
					openGraphImageGenSize:  v.OpenGraphImageGenSize,
				},
			),
		)
	}
	return rv, nil
}
