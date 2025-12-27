package project

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"time"

	"rafaelmartins.com/p/website/internal/cdocs"
	"rafaelmartins.com/p/website/internal/ogimage"
	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/templates"
)

type cDocs struct {
	proj   *Project
	otitle string
}

func (c *cDocs) GetDestination() string {
	return filepath.Join(c.proj.Repo, c.proj.cdocsDestination, "index.html")
}

func (c *cDocs) GetGenerator() (runner.Generator, error) {
	return c, nil
}

func (*cDocs) GetID() string {
	return "CDOCS"
}

func (c *cDocs) GetReader() (io.ReadCloser, error) {
	baseHtmlUrl := "https://github.com/" + c.proj.Owner + "/" + c.proj.Repo + "/blob/" + c.proj.proj.Head
	headerPath := ""
	if c.proj.CDocsBaseDirectory != nil {
		baseHtmlUrl += "/" + *c.proj.CDocsBaseDirectory
		headerPath = *c.proj.CDocsBaseDirectory
	}

	headers := []*cdocs.TemplateCtxHeader{}
	for _, h := range c.proj.CDocsHeaders {
		headerPath := path.Join(headerPath, h)
		hdr := []byte{}
		found := false
		for _, header := range c.proj.proj.Headers {
			if header.Name == headerPath {
				var err error
				hdr, err = header.Read()
				if err != nil {
					return nil, err
				}
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("cdocs: header not found: %s", h)
		}

		ast, err := cdocs.Parse(h, io.NopCloser(bytes.NewBuffer(hdr)))
		if err != nil {
			return nil, err
		}

		headers = append(headers, &cdocs.TemplateCtxHeader{
			Filename:  h,
			Header:    ast,
			GithubUrl: baseHtmlUrl + "/" + h,
		})
	}

	dctx, err := cdocs.NewTemplateCtx(headers)
	if err != nil {
		return nil, err
	}

	title := fmt.Sprintf("API Documentation: %s", c.proj.Repo)

	c.otitle = title
	if c.proj.CDocsOpenGraphTitle != "" {
		c.otitle = c.proj.CDocsOpenGraphTitle
	}

	tmpl := c.proj.CDocsTemplate
	if tmpl == "" {
		tmpl = "cdocs.html"
	}

	lctx := &templates.LayoutContext{
		WithSidebar: c.proj.CDocsWithSidebar,
	}

	buf := &bytes.Buffer{}
	if err := templates.Execute(buf, tmpl, nil, lctx, &templates.ContentContext{
		Title: title,
		URL:   c.proj.cdocsUrl,
		OpenGraph: templates.OpenGraphEntry{
			Title:       c.otitle,
			Description: c.proj.CDocsOpenGraphDescription,
			Image:       ogimage.URL(c.proj.cdocsUrl),
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

func (c *cDocs) GetTimeStamps() ([]time.Time, error) {
	if c.proj.Immutable {
		return nil, nil
	}

	rv, err := templates.GetTimestamps(c.proj.CDocsTemplate, !c.proj.Immutable)
	if err != nil {
		return nil, err
	}

	og, err := ogimage.GetTimeStamps()
	if err != nil {
		return nil, err
	}
	rv = append(rv, og...)

	return rv, nil
}

func (c *cDocs) GetImmutable() bool {
	return c.proj.Immutable
}

func (c *cDocs) GetByProducts(ch chan *runner.GeneratorByProduct) {
	ogimage.GenerateByProduct(ch, c.otitle, true, c.proj.CDocsOpenGraphImage, c.proj.CDocsOpenGraphImageGenColor, c.proj.CDocsOpenGraphImageGenDPI, c.proj.CDocsOpenGraphImageGenSize)
	close(ch)
}
