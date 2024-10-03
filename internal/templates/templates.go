package templates

import (
	"embed"
	"fmt"
	"io"
	"os"
	"text/template"
	"time"

	"github.com/rafaelmartins/website/internal/config"
	"github.com/rafaelmartins/website/internal/utils"
)

var (
	//go:embed embed/*
	content embed.FS

	ccfg *config.Config
)

type LayoutContext struct {
	WithSidebar bool
}

type AtomContentEntry struct {
	Updated time.Time
}

type SeriesContentEntry struct {
	Status string
}

type PostContentEntry struct {
	Author struct {
		Name  string
		Email string
	}
	Date time.Time
	// Tags []string
}

type ContentEntry struct {
	File  string
	URL   string
	Title string
	Body  string
	Post  *PostContentEntry
	Extra map[string]interface{}
}

type ContentPagination struct {
	Enabled   bool
	BaseURL   string
	AtomURL   string
	Current   int
	Total     int
	LinkFirst string
	LinkLast  string
}

type ContentContext struct {
	Title       string
	Description string
	URL         string
	Entry       *ContentEntry
	Entries     []*ContentEntry
	Atom        *AtomContentEntry
	Series      *SeriesContentEntry
	Pagination  *ContentPagination
	Extra       map[string]interface{}
}

type context struct {
	Config  *config.Config
	Layout  *LayoutContext
	Content *ContentContext
}

func SetConfig(cfg *config.Config) {
	ccfg = cfg
}

func GetTimestamps(name string) ([]time.Time, error) {
	// we always load the base.hml template, even if it is overwritten completely later
	// then we must always include the executable timestamp, as this template is embedded.
	ts, err := utils.ExecutableTimestamp()
	if err != nil {
		return nil, err
	}

	rv := []time.Time{ts}
	if st, err := os.Stat(name); err == nil {
		rv = append(rv, st.ModTime().UTC())
	} else if _, err := content.Open("embed/" + name); err == nil {
		// do nothing, executable timestamp already included
	} else {
		return nil, fmt.Errorf("templates: failed to find template: %s", name)
	}
	return rv, nil
}

func Execute(wr io.Writer, name string, fm template.FuncMap, lctx *LayoutContext, cctx *ContentContext) error {
	tmpl, err := template.New("base").Funcs(fm).ParseFS(content, "embed/base.html")
	if err != nil {
		return err
	}

	if _, err := os.Stat(name); err == nil {
		tmpl, err = tmpl.ParseFiles(name)
		if err != nil {
			return err
		}
	} else if _, err := content.Open("embed/" + name); err == nil {
		tmpl, err = tmpl.ParseFS(content, "embed/"+name)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("templates: failed to find template: %s", name)
	}

	llctx := lctx
	if llctx == nil {
		llctx = &LayoutContext{}
	}

	lcctx := cctx
	if lcctx == nil {
		lcctx = &ContentContext{}
	}

	return tmpl.ExecuteTemplate(wr, "base", &context{
		Config:  ccfg,
		Layout:  llctx,
		Content: lcctx,
	})
}
