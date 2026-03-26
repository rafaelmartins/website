package project

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/yuin/goldmark/parser"
	"rafaelmartins.com/p/website/internal/frontmatter"
	"rafaelmartins.com/p/website/internal/github"
	"rafaelmartins.com/p/website/internal/markdown"
	"rafaelmartins.com/p/website/internal/opengraph"
	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/templates"
)

func splitFileName(fileName string, isReadme bool, isRoot bool) (int, string, error) {
	if isReadme {
		return 0, "", nil
	}

	s := strings.SplitN(path.Base(fileName), "_", 2)
	if len(s) != 2 {
		return 0, "", fmt.Errorf("project: page: bad file name: %s", fileName)
	}

	i, err := strconv.ParseInt(s[0], 10, 32)
	if err != nil {
		return 0, "", err
	}
	idx := int(i)

	if idx == 0 && !isRoot {
		return 0, "", fmt.Errorf("project: page: root page must be index 0: %s", fileName)
	}

	name := ""
	if !isRoot {
		name = strings.TrimSuffix(s[1], path.Ext(s[1]))
	}
	return idx, name, nil
}

type projectPageResolver struct {
	src  string
	name string
}

func newPageResolver(file *github.RepositoryFile, isReadme bool, isRoot bool) (*projectPageResolver, error) {
	if file == nil {
		return nil, errors.New("project: page resolver: file is nil")
	}

	_, name, err := splitFileName(file.Name, isReadme, isRoot)
	if err != nil {
		return nil, err
	}
	return &projectPageResolver{
		src:  file.Name,
		name: name,
	}, nil
}

func (ppr *projectPageResolver) resolveUrl(current string) string {
	current = path.Clean(current)
	if path.Clean(ppr.name) == current {
		return "./"
	}

	prefix := ""
	if current != "." {
		prefix = ".."
	}
	return path.Join(prefix, ppr.name) + "/"
}

type ProjectPage struct {
	idx    int
	name   string
	title  string
	etitle string
	toc    string
	body   string
	menu   string
	src    string
	isRoot bool

	proj     *Project
	file     *github.RepositoryFile
	meta     *frontmatter.FrontMatter
	resolver *projectPageResolver

	og     *opengraph.OpenGraph
	images []string
}

func newPage(proj *Project, file *github.RepositoryFile, isReadme bool, isRoot bool) (*ProjectPage, error) {
	if file == nil {
		return nil, errors.New("project: page: file is nil")
	}

	idx, name, err := splitFileName(file.Name, isReadme, isRoot)
	if err != nil {
		return nil, err
	}

	resolver, err := newPageResolver(file, isReadme, isRoot)
	if err != nil {
		return nil, err
	}

	rv := &ProjectPage{
		idx:    idx,
		name:   name,
		src:    file.Name,
		isRoot: isRoot,

		proj:     proj,
		file:     file,
		resolver: resolver,
	}

	if err := rv.read(); err != nil {
		return nil, err
	}
	return rv, nil
}

func (pp *ProjectPage) read() error {
	src, err := pp.file.Read()
	if err != nil {
		return err
	}

	meta, data, err := frontmatter.Parse(src)
	if err != nil {
		return err
	}

	withToc := pp.proj.Toc
	if meta.Toc != nil {
		withToc = *meta.Toc
	}

	pc := parser.NewContext()
	pc.Set(pcProjectKey, pp.proj)
	pc.Set(pcBaseUrlKey, "https://github.com/"+pp.proj.Owner+"/"+pp.proj.Repo+"/blob/"+pp.proj.proj.Head)
	pc.Set(pcCurrentPageKey, pp.name)
	pc.Set(markdown.PcTocEnable, &withToc)

	toc, body, err := markdown.Render(gmMarkdown, data, pc)
	if err != nil {
		return err
	}
	if err := pc.Get(pcErrorKey); err != nil {
		return err.(error)
	}
	pp.meta = meta
	pp.toc = toc
	pp.body = body

	pp.images = nil
	if img := pc.Get(pcImagesKey); img != nil {
		pp.images = img.([]string)
	}

	etitle := pp.meta.Title
	if etitle == "" {
		if t := pc.Get(pcTitleKey); t != nil {
			etitle = t.(string)
		}
	}
	pp.etitle = etitle

	pp.menu = etitle
	if pp.meta.Menu != "" {
		pp.menu = pp.meta.Menu
	}

	pp.title = etitle
	if pp.title != "" && !pp.isRoot {
		pp.title = pp.proj.Repo + ": " + pp.title
	}
	if pp.title == "" {
		pp.title = pp.proj.Repo
	}
	return nil
}

func (pp *ProjectPage) GetDestination() string {
	if pp.isRoot {
		return filepath.Join(pp.proj.Repo, "index.html")
	}
	return filepath.Join(pp.proj.Repo, pp.name, "index.html")
}

func (pp *ProjectPage) GetGenerator() (runner.Generator, error) {
	return pp, nil
}

func (*ProjectPage) GetID() string {
	return "PROJECT"
}

func (pp *ProjectPage) getTemplate() string {
	rv := pp.proj.Template
	if rv == "" {
		rv = "project.html"
	}
	return rv
}

