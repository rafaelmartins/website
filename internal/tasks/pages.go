package tasks

import (
	"path"
	"path/filepath"

	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
)

type pageTaskImpl struct {
	baseDestination   string
	slug              string
	source            string
	extraDependencies []string
	highlightStyle    string
	template          string
	templateCtx       map[string]interface{}
	layoutCtx         *templates.LayoutContext
}

func (t *pageTaskImpl) GetDestination() string {
	return filepath.Join(t.slug, "index.html")
}

func (t *pageTaskImpl) GetGenerator() (runner.Generator, error) {
	return &generators.Markdown{
		URL: path.Join("/", t.baseDestination, t.slug, "index.html"),
		Sources: []*generators.MarkdownSource{
			{
				File: t.source,
				Slug: t.slug,
			},
		},
		ExtraDependencies: t.extraDependencies,
		HighlightStyle:    t.highlightStyle,
		Template:          t.template,
		TemplateCtx:       t.templateCtx,
		LayoutCtx:         t.layoutCtx,
	}, nil
}

type Pages struct {
	Sources           map[string]string
	ExtraDependencies []string
	HighlightStyle    string
	BaseDestination   string
	Template          string
	TemplateCtx       map[string]interface{}
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

	style := p.HighlightStyle
	if style == "" {
		style = "github"
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
	for k, v := range p.Sources {
		rv = append(rv,
			runner.NewTask(
				&pageTaskImpl{
					baseDestination:   p.BaseDestination,
					slug:              k,
					source:            v,
					extraDependencies: deps,
					highlightStyle:    style,
					template:          tmpl,
					templateCtx:       p.TemplateCtx,
					layoutCtx: &templates.LayoutContext{
						WithSidebar: p.WithSidebar,
					},
				},
			),
		)
	}
	return rv, nil
}
