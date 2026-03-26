package generators

import (
	"bytes"
	"errors"
	"io"
	"path"
	"path/filepath"
	"slices"
	"text/template"
	"time"

	"rafaelmartins.com/p/website/internal/content"
	"rafaelmartins.com/p/website/internal/opengraph"
	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/templates"
)

type ContentSource struct {
	File string
	URL  string
}

type Content struct {
	Title             string
	Description       string
	URL               string
	Slug              string
	License           string
	Toc               bool
	Search            *bool
	Sources           []*ContentSource
	IsPost            bool
	ExtraDependencies []string
	Template          string
	TemplateCtx       map[string]any
	Pagination        *templates.ContentPagination
	LayoutCtx         *templates.LayoutContext

	OpenGraph                    *opengraph.Config
	OpenGraphImageGen            *opengraph.OpenGraphImageGen
	OpenGraphPregeneratedBaseUrl string

	ctx *templates.ContentContext
	og  *opengraph.OpenGraph
}

func (*Content) GetID() string {
	return "CONTENT"
}

func (h *Content) GetReader() (io.ReadCloser, error) {
	if h.URL == "" {
		return nil, errors.New("content: missing url")
	}

	ctx := &templates.ContentContext{
		Title:       h.Title,
		Description: h.Description,
		URL:         h.URL,
		Slug:        h.Slug,
		License:     h.License,
		Search:      true,
		Atom:        &templates.AtomContentEntry{},
		Pagination:  h.Pagination,
		Extra:       h.TemplateCtx,
	}
	if h.Search != nil {
		ctx.Search = *h.Search
	}
	h.ctx = ctx

	atomUpdated := time.Time{}
	entries := []*templates.ContentEntry{}
	mt := ""
	md := ""
	var mog *opengraph.Config

	for _, src := range h.Sources {
		if src.File == "" {
			continue
		}

		var withToc *bool
		if h.Pagination == nil {
			withToc = &h.Toc
		}

		metadata, toc, body, err := content.Render(src.File, h.URL, withToc)
		if err != nil {
			return nil, err
		}

		entry := &templates.ContentEntry{
			File:  src.File,
			URL:   src.URL,
			Title: metadata.Title,
			Body:  body,
		}

		if h.IsPost {
			entry.Post = &templates.PostContentEntry{
				Published: metadata.Published.Time,
				Updated:   metadata.Updated.Time,
			}
			entry.Post.Author.Name = metadata.Author.Name
			entry.Post.Author.Email = metadata.Author.Email
			if atomUpdated.Before(entry.Post.Published) {
				atomUpdated = entry.Post.Published
			}
			if atomUpdated.Before(entry.Post.Updated) {
				atomUpdated = entry.Post.Updated
			}
		}

		entry.Extra = metadata.Extra

		if h.Pagination == nil {
			if ctx.Title == "" {
				ctx.Title = entry.Title
			}
			ctx.Entry = entry
			ctx.Toc = toc
			if metadata.License != "" {
				ctx.License = metadata.License
			}
			if metadata.Search != nil {
				ctx.Search = *metadata.Search
			}

			mt = metadata.Title
			md = metadata.Description
			mog = metadata.OpenGraph
			break
		}

		entries = append(entries, entry)
	}

	if h.Pagination != nil {
		ctx.Entries = entries
		ctx.Atom.Updated = atomUpdated
		if ctx.Atom.Updated.IsZero() {
			ctx.Atom.Updated = time.Unix(0, 0)
		}
	}

	baseurl := h.URL
	ogpregenerated := false
	if h.OpenGraphPregeneratedBaseUrl != "" {
		baseurl = h.OpenGraphPregeneratedBaseUrl
		ogpregenerated = true
	}
	og, err := opengraph.New(h.OpenGraphImageGen, ogpregenerated, baseurl, ctx.Title, ctx.Description, h.OpenGraph, mt, md, mog)
	if err != nil {
		return nil, err
	}
	h.og = og

	ctx.OpenGraph = h.og.GetTemplateContext()

	funcMap := template.FuncMap{
		"contentGetMetadata": content.GetMetadata,
	}

	buf := &bytes.Buffer{}
	if err := templates.Execute(buf, h.Template, funcMap, h.LayoutCtx, ctx); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (h *Content) GetPaths() ([]string, error) {
	rv, err := templates.GetPaths(h.Template)
	if err != nil {
		return nil, err
	}

	if h.OpenGraphImageGen != nil {
		rv = append(rv, h.OpenGraphImageGen.GetPaths()...)
	}

	for _, src := range h.Sources {
		if src.File == "" {
			continue
		}

		if h.Pagination != nil {
			if fd := filepath.Dir(src.File); !slices.Contains(rv, fd) {
				rv = append(rv, fd)
			}
		}

		rv = append(rv, src.File)
	}

	return append(rv, h.ExtraDependencies...), nil
}

func (*Content) GetImmutable() bool {
	return false
}

func (h *Content) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch == nil {
		return
	}

	if h.ctx.Entry != nil {
		assets, err := content.ListAssets(h.ctx.Entry.File)
		if err != nil {
			ch <- &runner.GeneratorByProduct{Err: err}
			return
		}
		for _, asset := range assets {
			fn, fp, err := content.OpenAsset(h.ctx.Entry.File, asset)
			if err != nil {
				ch <- &runner.GeneratorByProduct{Err: err}
				return
			}
			ch <- &runner.GeneratorByProduct{
				Filename: fn,
				Reader:   fp,
			}
		}
	}

	if h.og != nil {
		basedir := ""
		if len(h.Sources) > 0 {
			basedir = path.Dir(h.Sources[0].File)
		}
		h.og.GenerateByProduct(ch, basedir)
	}

	close(ch)
}
