package project

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/yuin/goldmark/parser"
	"rafaelmartins.com/p/website/internal/frontmatter"
	"rafaelmartins.com/p/website/internal/github"
	"rafaelmartins.com/p/website/internal/markdown"
	"rafaelmartins.com/p/website/internal/ogimage"
	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/templates"
)

type ProjectPage struct {
	idx    int
	name   string
	title  string
	menu   string
	src    string
	isRoot bool

	meta *frontmatter.FrontMatter
	data []byte

	proj *Project
	file *github.RepositoryFile

	otitle string
	images []string
}

func newPage(proj *Project, file *github.RepositoryFile, isReadme bool, isRoot bool) (*ProjectPage, error) {
	idx := 0
	name := ""

	if !isReadme {
		s := strings.SplitN(path.Base(file.Name), "_", 2)
		if len(s) != 2 {
			return nil, fmt.Errorf("project: page: bad file name: %s", file.Name)
		}

		i, err := strconv.ParseInt(s[0], 10, 32)
		if err != nil {
			return nil, err
		}
		idx = int(i)

		if !isRoot {
			name = strings.TrimSuffix(s[1], path.Ext(s[1]))
		}
	}

	rv := &ProjectPage{
		idx:    idx,
		name:   name,
		src:    file.Name,
		isRoot: isRoot,

		proj: proj,
		file: file,
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

	pp.meta, pp.data, err = frontmatter.Parse(src)
	if err != nil {
		return err
	}

	pp.title = pp.meta.Title
	if pp.title == "" {
		t, err := markdown.GetTitle(pp.data)
		if err != nil {
			return err
		}
		pp.title = t
	}

	pp.menu = pp.title
	if pp.meta.Menu != "" {
		pp.menu = pp.meta.Menu
	}
	return nil
}

func (pp *ProjectPage) resolveUrl(current string) string {
	current = path.Clean(current)
	if path.Clean(pp.name) == current {
		return "./"
	}

	prefix := ""
	if current != "." {
		prefix = ".."
	}
	return path.Join(prefix, pp.name) + "/"
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
		Date:        time.Now().UTC(),
	}

	if pp.proj.proj.LicenseSpdx != "" {
		tmpl.License.SPDX = pp.proj.proj.LicenseSpdx
		tmpl.License.URL = "https://spdx.org/licenses/" + pp.proj.proj.LicenseSpdx + ".html"
	} else if pp.proj.proj.LicenseData != nil {
		lic, err := pp.proj.proj.LicenseData.Read()
		if err != nil {
			return nil, err
		}
		tmpl.License.Data = string(lic)
	}

	if pp.isRoot && pp.proj.proj.LatestRelease != nil && pp.proj.proj.LatestRelease.Description != "" {
		pc := parser.NewContext()
		pc.Set(pcProjectKey, pp.proj)
		pc.Set(pcBaseUrlKey, "https://github.com/"+pp.proj.Owner+"/"+pp.proj.Repo+"/blob/"+pp.proj.proj.LatestRelease.Tag)
		pc.Set(pcCurrentPageKey, pp.name)

		body, err := markdown.Render(gmMarkdown, []byte(pp.proj.proj.LatestRelease.Description), pc)
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

	pc := parser.NewContext()
	pc.Set(pcProjectKey, pp.proj)
	pc.Set(pcBaseUrlKey, "https://github.com/"+pp.proj.Owner+"/"+pp.proj.Repo+"/blob/"+pp.proj.proj.Head)
	pc.Set(pcCurrentPageKey, pp.name)

	if pp.proj.LocalDirectory != nil {
		if err := pp.read(); err != nil {
			return nil, err
		}
	}

	body, err := markdown.Render(gmMarkdown, pp.data, pc)
	if err != nil {
		return nil, err
	}
	if err := pc.Get(pcErrorKey); err != nil {
		return nil, err.(error)
	}

	pp.images = nil
	if img := pc.Get(pcImagesKey); img != nil {
		pp.images = img.([]string)
	}

	title := pp.proj.Repo
	if t := pc.Get(pcTitleKey); t != nil {
		title = t.(string)
	}
	if pp.title != "" {
		title = pp.title
	}

	pp.otitle = title
	if pp.proj.OpenGraphTitle != "" {
		pp.otitle = pp.proj.OpenGraphTitle
	}

	odesc := pp.proj.proj.Description
	if pp.proj.OpenGraphDescription != "" {
		odesc = pp.proj.OpenGraphDescription
	}

	if pp.meta != nil {
		if pp.meta.OpenGraph.Title != "" {
			pp.otitle = pp.meta.OpenGraph.Title
		}
		if pp.meta.OpenGraph.Description != "" {
			odesc = pp.meta.OpenGraph.Description
		}
	}

	purl := path.Join(pp.proj.url, pp.name)
	if purl != "/" {
		purl += "/"
	}

	lctx := &templates.LayoutContext{
		WithSidebar: pp.proj.WithSidebar,
	}

	tmpl.Menus = nil
	for _, p := range pp.proj.pages {
		tmpl.Menus = append(tmpl.Menus, &templates.ProjectContentMenu{
			Active: p.name == pp.name,
			URL:    p.resolveUrl(pp.name),
			Title:  p.menu,
		})
	}

	buf := &bytes.Buffer{}
	if err := templates.Execute(buf, pp.getTemplate(), nil, lctx, &templates.ContentContext{
		Title:       title,
		Description: pp.proj.proj.Description,
		URL:         purl,
		OpenGraph: templates.OpenGraphEntry{
			Title:       pp.otitle,
			Description: odesc,
			Image:       ogimage.URL(purl),
		},
		Entry: &templates.ContentEntry{
			Title:   title,
			Body:    body,
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

	og, err := ogimage.GetPaths()
	if err != nil {
		return nil, err
	}
	return append(rv, og...), nil
}

func (pp *ProjectPage) GetImmutable() bool {
	return pp.proj.Immutable && pp.proj.LocalDirectory == nil
}

func (pp *ProjectPage) GetByProducts(ch chan *runner.GeneratorByProduct) {
	slices.Sort(pp.images)

	for _, img := range slices.Compact(pp.images) {
		rd, err := github.GetRepositoryFile(pp.proj.Owner, pp.proj.Repo, img, pp.proj.proj.Head)
		if err != nil {
			ch <- &runner.GeneratorByProduct{Err: err}
			break
		}

		ch <- &runner.GeneratorByProduct{
			Filename: filepath.FromSlash(img),
			Reader:   rd,
		}
	}

	image := pp.proj.OpenGraphImage
	ccolor := pp.proj.OpenGraphImageGenColor
	dpi := pp.proj.OpenGraphImageGenDPI
	size := pp.proj.OpenGraphImageGenSize

	if pp.meta != nil {
		// FIXME: handle image
		if pp.meta.OpenGraph.ImageGen.Color != nil {
			ccolor = pp.meta.OpenGraph.ImageGen.Color
		}
		if pp.meta.OpenGraph.ImageGen.DPI != nil {
			dpi = pp.meta.OpenGraph.ImageGen.DPI
		}
		if pp.meta.OpenGraph.ImageGen.Size != nil {
			size = pp.meta.OpenGraph.ImageGen.Size
		}
	}

	ogimage.GenerateByProduct(ch, pp.otitle, true, image, ccolor, dpi, size)
	close(ch)
}
