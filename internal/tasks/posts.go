package tasks

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"rafaelmartins.com/p/website/internal/content"
	"rafaelmartins.com/p/website/internal/generators"
	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/templates"
)

type postSource struct {
	source *generators.ContentSource
	slug   string
}

type postTaskImpl struct {
	baseDestination string
	slug            string
	source          *generators.ContentSource
	template        string
	templateCtx     map[string]any
	layoutCtx       *templates.LayoutContext
}

func (t *postTaskImpl) GetDestination() string {
	return filepath.Join(t.slug, "index.html")
}

func (t *postTaskImpl) GetGenerator() (runner.Generator, error) {
	url := path.Join("/", t.baseDestination, t.slug)
	if url != "/" {
		url += "/"
	}

	return &generators.Content{
		URL:         url,
		Slug:        t.slug,
		Sources:     []*generators.ContentSource{t.source},
		IsPost:      true,
		Template:    t.template,
		TemplateCtx: t.templateCtx,
		LayoutCtx:   t.layoutCtx,

		OpenGraphImageGenerate: true,
	}, nil
}

type Posts struct {
	SourceDir       string
	BaseDestination string
	Template        string
	TemplateCtx     map[string]any
	WithSidebar     bool
}

func (p *Posts) GetBaseDestination() string {
	return p.BaseDestination
}

func (p *Posts) getSources() ([]*postSource, error) {
	if p.SourceDir == "" {
		return nil, fmt.Errorf("posts: source dir not defined")
	}

	srcs, err := os.ReadDir(p.SourceDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	rv := []*postSource{}
	for _, src := range srcs {
		fpath := filepath.Join(p.SourceDir, src.Name())
		if !content.IsSupported(fpath) {
			continue
		}

		slug := strings.TrimSuffix(src.Name(), filepath.Ext(src.Name()))
		rv = append(rv,
			&postSource{
				source: &generators.ContentSource{
					File: fpath,
					URL:  path.Join("/", p.BaseDestination, slug) + "/",
				},
				slug: slug,
			},
		)
	}
	return rv, nil
}

func (p *Posts) GetSources() ([]*generators.ContentSource, error) {
	srcs, err := p.getSources()
	if err != nil {
		return nil, err
	}

	rv := []*generators.ContentSource{}
	for _, src := range srcs {
		rv = append(rv, src.source)
	}
	return rv, nil
}

func (p *Posts) GetTasks() ([]*runner.Task, error) {
	if p.SourceDir == "" {
		return nil, fmt.Errorf("posts: source dir not defined")
	}

	tmpl := p.Template
	if tmpl == "" {
		tmpl = "entry.html"
	}

	srcs, err := p.getSources()
	if err != nil {
		return nil, err
	}

	rv := []*runner.Task{}
	for _, src := range srcs {
		rv = append(rv,
			runner.NewTask(p,
				&postTaskImpl{
					baseDestination: p.BaseDestination,
					slug:            src.slug,
					source:          src.source,
					template:        tmpl,
					templateCtx:     p.TemplateCtx,
					layoutCtx: &templates.LayoutContext{
						WithSidebar: p.WithSidebar,
					},
				},
			),
		)
	}
	return rv, nil
}
