package generators

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/rafaelmartins/website/internal/cdocs"
	"github.com/rafaelmartins/website/internal/github"
	"github.com/rafaelmartins/website/internal/ogimage"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
)

type CDocs struct {
	Owner         string
	Repo          string
	Headers       []string
	BaseDirectory string
	URL           string
	Template      string
	LayoutCtx     *templates.LayoutContext
	Immutable     bool

	OpenGraphTitle         string
	OpenGraphDescription   string
	OpenGraphImage         string
	OpenGraphImageGenColor *uint32
	OpenGraphImageGenDPI   *float64
	OpenGraphImageGenSize  *float64

	headerCtx map[string]*github.RequestContext
	otitle    string
}

func (*CDocs) GetID() string {
	return "CDOCS"
}

func (d *CDocs) initHeaderCtx() {
	if d.headerCtx == nil {
		d.headerCtx = map[string]*github.RequestContext{}
		for _, h := range d.Headers {
			if _, found := d.headerCtx[h]; !found {
				d.headerCtx[h] = &github.RequestContext{}
			}
		}
	}
}

func (d *CDocs) GetReader() (io.ReadCloser, error) {
	d.initHeaderCtx()

	headers := []*cdocs.TemplateCtxHeader{}
	for _, h := range d.Headers {
		hdr, htmlUrl, err := github.Contents(d.headerCtx[h], d.Owner, d.Repo, path.Join(d.BaseDirectory, h), true)
		if err != nil {
			return nil, err
		}

		ast, err := cdocs.Parse(h, hdr)
		if err != nil {
			return nil, err
		}

		headers = append(headers, &cdocs.TemplateCtxHeader{
			Filename:  h,
			Header:    ast,
			GithubUrl: htmlUrl,
		})
	}

	dctx, err := cdocs.NewTemplateCtx(headers)
	if err != nil {
		return nil, err
	}

	title := fmt.Sprintf("API Documentation: %s", d.Repo)

	d.otitle = title
	if d.OpenGraphTitle != "" {
		d.otitle = d.OpenGraphTitle
	}

	buf := &bytes.Buffer{}
	if err := templates.Execute(buf, d.Template, nil, d.LayoutCtx, &templates.ContentContext{
		Title: title,
		URL:   d.URL,
		OpenGraph: templates.OpenGraphEntry{
			Title:       d.otitle,
			Description: d.OpenGraphDescription,
			Image:       ogimage.URL(d.URL),
		},
		Entry: &templates.ContentEntry{
			Title: title,
			CDocs: dctx,
		},
	}); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (d *CDocs) GetTimeStamps() ([]time.Time, error) {
	// we would be safe to just run this method frequently, as we support cache with
	// etag/last-modified, but it is easier to just disable this manually when adding
	// a new project than spam github servers for no good reason.
	if d.Immutable {
		return nil, nil
	}

	rv, err := templates.GetTimestamps(d.Template, !d.Immutable)
	if err != nil {
		return nil, err
	}

	og, err := ogimage.GetTimeStamps()
	if err != nil {
		return nil, err
	}
	rv = append(rv, og...)

	d.initHeaderCtx()

	for _, h := range d.Headers {
		if _, _, err := github.Contents(d.headerCtx[h], d.Owner, d.Repo, path.Join(d.BaseDirectory, h), false); err != nil {
			return nil, err
		}
		rv = append(rv, d.headerCtx[h].LastModifiedTime)
	}

	return rv, nil
}

func (d *CDocs) GetImmutable() bool {
	return d.Immutable
}

func (d *CDocs) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch == nil {
		return
	}

	ogimage.GenerateByProduct(ch, d.otitle, true, d.OpenGraphImage, d.OpenGraphImageGenColor, d.OpenGraphImageGenDPI, d.OpenGraphImageGenSize)
	close(ch)
}
