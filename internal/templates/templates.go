package templates

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"

	"rafaelmartins.com/p/website/internal/cdocs"
	"rafaelmartins.com/p/website/internal/config"
	"rafaelmartins.com/p/website/internal/meta"
	"rafaelmartins.com/p/website/internal/utils"
)

var (
	//go:embed embed/*
	embedded embed.FS

	content fs.FS = embed.FS{}

	ccfg       *config.Config
	cassetsDir string
	debug      bool
)

type LayoutContext struct {
	WithSidebar bool
}

type AtomContentEntry struct {
	Updated time.Time
}

type PostContentEntry struct {
	Author struct {
		Name  string
		Email string
	}
	Published time.Time
	Updated   time.Time
}

type ProjectContentLatestReleaseFile struct {
	File string
	URL  string
}

type ProjectContentLatestRelease struct {
	Name  string
	Tag   string
	Body  string
	URL   string
	Files []*ProjectContentLatestReleaseFile
}

type ProjectContentDocumentation struct {
	URL   string
	Label string
}

type ProjectContentMenu struct {
	Active bool
	URL    string
	Title  string
}

type ProjectContentLicense struct {
	SpdxId string
	Title  string
}

type ProjectContentEntry struct {
	Owner         string
	Repo          string
	URL           string
	Description   string
	Menus         []*ProjectContentMenu
	Licenses      []*ProjectContentLicense
	GoImport      string
	GoRepo        string
	CDocsEnabled  bool
	CDocsURL      string
	Stars         int
	Watching      int
	Forks         int
	LatestRelease *ProjectContentLatestRelease
	IsRoot        bool
}

type OpenGraphEntry struct {
	Title       string
	Description string
	Image       string
}

type ContentEntry struct {
	File    string
	URL     string
	Title   string
	Body    string
	Post    *PostContentEntry
	Project *ProjectContentEntry
	CDocs   *cdocs.TemplateCtx
	Extra   map[string]any
}

type ContentPagination struct {
	Enabled      bool
	BaseURL      string
	AtomURL      string
	Current      int
	Total        int
	LinkPrevious string
	LinkNext     string
}

type ContentContext struct {
	Title       string
	Description string
	URL         string
	Slug        string
	License     string
	Toc         string
	Search      bool
	OpenGraph   OpenGraphEntry
	Entry       *ContentEntry
	Entries     []*ContentEntry
	Atom        *AtomContentEntry
	Pagination  *ContentPagination
	Extra       map[string]any
}

var gen *meta.Metadata

type context struct {
	Config    *config.Config
	Generator *meta.Metadata
	Layout    *LayoutContext
	Content   *ContentContext
	Extra     map[string]any
	Time      time.Time
	Debug     bool
}

func SetConfig(cfg *config.Config) {
	ccfg = cfg
	if ccfg.Template == nil {
		var err error
		content, err = fs.Sub(embedded, "embed")
		if err != nil {
			panic(err)
		}
	}
}

func SetAssetsDir(assetsDir string) {
	cassetsDir = assetsDir
}

func SetDebug(d bool) {
	debug = d
}

func GetPaths(name string) ([]string, error) {
	rv := []string{}
	if ccfg != nil && ccfg.Template != nil {
		rv = append(rv, *ccfg.Template)
	} else {
		rv = append(rv, utils.Executable())
	}

	if _, err := os.Stat(name); err == nil {
		rv = append(rv, name)
	} else if _, err := content.Open(name); err == nil {
		// do nothing, executable timestamp already included
	} else {
		return nil, fmt.Errorf("templates: failed to find template: %s", name)
	}

	if ccfg != nil {
		rv = append(rv, ccfg.TemplatePartials...)
	}

	return rv, nil
}

func assetsUrl() string {
	return "/" + cassetsDir
}

func required(v reflect.Value) (reflect.Value, error) {
	if !v.IsValid() {
		return reflect.Value{}, errors.New("invalid value")
	}
	if v.IsZero() {
		return reflect.Value{}, errors.New("zero value")
	}
	return v, nil
}

func requiredAttr(v reflect.Value) (reflect.Value, error) {
	r, err := required(v)
	if err != nil {
		return reflect.Value{}, err
	}
	if r.Kind() == reflect.Interface {
		r = reflect.ValueOf(r.Interface())
	}
	if strings.ContainsAny(r.String(), "\t\n\r\"<>") {
		return reflect.Value{}, errors.New("value should not contain tabs, new lines, double quotes, html tags")
	}
	return v, nil
}

func volatile(v any, ondebug any) any {
	if debug {
		return ondebug
	}
	return v
}

func Execute(wr io.Writer, name string, fm template.FuncMap, lctx *LayoutContext, cctx *ContentContext) error {
	if fm == nil {
		fm = template.FuncMap{}
	}
	fm["assetsUrl"] = assetsUrl
	fm["required"] = required
	fm["requiredAttr"] = requiredAttr
	fm["volatile"] = volatile

	var tmpl *template.Template
	if ccfg != nil && ccfg.Template != nil {
		var err error
		tmpl, err = template.New("base").Funcs(fm).ParseFiles(*ccfg.Template)
		if err != nil {
			return err
		}
	}
	if tmpl == nil {
		tmpl = template.New("base").Funcs(fm)

		// autoload base unless it is the main template
		if name != "base.html" {
			t, err := tmpl.ParseFS(content, "base.html")
			if err != nil {
				return err
			}
			tmpl = t
		}
	}

	if _, err := os.Stat(name); err == nil {
		tmpl, err = tmpl.ParseFiles(name)
		if err != nil {
			return err
		}
	} else if _, err := content.Open(name); err == nil {
		tmpl, err = tmpl.ParseFS(content, name)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("templates: failed to find template: %s", name)
	}

	if ccfg != nil && len(ccfg.TemplatePartials) > 0 {
		var err error
		tmpl, err = tmpl.ParseFiles(ccfg.TemplatePartials...)
		if err != nil {
			return err
		}
	}

	llctx := lctx
	if llctx == nil {
		llctx = &LayoutContext{}
	}

	lcctx := cctx
	if lcctx == nil {
		lcctx = &ContentContext{}
	}

	if gen == nil {
		m, err := meta.GetMetadata()
		if err != nil {
			return err
		}
		gen = m
	}

	return tmpl.Option("missingkey=zero").ExecuteTemplate(wr, "base", &context{
		Config:    ccfg,
		Generator: gen,
		Layout:    llctx,
		Content:   lcctx,
		Extra:     ccfg.TemplateCtx,
		Time:      time.Now().UTC(),
		Debug:     debug,
	})
}
