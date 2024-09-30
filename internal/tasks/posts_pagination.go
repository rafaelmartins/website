package tasks

import (
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
)

type postForPagination struct {
	Slug string
	File string
	Date time.Time
}

type postPaginationTaskImpl struct {
	baseDestination string
	title           string
	sources         []*generators.MarkdownSource
	slug            string
	highlightStyle  string
	template        string
	templateCtx     map[string]interface{}
	pagination      *templates.ContentPagination
	layoutCtx       *templates.LayoutContext
}

func (t *postPaginationTaskImpl) GetDestination() string {
	return filepath.Join(t.slug, "index.html")
}

func (t *postPaginationTaskImpl) GetGenerator() (runner.Generator, error) {
	return &generators.Markdown{
		Title:          t.title,
		URL:            path.Join("/", t.baseDestination, t.slug, "index.html"),
		Sources:        t.sources,
		IsPost:         true,
		HighlightStyle: t.highlightStyle,
		Template:       t.template,
		TemplateCtx:    t.templateCtx,
		Pagination:     t.pagination,
		LayoutCtx:      t.layoutCtx,
	}, nil
}

type PostsPagination struct {
	Title           string
	SourceDir       string
	PostsPerPage    int
	SortReverse     bool
	HighlightStyle  string
	BaseDestination string
	Template        string
	TemplateCtx     map[string]interface{}
	WithSidebar     bool
}

func (p *PostsPagination) GetBaseDestination() string {
	return p.BaseDestination
}

func (p *PostsPagination) GetTasks() ([]*runner.Task, error) {
	if p.SourceDir == "" {
		return nil, fmt.Errorf("posts: source dir not defined")
	}

	if p.PostsPerPage == 0 {
		return nil, nil
	}

	tmpl := p.Template
	if tmpl == "" {
		tmpl = "pagination.html"
	}

	style := p.HighlightStyle
	if style == "" {
		style = "github"
	}

	srcs, err := os.ReadDir(p.SourceDir)
	if err != nil {
		return nil, err
	}

	posts := []*postForPagination{}
	for _, src := range srcs {
		if filepath.Ext(src.Name()) != ".md" {
			continue
		}

		post := &postForPagination{
			Slug: strings.TrimSuffix(src.Name(), ".md"),
			File: filepath.Join(p.SourceDir, src.Name()),
		}

		dt, err := generators.MarkdownParseDate(post.File)
		if err != nil {
			return nil, err
		}

		post.Date = dt
		posts = append(posts, post)
	}

	slices.SortStableFunc(posts, func(a *postForPagination, b *postForPagination) int {
		if p.SortReverse {
			return b.Date.Compare(a.Date)
		}
		return a.Date.Compare(b.Date)
	})

	ppp := p.PostsPerPage
	if ppp < 0 {
		ppp = len(posts)
	}

	layoutCtx := &templates.LayoutContext{
		WithSidebar: p.WithSidebar,
	}

	page := 1
	total := int(math.Ceil(float64(len(posts)) / float64(ppp)))

	if len(posts) == 0 {
		return []*runner.Task{
			runner.NewTask(
				&postPaginationTaskImpl{
					baseDestination: p.BaseDestination,
					title:           p.Title,
					sources:         nil,
					slug:            "",
					highlightStyle:  style,
					template:        tmpl,
					templateCtx:     p.TemplateCtx,
					pagination:      &templates.ContentPagination{},
					layoutCtx:       layoutCtx,
				},
			),
		}, nil
	}

	rv := []*runner.Task{}
	for chk := range slices.Chunk(posts, ppp) {
		srcs := []*generators.MarkdownSource{}
		for _, s := range chk {
			srcs = append(srcs,
				&generators.MarkdownSource{
					File: s.File,
					Slug: path.Join(p.BaseDestination, s.Slug),
				},
			)
		}

		pagination := &templates.ContentPagination{
			BaseURL: path.Join("/", p.BaseDestination, "page"),
			Current: page,
			Total:   total,
		}
		if page > 1 {
			pagination.LinkFirst = path.Join(pagination.BaseURL, "1") + "/"
		}
		if page < total {
			pagination.LinkLast = path.Join(pagination.BaseURL, strconv.FormatInt(int64(total), 10)) + "/"
		}

		if page == 1 {
			rv = append(rv,
				runner.NewTask(
					&postPaginationTaskImpl{
						baseDestination: p.BaseDestination,
						title:           p.Title,
						sources:         srcs,
						slug:            "",
						highlightStyle:  style,
						template:        tmpl,
						templateCtx:     p.TemplateCtx,
						pagination:      pagination,
						layoutCtx:       layoutCtx,
					},
				),
			)
		}
		rv = append(rv,
			runner.NewTask(
				&postPaginationTaskImpl{
					baseDestination: p.BaseDestination,
					title:           p.Title,
					sources:         srcs,
					slug:            path.Join("page", strconv.FormatInt(int64(page), 10)),
					highlightStyle:  style,
					template:        tmpl,
					templateCtx:     p.TemplateCtx,
					pagination:      pagination,
					layoutCtx:       layoutCtx,
				},
			),
		)
		page++
	}

	return rv, nil
}
