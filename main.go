package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/rafaelmartins/website/internal/config"
	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/meta"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/tasks"
	"github.com/rafaelmartins/website/internal/templates"
	"github.com/rafaelmartins/website/internal/webserver"
)

var (
	fBuildDir   = flag.String("d", "_build", "build directory")
	fConfigFile = flag.String("c", "config.yml", "configuration file")
	fListenAddr = flag.String("a", ":3000", "development web server listen address")
	fRunServer  = flag.Bool("r", false, "run development server")
	fForce      = flag.Bool("f", false, "force re-running all tasks")
	fVersion    = flag.Bool("v", false, "show version and exit")

	cfg        *config.Config      = nil
	taskGroups []*runner.TaskGroup = nil

	force = false
)

func getTaskGroups(c *config.Config) ([]*runner.TaskGroup, error) {
	assetsDir := c.Assets.BaseDestination
	if assetsDir == "" {
		assetsDir = "assets"
	}

	rv := []*runner.TaskGroup{
		// assets required by embedded templates
		runner.NewTaskGroup(
			&tasks.NpmPackage{
				Name:    "anchor-js",
				Version: "5.0.0",
				Files: []string{
					"anchor.min.js",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&tasks.NpmPackage{
				Name:    "bulma",
				Version: "1.0.2",
				Files: []string{
					"css/versions/bulma-no-dark-mode.min.css",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&tasks.NpmPackage{
				Name:    "@fortawesome/fontawesome-free",
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
			&tasks.NpmPackage{
				Name:    "github-markdown-css",
				Version: "5.7.0",
				Files: []string{
					"github-markdown.min.css",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&tasks.NpmPackage{
				Name:    "@fontsource-variable/nunito",
				Version: "5.1.0",
				Files: []string{
					"wght.min.css",
					"files/nunito-cyrillic-ext-wght-normal.woff2",
					"files/nunito-cyrillic-wght-normal.woff2",
					"files/nunito-vietnamese-wght-normal.woff2",
					"files/nunito-latin-ext-wght-normal.woff2",
					"files/nunito-latin-wght-normal.woff2",
				},
				BaseDestination: assetsDir,
			},
		),
	}

	for _, js := range c.Assets.Npm {
		rv = append(rv,
			runner.NewTaskGroup(
				&tasks.NpmPackage{
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

		prettyURL := true
		if pg.PrettyURL != nil && !*pg.PrettyURL {
			prettyURL = false
		}

		rv = append(rv,
			runner.NewTaskGroup(
				&tasks.Pages{
					Sources:           src,
					ExtraDependencies: pg.ExtraDependencies,
					HighlightStyle:    pg.HighlightStyle,
					PrettyURL:         prettyURL,
					BaseDestination:   pg.BaseDestination,
					Template:          pg.Template,
					TemplateCtx:       pg.TemplateCtx,
					WithSidebar:       pg.WithSidebar,
				},
			),
		)
	}

	globalPostSources := []*generators.MarkdownSource{}
	for _, ps := range c.Posts.Groups {
		sortReverse := true
		if ps.SortReverse != nil && !*ps.SortReverse {
			sortReverse = false
		}

		posts := &tasks.Posts{
			SourceDir:       ps.SourceDir,
			HighlightStyle:  ps.HighlightStyle,
			BaseDestination: ps.BaseDestination,
			Template:        ps.Template,
			TemplateCtx:     ps.TemplateCtx,
			WithSidebar:     ps.WithSidebar,
		}
		postsSources, err := posts.GetSources()
		if err != nil {
			return nil, err
		}
		globalPostSources = append(globalPostSources, postsSources...)

		rv = append(rv,
			runner.NewTaskGroup(posts),
			runner.NewTaskGroup(
				&tasks.Pagination{
					Title:           ps.Title,
					Description:     ps.Description,
					Sources:         postsSources,
					SeriesStatus:    ps.SeriesStatus,
					PostsPerPage:    ps.PostsPerPage,
					SortReverse:     sortReverse,
					HighlightStyle:  ps.HighlightStyle,
					BaseDestination: ps.BaseDestination,
					Template:        ps.TemplatePagination,
					WithSidebar:     ps.WithSidebar,
				},
			),
			runner.NewTaskGroup(
				&tasks.Pagination{
					Title:           ps.Title,
					Description:     ps.Description,
					Sources:         postsSources,
					SeriesStatus:    ps.SeriesStatus,
					PostsPerPage:    ps.PostsPerPageAtom,
					SortReverse:     true,
					Atom:            true,
					HighlightStyle:  ps.HighlightStyle,
					BaseDestination: ps.BaseDestination,
					Template:        ps.TemplateAtom,
				},
			),
		)
	}

	sortReverse := true
	if c.Posts.SortReverse != nil && !*c.Posts.SortReverse {
		sortReverse = false
	}

	rv = append(rv,
		runner.NewTaskGroup(
			&tasks.Pagination{
				Title:           c.Posts.Title,
				Description:     c.Posts.Description,
				Sources:         globalPostSources,
				PostsPerPage:    c.Posts.PostsPerPage,
				SortReverse:     sortReverse,
				HighlightStyle:  c.Posts.HighlightStyle,
				BaseDestination: c.Posts.BaseDestination,
				Template:        c.Posts.TemplatePagination,
				WithSidebar:     c.Posts.WithSidebar,
			},
		),
		runner.NewTaskGroup(
			&tasks.Pagination{
				Title:           c.Posts.Title,
				Description:     c.Posts.Description,
				Sources:         globalPostSources,
				PostsPerPage:    c.Posts.PostsPerPageAtom,
				SortReverse:     true,
				Atom:            true,
				HighlightStyle:  c.Posts.HighlightStyle,
				BaseDestination: c.Posts.BaseDestination,
				Template:        c.Posts.TemplateAtom,
			},
		),
	)

	for _, pj := range c.Projects {
		sidebar := true
		if pj.WithSidebar != nil && !*pj.WithSidebar {
			sidebar = false
		}

		for _, repo := range pj.Repositories {
			// immutable by default, only disable manually for development
			immutable := true
			if repo.Immutable != nil && !*repo.Immutable {
				immutable = false
			}

			rv = append(rv,
				runner.NewTaskGroup(
					&tasks.Project{
						Owner:           repo.Owner,
						Repo:            repo.Repo,
						BaseDestination: pj.BaseDestination,
						Template:        pj.Template,
						Immutable:       immutable,
						WithSidebar:     sidebar,
					},
				),
			)
		}
	}
	return rv, nil
}

func build() error {
	if force || cfg == nil || !cfg.IsUpToDate() {
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
	err := runner.Run(taskGroups, *fBuildDir, cfg, force)
	if force {
		// force only first time
		force = false
	}
	return err
}

func main() {
	flag.Parse()

	if *fVersion {
		md, err := meta.GetMetadata()
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		fmt.Printf("%s\n", md)
		return
	}

	force = *fForce

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
