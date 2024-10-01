package main

import (
	"flag"
	"log"

	"github.com/rafaelmartins/website/internal/assets"
	"github.com/rafaelmartins/website/internal/config"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/tasks"
	"github.com/rafaelmartins/website/internal/templates"
	"github.com/rafaelmartins/website/internal/webserver"
)

var (
	fBuildDir   = flag.String("d", "_build", "build directory")
	fConfigFile = flag.String("f", "config.yml", "configuration file")
	fListenAddr = flag.String("a", ":3000", "development web server listen address")
	fRunServer  = flag.Bool("r", false, "run development server")

	cfg        *config.Config      = nil
	taskGroups []*runner.TaskGroup = nil
)

func getTaskGroups(c *config.Config) ([]*runner.TaskGroup, error) {
	assetsDir := c.Assets.BaseDestination
	if assetsDir == "" {
		assetsDir = "assets"
	}

	rv := []*runner.TaskGroup{
		// assets required by embedded templates
		runner.NewTaskGroup(
			&assets.CdnjsLibrary{
				Name:    "anchor-js",
				Version: "5.0.0",
				Files: []string{
					"anchor.min.js",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&assets.CdnjsLibrary{
				Name:    "bulma",
				Version: "1.0.2",
				Files: []string{
					"css/bulma.min.css",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&assets.CdnjsLibrary{
				Name:    "font-awesome",
				Version: "6.6.0",
				Files: []string{
					"css/all.min.css",
					"webfonts/fa-brands-400.ttf",
					"webfonts/fa-brands-400.woff2",
					"webfonts/fa-regular-400.ttf",
					"webfonts/fa-regular-400.woff2",
					"webfonts/fa-solid-900.ttf",
					"webfonts/fa-solid-900.woff2",
					"webfonts/fa-v4compatibility.ttf",
					"webfonts/fa-v4compatibility.woff2",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&assets.CdnjsLibrary{
				Name:    "github-markdown-css",
				Version: "5.6.1",
				Files: []string{
					"github-markdown.min.css",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&assets.NunitoFont{
				BaseDestination: assetsDir,
			},
		),
	}

	for _, js := range c.Assets.Cdnjs {
		rv = append(rv,
			runner.NewTaskGroup(
				&assets.CdnjsLibrary{
					Name:            js.Name,
					Version:         js.Version,
					Files:           js.Files,
					BaseDestination: assetsDir,
				},
			),
		)
	}

	for _, f := range c.Files {
		rv = append(rv,
			runner.NewTaskGroup(
				&tasks.Files{
					Paths:           f.Paths,
					BaseDestination: f.BaseDestination,
				},
			),
		)
	}

	for _, pg := range c.Pages {
		src := map[string]string{}
		for _, s := range pg.Sources {
			src[s.Slug] = s.File
		}

		rv = append(rv,
			runner.NewTaskGroup(
				&tasks.Pages{
					Sources:           src,
					ExtraDependencies: pg.ExtraDependencies,
					HighlightStyle:    pg.HighlightStyle,
					BaseDestination:   pg.BaseDestination,
					Template:          pg.Template,
					TemplateCtx:       pg.TemplateCtx,
					WithSidebar:       pg.WithSidebar,
				},
			),
		)
	}

	for _, ps := range c.Posts {
		sortReverse := true
		if ps.SortReverse != nil && !*ps.SortReverse {
			sortReverse = false
		}

		ppp := 10
		if ps.PostsPerPage != nil {
			ppp = *ps.PostsPerPage
		}

		pppa := 10
		if ps.PostsPerPageAtom != nil {
			pppa = *ps.PostsPerPageAtom
		}

		rv = append(rv,
			runner.NewTaskGroup(
				&tasks.Posts{
					SourceDir:       ps.SourceDir,
					HighlightStyle:  ps.HighlightStyle,
					BaseDestination: ps.BaseDestination,
					Template:        ps.Template,
					TemplateCtx:     ps.TemplateCtx,
					WithSidebar:     ps.WithSidebar,
				},
			),
			runner.NewTaskGroup(
				&tasks.PostsPagination{
					Title:           ps.Title,
					Description:     ps.Description,
					SourceDir:       ps.SourceDir,
					SeriesStatus:    ps.SeriesStatus,
					PostsPerPage:    ppp,
					SortReverse:     sortReverse,
					HighlightStyle:  ps.HighlightStyle,
					BaseDestination: ps.BaseDestination,
					Template:        ps.TemplatePagination,
					WithSidebar:     ps.WithSidebar,
				},
			),
			runner.NewTaskGroup(
				&tasks.Atom{
					Title:           ps.Title,
					SourceDir:       ps.SourceDir,
					PostsPerPage:    pppa,
					HighlightStyle:  ps.HighlightStyle,
					BaseDestination: ps.BaseDestination,
					Template:        ps.TemplateAtom,
				},
			),
		)
	}

	for _, pj := range c.Projects {
		// immutable by default, only disable manually for development
		immutable := true
		if pj.Immutable != nil && !*pj.Immutable {
			immutable = false
		}

		rv = append(rv,
			runner.NewTaskGroup(
				&tasks.Project{
					Owner:           pj.Owner,
					Repo:            pj.Repo,
					BaseDestination: pj.BaseDestination,
					Template:        pj.Template,
					Immutable:       immutable,
					WithSidebar:     pj.WithSidebar,
				},
			),
		)
	}
	return rv, nil
}

func build() error {
	if cfg == nil || !cfg.IsUpToDate() {
		var err error
		cfg, err = config.New(*fConfigFile)
		if err != nil {
			return err
		}
		templates.SetConfig(cfg)

		tg, err := getTaskGroups(cfg)
		if err != nil {
			return err
		}
		taskGroups = tg
	}
	return runner.Run(*fBuildDir, cfg, taskGroups)
}

func main() {
	flag.Parse()

	if *fRunServer {
		if err := webserver.ListenAndServeWithReloader(*fListenAddr, *fBuildDir, build); err != nil {
			log.Fatalf("error: %s", err)
		}
	} else {
		if err := build(); err != nil {
			log.Fatalf("error: %s", err)
		}
	}
}
