package tasks

import (
	"path"
	"path/filepath"

	"rafaelmartins.com/p/website/internal/generators"
	"rafaelmartins.com/p/website/internal/opengraph"
	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/templates"
)

type pageTaskImpl struct {
	baseDestination   string
	title             string
	description       string
	slug              string
	source            string
	license           string
	toc               bool
	search            *bool
	extraDependencies []string
	prettyURL         bool
	template          string
	templateCtx       map[string]any
	layoutCtx         *templates.LayoutContext

	openGraph         *opengraph.Config
	openGraphImageGen *opengraph.OpenGraphImageGen
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
		Title:       t.title,
		Description: t.description,
		URL:         url,
		Slug:        t.slug,
		License:     t.license,
		Toc:         t.toc,
		Search:      t.search,
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

		OpenGraph:         t.openGraph,
		OpenGraphImageGen: t.openGraphImageGen,
	}, nil
}

type PageSource struct {
	Title       string
	Description string
	Slug        string
	File        string
	License     string
	Toc         bool
	Search      *bool
	OpenGraph   *opengraph.Config
}

type Pages struct {
	Sources           []*PageSource
	ExtraDependencies []string
	PrettyURL         bool
	BaseDestination   string
	Template          string
	TemplateCtx       map[string]any
	WithSidebar       bool
	OpenGraphImageGen *opengraph.OpenGraphImageGen
}

func (p *Pages) GetBaseDestination() string {
	return p.BaseDestination
}

func (p *Pages) GetTasks() ([]*runner.Task, error) {
	tmpl := p.Template
	if tmpl == "" {
		tmpl = "base.html"
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
					title:             v.Title,
					description:       v.Description,
					slug:              v.Slug,
					source:            v.File,
					license:           v.License,
					toc:               v.Toc,
					search:            v.Search,
					extraDependencies: deps,
					prettyURL:         p.PrettyURL,
					template:          tmpl,
					templateCtx:       p.TemplateCtx,
					layoutCtx: &templates.LayoutContext{
						WithSidebar: p.WithSidebar,
					},
					openGraph:         v.OpenGraph,
					openGraphImageGen: p.OpenGraphImageGen,
				},
			),
		)
	}
	return rv, nil
}
