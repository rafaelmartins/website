package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"rafaelmartins.com/p/website/internal/assets"
	"rafaelmartins.com/p/website/internal/cdocs"
	"rafaelmartins.com/p/website/internal/config"
	"rafaelmartins.com/p/website/internal/kicad"
	"rafaelmartins.com/p/website/internal/meta"
	"rafaelmartins.com/p/website/internal/ogimage"
	"rafaelmartins.com/p/website/internal/project"
	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/tasks"
	"rafaelmartins.com/p/website/internal/templates"
	"rafaelmartins.com/p/website/internal/webserver"
)

type stringMapFlag map[string]string

func (s *stringMapFlag) String() string {
	return fmt.Sprint(*s)
}

func (s *stringMapFlag) Set(value string) error {
	p := strings.SplitN(value, "=", 2)
	if len(p) != 2 {
		return fmt.Errorf("failed to parse value, missing '=': %s", value)
	}

	if *s == nil {
		*s = stringMapFlag{}
	}
	map[string]string(*s)[p[0]] = p[1]
	return nil
}

func stringSlice(name string, usage string) *stringMapFlag {
	rv := new(stringMapFlag)
	flag.Var(rv, name, usage)
	return rv
}

var (
	fBuildDir   = flag.String("d", "_build", "build directory")
	fConfigFile = flag.String("c", "config.yml", "configuration file")
	fListenAddr = flag.String("a", ":3000", "development web server listen address")
	fCDocs      = flag.String("x", "", "dump cdocs ast and template context for given header and exit")
	fLocalDir   = stringSlice("l", "use local git repository for given project (format \"owner/repo=dir\")")
	fRunServer  = flag.Bool("r", false, "run development server")
	fForce      = flag.Bool("f", false, "force re-running all tasks")
	fKicad      = flag.Bool("k", false, "kicad assets mode")
	fVersion    = flag.Bool("v", false, "show version and exit")

	cfg        *config.Config      = nil
	kcfg       *kicad.Config       = nil
	taskGroups []*runner.TaskGroup = nil

	force = false
)

