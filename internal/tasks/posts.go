package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
)

type postTaskImpl struct {
	slug           string
	source         string
	highlightStyle string
	template       string
	templateCtx    map[string]interface{}
	layoutCtx      *templates.LayoutContext
}

func (t *postTaskImpl) GetDestination() string {
	return filepath.Join(t.slug, "index.html")
}

func (t *postTaskImpl) GetGenerator() (runner.Generator, error) {
	return &generators.Markdown{
		Sources: []*generators.MarkdownSource{
			{
				File: t.source,
				Slug: t.slug,
			},
		},
		IsPost:         true,
		HighlightStyle: t.highlightStyle,
		Template:       t.template,
		TemplateCtx:    t.templateCtx,
		LayoutCtx:      t.layoutCtx,
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

	srcs, err := os.ReadDir(p.SourceDir)
	if err != nil {
		return nil, err
	}

	rv := []*runner.Task{}
	for _, src := range srcs {
		if filepath.Ext(src.Name()) != ".md" {
			continue
		}

		rv = append(rv,
			runner.NewTask(
				&postTaskImpl{
					slug:           strings.TrimSuffix(src.Name(), ".md"),
					source:         filepath.Join(p.SourceDir, src.Name()),
					highlightStyle: style,
					template:       tmpl,
					templateCtx:    p.TemplateCtx,
					layoutCtx: &templates.LayoutContext{
						WithSidebar: p.WithSidebar,
					},
				},
			),
		)
	}
	return rv, nil
}
