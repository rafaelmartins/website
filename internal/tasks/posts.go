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

type PostsSources struct {
	Dir             string
	BaseDestination string
}

func (p *PostsSources) List() ([]*generators.ContentSource, error) {
	if p.Dir == "" {
		return nil, fmt.Errorf("posts: source dir not defined")
	}

	srcs, err := os.ReadDir(p.Dir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	rv := []*generators.ContentSource{}
	for _, src := range srcs {
		fpath := filepath.Join(p.Dir, src.Name())
		if !content.IsSupported(fpath) {
			continue
		}

		slug := strings.TrimSuffix(src.Name(), filepath.Ext(src.Name()))
		rv = append(rv,
			&generators.ContentSource{
				File: fpath,
				URL:  path.Join("/", p.BaseDestination, slug) + "/",
			},
		)
	}
	return rv, nil
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
	SourceDir   PostsSources
	Template    string
	TemplateCtx map[string]any
	WithSidebar bool
}

func (p *Posts) GetBaseDestination() string {
	return p.SourceDir.BaseDestination
}

func (p *Posts) GetTasks() ([]*runner.Task, error) {
	if p.SourceDir.Dir == "" {
		return nil, fmt.Errorf("posts: source dir not defined")
	}

	tmpl := p.Template
	if tmpl == "" {
		tmpl = "entry.html"
	}

	srcs, err := p.SourceDir.List()
	if err != nil {
		return nil, err
	}

	rv := []*runner.Task{}
	for _, src := range srcs {
		name := filepath.Base(src.File)
		slug := strings.TrimSuffix(name, filepath.Ext(name))
		rv = append(rv,
			runner.NewTask(p,
				&postTaskImpl{
					baseDestination: p.SourceDir.BaseDestination,
					slug:            slug,
					source:          src,
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