func getTaskGroups(c *config.Config) ([]*runner.TaskGroup, error) {
	assetsDir := c.Assets.BaseDestination
	if assetsDir == "" {
		assetsDir = "assets"
	}
	templates.SetAssetsDir(assetsDir)

	rv := []*runner.TaskGroup{
		// assets embedded
		runner.NewTaskGroup(
			&tasks.Embed{
				FS:              assets.Assets,
				BaseDestination: assetsDir,
			},
		),

		// assets required by embedded templates
		runner.NewTaskGroup(
			&tasks.NpmPackage{
				Name:    "anchor-js",
				Version: "5.0.0",
				Files: []string{
					"anchor.js",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&tasks.NpmPackage{
				Name:    "bulma",
				Version: "1.0.4",
				Files: []string{
					"css/versions/bulma-no-dark-mode.css",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&tasks.NpmPackage{
				Name:    "@fortawesome/fontawesome-free",
				Version: "7.1.0",
				Files: []string{
					"css/all.css",
					"webfonts/fa-brands-400.woff2",
					"webfonts/fa-regular-400.woff2",
					"webfonts/fa-solid-900.woff2",
					"webfonts/fa-v4compatibility.woff2",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&tasks.NpmPackage{
				Name:    "@fontsource-variable/atkinson-hyperlegible-next",
				Version: "5.2.6",
				Files: []string{
					"wght.css",
					"wght-italic.css",
					"files/atkinson-hyperlegible-next-latin-wght-normal.woff2",
					"files/atkinson-hyperlegible-next-latin-wght-italic.woff2",
					"files/atkinson-hyperlegible-next-latin-ext-wght-normal.woff2",
					"files/atkinson-hyperlegible-next-latin-ext-wght-italic.woff2",
				},
				BaseDestination: assetsDir,
			},
		),
		runner.NewTaskGroup(
			&tasks.NpmPackage{
				Name:    "@fontsource-variable/atkinson-hyperlegible-mono",
				Version: "5.2.5",
				Files: []string{
					"wght.css",
					"wght-italic.css",
					"files/atkinson-hyperlegible-mono-latin-wght-normal.woff2",
					"files/atkinson-hyperlegible-mono-latin-wght-italic.woff2",
					"files/atkinson-hyperlegible-mono-latin-ext-wght-normal.woff2",
					"files/atkinson-hyperlegible-mono-latin-ext-wght-italic.woff2",
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
				Slug:    s.Slug,
				File:    s.File,
				License: s.License,

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
					PrettyURL:         prettyURL,
					BaseDestination:   pg.BaseDestination,
					Template:          pg.Template,
					TemplateCtx:       pg.TemplateCtx,
					WithSidebar:       pg.WithSidebar,
				},
			),
		)
	}

	globalPostSources := []*tasks.PostsSources{}
	for _, ps := range c.Posts.Groups {
		sortReverse := true
		if ps.SortReverse != nil && !*ps.SortReverse {
			sortReverse = false
		}

		posts := &tasks.Posts{
			SourceDir: tasks.PostsSources{
				Dir:             ps.SourceDir,
				BaseDestination: ps.BaseDestination,
			},
			Template:    ps.Template,
			TemplateCtx: ps.TemplateCtx,
			WithSidebar: ps.WithSidebar,
		}
		postsSources := &tasks.PostsSources{
			Dir:             ps.SourceDir,
			BaseDestination: ps.BaseDestination,
		}
		globalPostSources = append(globalPostSources, postsSources)

		rv = append(rv,
			runner.NewTaskGroup(posts),
			runner.NewTaskGroup(
				&tasks.Pagination{
					Title:           ps.Title,
					Description:     ps.Description,
					SourceDirs:      []*tasks.PostsSources{postsSources},
					PostsPerPage:    ps.PostsPerPage,
					SortReverse:     sortReverse,
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
					SourceDirs:      []*tasks.PostsSources{postsSources},
					PostsPerPage:    ps.PostsPerPageAtom,
					SortReverse:     true,
					Atom:            true,
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
				SourceDirs:      globalPostSources,
				PostsPerPage:    c.Posts.PostsPerPage,
				SortReverse:     sortReverse,
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
				SourceDirs:      globalPostSources,
				PostsPerPage:    c.Posts.PostsPerPageAtom,
				SortReverse:     true,
				Atom:            true,
				BaseDestination: c.Posts.BaseDestination,
				Template:        c.Posts.TemplateAtom,
				TemplateCtx:     c.Posts.TemplateCtx,
			},
		),
	)

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

	for _, pj := range c.Projects {
		for _, repo := range pj.Repositories {
			localDir := (*string)(nil)
			if v, ok := map[string]string(*fLocalDir)[repo.Owner+"/"+repo.Repo]; ok {
				log.Printf("project %s/%s using local directory: %s", repo.Owner, repo.Repo, v)
				localDir = &v
			}

			// immutable by default, only disable manually for development
			immutable := true
			if repo.Immutable != nil && !*repo.Immutable {
				immutable = false
			}

			licenses := []*project.ProjectLicense{}
			for _, lic := range repo.Licenses {
				licenses = append(licenses, &project.ProjectLicense{
					SpdxId: lic.SpdxId,
					Title:  lic.Title,
				})
			}

			rv = append(rv,
				runner.NewTaskGroup(
					&project.Project{
						Owner:    repo.Owner,
						Repo:     repo.Repo,
						Licenses: licenses,

						GoImport: repo.Go.Import,
						GoRepo:   repo.Go.Repo,

						Force:                  *fForce,
						LocalDirectory:         localDir,
						BaseDestination:        pj.BaseDestination,
						Template:               pj.Template,
						Immutable:              immutable,
						OpenGraphTitle:         repo.OpenGraph.Title,
						OpenGraphDescription:   repo.OpenGraph.Description,
						OpenGraphImage:         repo.OpenGraph.Image,
						OpenGraphImageGenColor: repo.OpenGraph.ImageGen.Color,
						OpenGraphImageGenDPI:   repo.OpenGraph.ImageGen.DPI,
						OpenGraphImageGenSize:  repo.OpenGraph.ImageGen.Size,

						CDocsDestination:            repo.CDocs.Destination,
						CDocsHeaders:                repo.CDocs.Headers,
						CDocsBaseDirectory:          repo.CDocs.BaseDirectory,
						CDocsTemplate:               repo.CDocs.Template,
						CDocsOpenGraphTitle:         repo.CDocs.OpenGraph.Title,
						CDocsOpenGraphDescription:   repo.CDocs.OpenGraph.Description,
						CDocsOpenGraphImage:         repo.CDocs.OpenGraph.Image,
						CDocsOpenGraphImageGenColor: repo.CDocs.OpenGraph.ImageGen.Color,
						CDocsOpenGraphImageGenDPI:   repo.CDocs.OpenGraph.ImageGen.DPI,
						CDocsOpenGraphImageGenSize:  repo.CDocs.OpenGraph.ImageGen.Size,
					},
				),
			)

			for _, kicadProject := range repo.Kicad.Projects {
				rv = append(rv,
					runner.NewTaskGroup(
						&tasks.Kicad{
							Owner:    repo.Owner,
							Repo:     repo.Repo,
							UrlOrTag: kicadProject,

							BaseDestination:     pj.BaseDestination,
							Destination:         repo.Kicad.Destination,
							IncludeNameRevision: repo.Kicad.IncludeNameRevision,
							Immutable:           immutable,
						},
					),
				)
			}
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

func buildKicad() error {
	if force || kcfg == nil || !kcfg.IsUpToDate() {
		var err error
		kcfg, err = kicad.NewConfig(*fConfigFile)
		if err != nil {
			return err
		}

		tg, err := kicad.GetTasksGroups(kcfg)
		if err != nil {
			return err
		}
		taskGroups = tg
	}
	err := runner.Run(taskGroups, *fBuildDir, kcfg, force)
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

	buildFunc := build
	if *fKicad {
		buildFunc = buildKicad
	}

	if *fRunServer {
		if err := webserver.ListenAndServeWithReloader(*fListenAddr, *fBuildDir, buildFunc); err != nil {
			log.Fatalf("error: %s", err)
		}
	} else {
		if err := buildFunc(); err != nil {
			log.Fatalf("error: %s", err)
		}
	}
}
