package tasks

import (
	"path"
	"path/filepath"
	"slices"
	"time"

	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
)

type postForAtom struct {
	Source *generators.MarkdownSource
	Date   time.Time
}

type atomTaskImpl struct {
	baseDestination string
	title           string
	sources         []*generators.MarkdownSource
	slug            string
	highlightStyle  string
	template        string
	templateCtx     map[string]interface{}
}

func (t *atomTaskImpl) GetDestination() string {
	return filepath.Join(t.slug, "atom.xml")
}

func (t *atomTaskImpl) GetGenerator() (runner.Generator, error) {
	return &generators.Markdown{
		Title:          t.title,
		URL:            path.Join("/", t.baseDestination, t.slug, "atom.xml"),
		Sources:        t.sources,
		IsPost:         true,
		HighlightStyle: t.highlightStyle,
		Template:       t.template,
		TemplateCtx:    t.templateCtx,
		Pagination:     &templates.ContentPagination{},
	}, nil
}

type Atom struct {
	Title           string
	Sources         []*generators.MarkdownSource
	PostsPerPage    int
	HighlightStyle  string
	BaseDestination string
	Template        string
	TemplateCtx     map[string]interface{}
}

func (p *Atom) GetBaseDestination() string {
	return p.BaseDestination
}

func (p *Atom) GetTasks() ([]*runner.Task, error) {
	if len(p.Sources) == 0 {
		return nil, nil
	}

	if p.PostsPerPage == 0 {
		return nil, nil
	}

	tmpl := p.Template
	if tmpl == "" {
		tmpl = "atom.xml"
	}

	style := p.HighlightStyle
	if style == "" {
		style = "github"
	}

	posts := []*postForAtom{}
	for _, src := range p.Sources {
		post := &postForAtom{
			Source: src,
		}

		dt, err := generators.MarkdownParseDate(post.Source.File)
		if err != nil {
			return nil, err
		}

		post.Date = dt
		posts = append(posts, post)
	}

	slices.SortStableFunc(posts, func(a *postForAtom, b *postForAtom) int {
		return b.Date.Compare(a.Date)
	})

	ppp := p.PostsPerPage
	if ppp < 0 {
		ppp = len(posts)
	}

	rv := []*runner.Task{}
	for chk := range slices.Chunk(posts, ppp) {
		srcs := []*generators.MarkdownSource{}
		for _, s := range chk {
			srcs = append(srcs, s.Source)
		}

		rv = append(rv,
			runner.NewTask(
				&atomTaskImpl{
					baseDestination: p.BaseDestination,
					title:           p.Title,
					sources:         srcs,
					slug:            "",
					highlightStyle:  style,
					template:        tmpl,
					templateCtx:     p.TemplateCtx,
				},
			),
		)
		break
	}

	if len(rv) == 0 {
		rv = append(rv,
			runner.NewTask(
				&atomTaskImpl{
					baseDestination: p.BaseDestination,
					title:           p.Title,
					sources:         nil,
					slug:            "",
					highlightStyle:  style,
					template:        tmpl,
					templateCtx:     p.TemplateCtx,
				},
			),
		)
	}

	return rv, nil
}
