package tasks

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
)

type postSource struct {
	source *generators.MarkdownSource
	slug   string
}

type postTaskImpl struct {
	baseDestination string
	slug            string
	source          *generators.MarkdownSource
	highlightStyle  string
	template        string
	templateCtx     map[string]interface{}
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

	return &generators.Markdown{
		URL:            url,
		Sources:        []*generators.MarkdownSource{t.source},
		IsPost:         true,
		HighlightStyle: t.highlightStyle,
		Template:       t.template,
		TemplateCtx:    t.templateCtx,
		LayoutCtx:      t.layoutCtx,

		OpenGraphImageGenerate: true,
	}, nil
}

type Posts struct {
	SourceDir       string
	HighlightStyle  string
	BaseDestination string
	Template        string
	TemplateCtx     map[string]interface{}
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
		if filepath.Ext(src.Name()) != ".md" {
			continue
		}

		slug := strings.TrimSuffix(src.Name(), ".md")
		rv = append(rv,
			&postSource{
				source: &generators.MarkdownSource{
					File: filepath.Join(p.SourceDir, src.Name()),
					URL:  path.Join("/", p.BaseDestination, slug) + "/",
				},
				slug: slug,
			},
		)
	}
	return rv, nil
}

func (p *Posts) GetSources() ([]*generators.MarkdownSource, error) {
	srcs, err := p.getSources()
	if err != nil {
		return nil, err
	}

	rv := []*generators.MarkdownSource{}
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

	style := p.HighlightStyle
	if style == "" {
		style = "github"
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
					highlightStyle:  style,
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
