package tasks

import (
	"math"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/ogimage"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
)

type paginationPost struct {
	Source *generators.MarkdownSource
	Date   time.Time
}

type paginationTaskImpl struct {
	atom            bool
	baseDestination string
	title           string
	description     string
	sources         []*generators.MarkdownSource
	slug            string
	highlightStyle  string
	template        string
	templateCtx     map[string]interface{}
	pagination      *templates.ContentPagination
	layoutCtx       *templates.LayoutContext

	openGraphTitle         string
	openGraphDescription   string
	openGraphImage         string
	openGraphImageURL      string
	openGraphImageGenerate bool
	openGraphImageGenColor *uint32
	openGraphImageGenDPI   *float64
	openGraphImageGenSize  *float64
}

func (t *paginationTaskImpl) GetDestination() string {
	if t.atom {
		return filepath.Join(t.slug, "atom.xml")
	}
	return filepath.Join(t.slug, "index.html")
}

func (t *paginationTaskImpl) GetGenerator() (runner.Generator, error) {
	return &generators.Markdown{
		Title:          t.title,
		Description:    t.description,
		URL:            path.Join("/", t.baseDestination, t.slug) + "/",
		Slug:           t.slug,
		Sources:        t.sources,
		IsPost:         true,
		HighlightStyle: t.highlightStyle,
		Template:       t.template,
		TemplateCtx:    t.templateCtx,
		Pagination:     t.pagination,
		LayoutCtx:      t.layoutCtx,

		OpenGraphTitle:         t.openGraphTitle,
		OpenGraphDescription:   t.openGraphDescription,
		OpenGraphImage:         t.openGraphImage,
		OpenGraphImageURL:      t.openGraphImageURL,
		OpenGraphImageGenerate: t.openGraphImageGenerate,
		OpenGraphImageGenColor: t.openGraphImageGenColor,
		OpenGraphImageGenDPI:   t.openGraphImageGenDPI,
		OpenGraphImageGenSize:  t.openGraphImageGenSize,
	}, nil
}

type Pagination struct {
	Atom            bool
	Title           string
	Description     string
	Sources         []*generators.MarkdownSource
	PostsPerPage    int
	SortReverse     bool
	HighlightStyle  string
	BaseDestination string
	Template        string
	TemplateCtx     map[string]interface{}
	WithSidebar     bool

	OpenGraphTitle         string
	OpenGraphDescription   string
	OpenGraphImage         string
	OpenGraphImageGenColor *uint32
	OpenGraphImageGenDPI   *float64
	OpenGraphImageGenSize  *float64
}

func (p *Pagination) GetBaseDestination() string {
	return p.BaseDestination
}

func (p *Pagination) GetTasks() ([]*runner.Task, error) {
	if p.PostsPerPage == 0 {
		return nil, nil
	}

	tmpl := p.Template
	if tmpl == "" {
		tmpl = "pagination.html"
		if p.Atom {
			tmpl = "atom.xml"
		}
	}

	style := p.HighlightStyle
	if style == "" {
		style = "github"
	}

	posts := []*paginationPost{}
	for _, src := range p.Sources {
		post := &paginationPost{
			Source: src,
		}

		m, err := generators.MarkdownGetMetadata(post.Source.File)
		if err != nil {
			return nil, err
		}

		post.Date = m.Date.Time
		posts = append(posts, post)
	}

	slices.SortStableFunc(posts, func(a *paginationPost, b *paginationPost) int {
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
			runner.NewTask(p,
				&paginationTaskImpl{
					atom:            p.Atom,
					baseDestination: p.BaseDestination,
					title:           p.Title,
					description:     p.Description,
					sources:         nil,
					slug:            "",
					highlightStyle:  style,
					template:        tmpl,
					templateCtx:     p.TemplateCtx,
					pagination: &templates.ContentPagination{
						Enabled: p.PostsPerPage > 0,
						AtomURL: path.Join("/", p.BaseDestination, "atom.xml"),
					},
					layoutCtx: layoutCtx,

					openGraphImageGenerate: !p.Atom,
					openGraphTitle:         p.OpenGraphTitle,
					openGraphDescription:   p.OpenGraphDescription,
					openGraphImage:         p.OpenGraphImage,
					openGraphImageGenColor: p.OpenGraphImageGenColor,
					openGraphImageGenDPI:   p.OpenGraphImageGenDPI,
					openGraphImageGenSize:  p.OpenGraphImageGenSize,
				},
			),
		}, nil
	}

	rv := []*runner.Task{}
	for chk := range slices.Chunk(posts, ppp) {
		srcs := []*generators.MarkdownSource{}
		for _, s := range chk {
			srcs = append(srcs, s.Source)
		}

		pagination := &templates.ContentPagination{
			Enabled: p.PostsPerPage > 0,
			BaseURL: path.Join("/", p.BaseDestination, "page"),
			AtomURL: path.Join("/", p.BaseDestination, "atom.xml"),
			Current: page,
			Total:   total,
		}
		if page > 1 {
			pagination.LinkPrevious = path.Join(pagination.BaseURL, strconv.FormatInt(int64(page-1), 10)) + "/"
		}
		if page < total {
			pagination.LinkNext = path.Join(pagination.BaseURL, strconv.FormatInt(int64(page+1), 10)) + "/"
		}

		if page == 1 {
			rv = append(rv,
				runner.NewTask(p,
					&paginationTaskImpl{
						atom:            p.Atom,
						baseDestination: p.BaseDestination,
						title:           p.Title,
						description:     p.Description,
						sources:         srcs,
						slug:            "",
						highlightStyle:  style,
						template:        tmpl,
						templateCtx:     p.TemplateCtx,
						pagination:      pagination,
						layoutCtx:       layoutCtx,

						openGraphImageGenerate: !p.Atom,
						openGraphTitle:         p.OpenGraphTitle,
						openGraphDescription:   p.OpenGraphDescription,
						openGraphImage:         p.OpenGraphImage,
						openGraphImageGenColor: p.OpenGraphImageGenColor,
						openGraphImageGenDPI:   p.OpenGraphImageGenDPI,
						openGraphImageGenSize:  p.OpenGraphImageGenSize,
					},
				),
			)
			if p.Atom {
				break
			}
		}

		iurl := path.Join("/", p.BaseDestination)
		if iurl != "/" {
			iurl += "/"
		}
		rv = append(rv,
			runner.NewTask(p,
				&paginationTaskImpl{
					atom:            p.Atom,
					baseDestination: p.BaseDestination,
					title:           p.Title,
					description:     p.Description,
					sources:         srcs,
					slug:            path.Join("page", strconv.FormatInt(int64(page), 10)),
					highlightStyle:  style,
					template:        tmpl,
					templateCtx:     p.TemplateCtx,
					pagination:      pagination,
					layoutCtx:       layoutCtx,

					openGraphImageGenerate: false,
					openGraphImageURL:      ogimage.URL(iurl),
					openGraphTitle:         p.OpenGraphTitle,
					openGraphDescription:   p.OpenGraphDescription,
				},
			),
		)
		page++
	}

	return rv, nil
}
