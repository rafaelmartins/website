package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rafaelmartins/website/internal/cdocs"
	"github.com/rafaelmartins/website/internal/config"
	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/meta"
	"github.com/rafaelmartins/website/internal/ogimage"
	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/tasks"
	"github.com/rafaelmartins/website/internal/templates"
	"github.com/rafaelmartins/website/internal/webserver"
)

var (
	fBuildDir   = flag.String("d", "_build", "build directory")
	fConfigFile = flag.String("c", "config.yml", "configuration file")
	fListenAddr = flag.String("a", ":3000", "development web server listen address")
	fCDocs      = flag.String("x", "", "dump cdocs ast and template context for given header and exit")
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
				Version: "1.0.3",
				Files: []string{
					"css/versions/bulma-no-dark-mode.min.css",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&tasks.NpmPackage{
				Name:    "@fortawesome/fontawesome-free",
				Version: "6.7.2",
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
				Version: "5.8.1",
				Files: []string{
					"github-markdown-light.min.css",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&tasks.NpmPackage{
				Name:    "@fontsource-variable/nunito",
				Version: "5.1.1",
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
		src := []*tasks.PageSource{}
		for _, s := range pg.Sources {
			gen := true
			if s.OpenGraph.ImageGen.Generate != nil && !*s.OpenGraph.ImageGen.Generate {
				gen = false
			}

			src = append(src, &tasks.PageSource{
				Slug: s.Slug,
				File: s.File,

				OpenGraphTitle:         s.OpenGraph.Title,
				OpenGraphDescription:   s.OpenGraph.Description,
				OpenGraphImage:         s.OpenGraph.Image,
				OpenGraphImageGenerate: gen,
				OpenGraphImageGenColor: s.OpenGraph.ImageGen.Color,
				OpenGraphImageGenDPI:   s.OpenGraph.ImageGen.DPI,
				OpenGraphImageGenSize:  s.OpenGraph.ImageGen.Size,
			})
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
					PostsPerPage:    ps.PostsPerPage,
					SortReverse:     sortReverse,
					HighlightStyle:  ps.HighlightStyle,
					BaseDestination: ps.BaseDestination,
					Template:        ps.TemplatePagination,
					TemplateCtx:     ps.TemplateCtx,
					WithSidebar:     ps.WithSidebar,

					OpenGraphTitle:         ps.OpenGraph.Title,
					OpenGraphDescription:   ps.OpenGraph.Description,
					OpenGraphImage:         ps.OpenGraph.Image,
					OpenGraphImageGenColor: ps.OpenGraph.ImageGen.Color,
					OpenGraphImageGenDPI:   ps.OpenGraph.ImageGen.DPI,
					OpenGraphImageGenSize:  ps.OpenGraph.ImageGen.Size,
				},
			),
			runner.NewTaskGroup(
				&tasks.Pagination{
					Title:           ps.Title,
					Description:     ps.Description,
					Sources:         postsSources,
					PostsPerPage:    ps.PostsPerPageAtom,
					SortReverse:     true,
					Atom:            true,
					HighlightStyle:  ps.HighlightStyle,
					BaseDestination: ps.BaseDestination,
					Template:        ps.TemplateAtom,
					TemplateCtx:     ps.TemplateCtx,
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
				TemplateCtx:     c.Posts.TemplateCtx,
				WithSidebar:     c.Posts.WithSidebar,

				OpenGraphTitle:         c.Posts.OpenGraph.Title,
				OpenGraphDescription:   c.Posts.OpenGraph.Description,
				OpenGraphImage:         c.Posts.OpenGraph.Image,
				OpenGraphImageGenColor: c.Posts.OpenGraph.ImageGen.Color,
				OpenGraphImageGenDPI:   c.Posts.OpenGraph.ImageGen.DPI,
				OpenGraphImageGenSize:  c.Posts.OpenGraph.ImageGen.Size,
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
				TemplateCtx:     c.Posts.TemplateCtx,
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

			dsidebar := true
			if repo.CDocs.WithSidebar != nil && !*repo.CDocs.WithSidebar {
				dsidebar = false
			}

			rv = append(rv,
				runner.NewTaskGroup(
					&tasks.Project{
						Owner: repo.Owner,
						Repo:  repo.Repo,

						CDocsDestination:            repo.CDocs.Destination,
						CDocsHeaders:                repo.CDocs.Headers,
						CDocsBaseDirectory:          repo.CDocs.BaseDirectory,
						CDocsTemplate:               repo.CDocs.Template,
						CDocsWithSidebar:            dsidebar,
						CDocsOpenGraphTitle:         repo.CDocs.OpenGraph.Title,
						CDocsOpenGraphDescription:   repo.CDocs.OpenGraph.Description,
						CDocsOpenGraphImage:         repo.CDocs.OpenGraph.Image,
						CDocsOpenGraphImageGenColor: repo.CDocs.OpenGraph.ImageGen.Color,
						CDocsOpenGraphImageGenDPI:   repo.CDocs.OpenGraph.ImageGen.DPI,
						CDocsOpenGraphImageGenSize:  repo.CDocs.OpenGraph.ImageGen.Size,

						BaseDestination:        pj.BaseDestination,
						Template:               pj.Template,
						Immutable:              immutable,
						WithSidebar:            sidebar,
						OpenGraphTitle:         repo.OpenGraph.Title,
						OpenGraphDescription:   repo.OpenGraph.Description,
						OpenGraphImage:         repo.OpenGraph.Image,
						OpenGraphImageGenColor: repo.OpenGraph.ImageGen.Color,
						OpenGraphImageGenDPI:   repo.OpenGraph.ImageGen.DPI,
						OpenGraphImageGenSize:  repo.OpenGraph.ImageGen.Size,
					},
				),
			)
		}
	}

	for _, qr := range c.QRCode {
		rv = append(rv,
			runner.NewTaskGroup(
				&tasks.QRCode{
					SourceFile:      qr.SourceFile,
					SourceContent:   qr.SourceContent,
					DestinationFile: qr.DestinationFile,
					Size:            qr.Size,
					ForegroundColor: qr.ForegroundColor,
					BackgroundColor: qr.BackgroundColor,
					WithoutBorders:  qr.WithoutBorders,
				},
			),
		)
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

		if err := ogimage.SetGlobals(
			cfg.OpenGraphImageGen.Template,
			cfg.OpenGraphImageGen.Mask.MinX,
			cfg.OpenGraphImageGen.Mask.MinY,
			cfg.OpenGraphImageGen.Mask.MaxX,
			cfg.OpenGraphImageGen.Mask.MaxY,
			cfg.OpenGraphImageGen.DefaultColor,
			cfg.OpenGraphImageGen.DefaultDPI,
			cfg.OpenGraphImageGen.DefaultSize,
		); err != nil {
			return err
		}

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

	if *fCDocs != "" {
		fp, err := os.Open(*fCDocs)
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		hdr, err := cdocs.Parse(*fCDocs, fp)
		if hdr != nil {
			hdr.Dump(os.Stdout)
		}
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		ctx, err := cdocs.NewTemplateCtx([]*cdocs.TemplateCtxHeader{{Filename: *fCDocs, Header: hdr}})
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		ctx.Dump(os.Stdout)
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
