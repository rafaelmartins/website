package project

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"path/filepath"

	"rafaelmartins.com/p/website/internal/cdocs"
	"rafaelmartins.com/p/website/internal/github"
	"rafaelmartins.com/p/website/internal/opengraph"
	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/templates"
)

type cDocs struct {
	proj *Project
	og   *opengraph.OpenGraph
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

func (c *cDocs) getTemplate() string {
	rv := c.proj.CDocsTemplate
	if rv == "" {
		rv = "cdocs.html"
	}
	return rv
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
		var header *github.RepositoryFile
		for _, hh := range c.proj.proj.Headers {
			if hh.Name == path.Join(headerPath, h) {
				header = hh
				break
			}
		}
		if header == nil {
			return nil, fmt.Errorf("cdocs: header not found: %s", h)
		}

		data, err := header.Read()
		if err != nil {
			return nil, err
		}

		ast, err := cdocs.Parse(h, io.NopCloser(bytes.NewBuffer(data)))
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

	title := fmt.Sprintf("%s: API Documentation", c.proj.Repo)

	og, err := opengraph.New(c.proj.OpenGraphImageGen, false, c.proj.cdocsUrl, title, "", c.proj.CDocsOpenGraph, "", "", nil)
	if err != nil {
		return nil, err
	}
	c.og = og

	buf := &bytes.Buffer{}
	if err := templates.Execute(buf, c.getTemplate(), nil, nil, &templates.ContentContext{
		Title:     title,
		URL:       c.proj.cdocsUrl,
		License:   c.proj.license,
		Search:    true, // FIXME ???
		OpenGraph: og.GetTemplateContext(),
		Entry: &templates.ContentEntry{
			Title: title,
			CDocs: dctx,
		},
	}); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (c *cDocs) GetPaths() ([]string, error) {
	if c.proj.Immutable && c.proj.LocalDirectory == nil {
		return nil, nil
	}

	rv, err := templates.GetPaths(c.getTemplate())
	if err != nil {
		return nil, err
	}

	if c.proj.LocalDirectory != nil {
		for _, header := range c.proj.proj.Headers {
			rv = append(rv, filepath.Join(*c.proj.LocalDirectory, header.Name))
		}
	}

	if c.proj.OpenGraphImageGen != nil {
		rv = append(rv, c.proj.OpenGraphImageGen.GetPaths()...)
	}
	return rv, nil
}

func (c *cDocs) GetImmutable() bool {
	return c.proj.Immutable && c.proj.LocalDirectory == nil
}

func (c *cDocs) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		c.og.GenerateByProduct(ch, "")
		close(ch)
	}
}
