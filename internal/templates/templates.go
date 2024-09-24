package templates

import (
	"embed"
	"html/template"
	"io"
	"log"
	"os"
	"time"

	"github.com/rafaelmartins/website/internal/config"
	"github.com/rafaelmartins/website/internal/markdown"
	"github.com/rafaelmartins/website/internal/utils"
)

var (
	//go:embed html/*.html
	content embed.FS

	ccfg *config.Config

	tmplFm = template.FuncMap{
		"markdownMetadataProperty": func(f string, prop string, dflt interface{}) interface{} {
			rv, err := markdown.GetMetadataProperty(f, prop, dflt)
			if err != nil {
				log.Print(err)
				return dflt
			}
			return rv
		},
	}
)

type LayoutContext struct {
	WithSidebar bool
}

type context struct {
	Config  *config.Config
	Layout  *LayoutContext
	Content map[string]interface{}
}

func SetConfig(cfg *config.Config) {
	ccfg = cfg
}

func GetTimestamps(name string) ([]time.Time, error) {
	rv := []time.Time{}
	if st, err := os.Stat(name); err == nil {
		rv = append(rv, st.ModTime().UTC())
	} else if _, err := template.New("base").Funcs(tmplFm).ParseFS(content, "html/"+name); err != nil {
		return nil, err
	}
	ts, err := utils.ExecutableTimestamp()
	if err != nil {
		return nil, err
	}
	return append(rv, ts), nil
}

func Execute(wr io.Writer, name string, lctx *LayoutContext, cctx map[string]interface{}) error {
	tmpl, err := template.New("base").Funcs(tmplFm).ParseFS(content, "html/base.html")
	if err != nil {
		return err
	}

	if _, err := os.Stat(name); err == nil {
		tmpl, err = tmpl.ParseFiles(name)
		if err != nil {
			return err
		}
	} else {
		tmpl, err = tmpl.ParseFS(content, "html/"+name)
		if err != nil {
			return err
		}
	}

	llctx := lctx
	if llctx == nil {
		llctx = &LayoutContext{}
	}

	return tmpl.ExecuteTemplate(wr, "base", &context{
		Config:  ccfg,
		Layout:  llctx,
		Content: cctx,
	})
}