func (pp *ProjectPage) GetReader() (io.ReadCloser, error) {
	tmpl := &templates.ProjectContentEntry{
		Owner:       pp.proj.Owner,
		Repo:        pp.proj.Repo,
		URL:         pp.proj.proj.HomepageUrl,
		Description: pp.proj.proj.Description,
		GoImport:    pp.proj.GoImport,
		GoRepo:      pp.proj.GoRepo,
		CDocsURL:    pp.proj.cdocsUrl,
		Stars:       pp.proj.proj.Stars,
		Watching:    pp.proj.proj.Watchers,
		Forks:       pp.proj.proj.Forks,
		IsRoot:      pp.isRoot,
	}

	if pp.proj.LocalDirectory != nil {
		if err := pp.read(); err != nil {
			return nil, err
		}
	}

	if len(pp.proj.Licenses) > 0 {
		for _, lic := range pp.proj.Licenses {
			tmpl.Licenses = append(tmpl.Licenses, &templates.ProjectContentLicense{
				SpdxId: lic.SpdxId,
				Title:  lic.Title,
			})
		}
	}

	if pp.isRoot && pp.proj.proj.LatestRelease != nil && pp.proj.proj.LatestRelease.Description != "" {
		pc := parser.NewContext()
		pc.Set(pcProjectKey, pp.proj)
		pc.Set(pcBaseUrlKey, "https://github.com/"+pp.proj.Owner+"/"+pp.proj.Repo+"/blob/"+pp.proj.proj.LatestRelease.Tag)
		pc.Set(pcCurrentPageKey, pp.name)

		_, body, err := markdown.Render(gmMarkdown, []byte(pp.proj.proj.LatestRelease.Description), pc)
		if err != nil {
			return nil, err
		}

		tmpl.LatestRelease = &templates.ProjectContentLatestRelease{
			Name: pp.proj.proj.LatestRelease.Name,
			Tag:  pp.proj.proj.LatestRelease.Tag,
			Body: body,
			URL:  pp.proj.proj.LatestRelease.Url,
		}
		for _, asset := range pp.proj.proj.LatestRelease.Assets {
			tmpl.LatestRelease.Files = append(tmpl.LatestRelease.Files,
				&templates.ProjectContentLatestReleaseFile{
					File: asset.Name,
					URL:  asset.DownloadUrl,
				},
			)
		}
		slices.SortFunc(tmpl.LatestRelease.Files, func(a *templates.ProjectContentLatestReleaseFile, b *templates.ProjectContentLatestReleaseFile) int {
			return strings.Compare(a.File, b.File)
		})
	}

	purl := path.Join(pp.proj.url, pp.name)
	if purl != "/" {
		purl += "/"
	}

	lctx := &templates.LayoutContext{
		WithSidebar: pp.isRoot,
	}

	tmpl.Menus = nil
	for _, p := range pp.proj.pages {
		// FIXME: max 9 sub pages???
		if p.idx%10 != 0 {
			continue
		}

		tmpl.Menus = append(tmpl.Menus, &templates.ProjectContentMenu{
			Active: p.idx/10 == pp.idx/10,
			URL:    p.resolver.resolveUrl(pp.name),
			Title:  p.menu,
		})
	}

	og, err := opengraph.New(pp.proj.OpenGraphImageGen, false, purl, pp.title, pp.proj.proj.Description, pp.proj.OpenGraph, pp.meta.Title, pp.meta.Description, pp.meta.OpenGraph)
	if err != nil {
		return nil, err
	}
	pp.og = og

	buf := &bytes.Buffer{}
	if err := templates.Execute(buf, pp.getTemplate(), nil, lctx, &templates.ContentContext{
		Title:       pp.title,
		Description: pp.proj.proj.Description,
		URL:         purl,
		License:     pp.proj.license,
		Toc:         pp.toc,
		Search:      true, // FIXME ???
		OpenGraph:   og.GetTemplateContext(),
		Entry: &templates.ContentEntry{
			Title:   pp.etitle,
			Body:    pp.body,
			Project: tmpl,
		},
	}); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (pp *ProjectPage) GetPaths() ([]string, error) {
	if pp.proj.Immutable && pp.proj.LocalDirectory == nil {
		return nil, nil
	}

	rv, err := templates.GetPaths(pp.getTemplate())
	if err != nil {
		return nil, err
	}

	if pp.proj.LocalDirectory != nil {
		if pp.proj.proj.Docs != nil {
			rv = append(rv, filepath.Join(*pp.proj.LocalDirectory, "docs"))
		}
		if pp.proj.proj.Readme != nil {
			rv = append(rv, filepath.Join(*pp.proj.LocalDirectory, "README.md"))
		}
	}

	if pp.proj.OpenGraphImageGen != nil {
		rv = append(rv, pp.proj.OpenGraphImageGen.GetPaths()...)
	}
	return rv, nil
}

func (pp *ProjectPage) GetImmutable() bool {
	return pp.proj.Immutable && pp.proj.LocalDirectory == nil
}

func (pp *ProjectPage) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		pp.og.GenerateByProduct(ch, "")
		close(ch)
	}
}
