package generators

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func mdGetGoldmark(style string) goldmark.Markdown {
	opt := []highlighting.Option{}
	if style != "" {
		opt = append(opt, highlighting.WithStyle(style))
	}

	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			emoji.Emoji,
			meta.Meta,
			highlighting.NewHighlighting(opt...),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
}

func mdGetMetadata(f string, prop string, dflt interface{}) (interface{}, error) {
	src, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}

	context := parser.NewContext()
	mdGetGoldmark("").Parser().Parse(text.NewReader(src), parser.WithContext(context))

	if m := meta.Get(context); m != nil {
		if v, ok := m[prop]; ok {
			return v, nil
		}
	}
	return dflt, nil
}

func mdParseDateFromInterface(itf interface{}) (time.Time, error) {
	date, ok := itf.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("html: invalid date: %+v", itf)
	}

	dt, err := time.Parse(time.DateTime, date)
	if err != nil {
		dt, err = time.Parse(time.DateOnly, date)
		if err != nil {
			return time.Time{}, err
		}
	}
	return dt, nil
}

func MarkdownParseDate(f string) (time.Time, error) {
	itf, err := mdGetMetadata(f, "date", "")
	if err != nil {
		return time.Time{}, err
	}

	return mdParseDateFromInterface(itf)
}

type MarkdownSource struct {
	File string
	Slug string
}

type Markdown struct {
	Title             string
	Sources           []*MarkdownSource
	IsPost            bool
	ExtraDependencies []string
	HighlightStyle    string
	Template          string
	TemplateCtx       map[string]interface{}
	Pagination        *templates.ContentPagination
	LayoutCtx         *templates.LayoutContext
}

func (*Markdown) GetID() string {
	return "HTML"
}

func (h *Markdown) GetReader() (io.ReadCloser, error) {
	ctx := &templates.ContentContext{
		Title:      template.HTML(h.Title),
		Pagination: h.Pagination,
		Extra:      h.TemplateCtx,
	}
	entries := []*templates.ContentEntry{}

	for _, src := range h.Sources {
		if src.File == "" {
			continue
		}

		fc, err := os.ReadFile(src.File)
		if err != nil {
			return nil, err
		}

		buf := &bytes.Buffer{}
		context := parser.NewContext()
		if err := mdGetGoldmark(h.HighlightStyle).Convert(fc, buf, parser.WithContext(context)); err != nil {
			return nil, err
		}
		body := buf.String()
		metadata := meta.Get(context)

		entry := &templates.ContentEntry{
			File: src.File,
			URL:  path.Join("/", src.Slug) + "/",
			Body: template.HTML(body),
		}

		if titleItf, ok := metadata["title"]; ok {
			if title, ok := titleItf.(string); ok {
				entry.Title = template.HTML(title)
				delete(metadata, "title")
			}
		}

		if h.IsPost {
			post := &templates.PostContentEntry{}

			if dateItf, ok := metadata["date"]; ok {
				dt, err := mdParseDateFromInterface(dateItf)
				if err != nil {
					return nil, err
				}
				post.Date = dt
				delete(metadata, "date")
			} else {
				return nil, fmt.Errorf("html: post missing date: %s", src.File)
			}

			if authorItf, ok := metadata["author"]; ok {
				if authorMap, ok := authorItf.(map[interface{}]interface{}); ok {
					if nameItf, ok := authorMap["name"]; ok {
						if name, ok := nameItf.(string); ok {
							post.Author.Name = name
						}
					}
					if emailItf, ok := authorMap["email"]; ok {
						if email, ok := emailItf.(string); ok {
							post.Author.Email = email
						}
					}
				}
				delete(metadata, "author")
			}

			// if tagsItf, ok := metadata["tags"]; ok {
			// 	if tagsSlice, ok := tagsItf.([]interface{}); ok {
			// 		for _, tagItf := range tagsSlice {
			// 			if tag, ok := tagItf.(string); ok {
			// 				post.Tags = append(post.Tags, tag)
			// 			}
			// 		}
			// 		delete(metadata, "tags")
			// 	}
			// }

			entry.Post = post
		}

		entry.Extra = metadata

		if h.Pagination == nil {
			if h.Title == "" {
				ctx.Title = entry.Title
			}
			ctx.Entry = entry
			break
		}

		entries = append(entries, entry)
	}

	if h.Pagination != nil {
		ctx.Entries = entries
	}

	funcMap := template.FuncMap{
		"markdownGetMetadata": func(f string, prop string, dflt interface{}) interface{} {
			rv, err := mdGetMetadata(f, prop, dflt)
			if err != nil {
				log.Print(err)
				return dflt
			}
			return rv
		},
	}

	buf := &bytes.Buffer{}
	if err := templates.Execute(buf, h.Template, funcMap, h.LayoutCtx, ctx); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (h *Markdown) GetTimeStamps() ([]time.Time, error) {
	rv, err := templates.GetTimestamps(h.Template)
	if err != nil {
		return nil, err
	}

	for _, src := range h.Sources {
		if src.File == "" {
			continue
		}

		st, err := os.Stat(src.File)
		if err != nil {
			return nil, err
		}
		rv = append(rv, st.ModTime().UTC())
	}

	for _, dep := range h.ExtraDependencies {
		st, err := os.Stat(dep)
		if err != nil {
			return nil, err
		}
		rv = append(rv, st.ModTime().UTC())
	}

	return rv, nil
}

func (*Markdown) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		close(ch)
	}
}
