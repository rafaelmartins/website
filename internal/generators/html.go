package generators

import (
	"bytes"
	"html/template"
	"io"
	"os"
	"time"

	"github.com/rafaelmartins/website/internal/markdown"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
)

type HTML struct {
	Source            string
	ExtraDependencies []string
	HighlightStyle    string
	Template          string
	TemplateCtx       map[string]interface{}
	LayoutCtx         *templates.LayoutContext
}

func (*HTML) GetID() string {
	return "HTML"
}

func (h *HTML) GetReader() (io.ReadCloser, error) {
	metadata := map[string]interface{}{}

	for k, v := range h.TemplateCtx {
		metadata[k] = v
	}

	if h.Source != "" {
		body, m, err := markdown.ParseFile(h.HighlightStyle, h.Source)
		if err != nil {
			return nil, err
		}

		for k, v := range m {
			metadata[k] = v
		}
		metadata["body"] = template.HTML(body)
	}

	buf := &bytes.Buffer{}
	if err := templates.Execute(buf, h.Template, h.LayoutCtx, metadata); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (h *HTML) GetTimeStamps() ([]time.Time, error) {
	rv, err := templates.GetTimestamps(h.Template)
	if err != nil {
		return nil, err
	}

	if h.Source != "" {
		st, err := os.Stat(h.Source)
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

func (*HTML) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		close(ch)
	}
}
