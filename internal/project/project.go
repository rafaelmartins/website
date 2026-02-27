package project

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"rafaelmartins.com/p/website/internal/github"
)

type ProjectLicense struct {
	SpdxId string
	Title  string
}

type Project struct {
	Owner    string
	Repo     string
	Licenses []*ProjectLicense

	Files []string

	GoImport string
	GoRepo   string

	Force                  bool
	LocalDirectory         *string
	BaseDestination        string
	Template               string
	Immutable              bool
	OpenGraphTitle         string
	OpenGraphDescription   string
	OpenGraphImage         string
	OpenGraphImageGenColor *string
	OpenGraphImageGenDPI   *float64
	OpenGraphImageGenSize  *float64

	CDocsDestination            string
	CDocsHeaders                []string
	CDocsBaseDirectory          *string
	CDocsTemplate               string
	CDocsOpenGraphTitle         string
	CDocsOpenGraphDescription   string
	CDocsOpenGraphImage         string
	CDocsOpenGraphImageGenColor *string
	CDocsOpenGraphImageGenDPI   *float64
	CDocsOpenGraphImageGenSize  *float64

	proj             *github.Repository
	subdir           string
	pages            []*ProjectPage
	pageResolvers    []*projectPageResolver
	url              string
	cdocsDestination string
	cdocsUrl         string
	license          string
}

func (p *Project) init() error {
	if p.proj != nil {
		if err := p.proj.ReloadLocalDir(); err != nil {
			return err
		}
		return p.reload()
	}

	proj, err := github.GetRepository(p.Owner, p.Repo, p.CDocsBaseDirectory, p.LocalDirectory)
	if err != nil {
		return err
	}
	p.proj = proj

	p.url = path.Join("/", p.GetBaseDestination(), p.Repo)
	if p.url != "/" {
		p.url += "/"
	}

	p.cdocsDestination = p.CDocsDestination
	if p.cdocsDestination == "" {
		p.cdocsDestination = "api"
	}

	p.cdocsUrl = ""
	if len(p.CDocsHeaders) > 0 {
		p.cdocsUrl = path.Join("/", p.GetBaseDestination(), p.Repo, p.cdocsDestination)
		if p.cdocsUrl != "/" {
			p.cdocsUrl += "/"
		}
	}

	p.license = ""
	if len(p.Licenses) > 0 {
		p.license = p.Licenses[0].SpdxId
	} else if p.proj.LicenseSpdx != "" {
		p.license = p.proj.LicenseSpdx
	}
	return p.reload()
}

func (p *Project) reload() error {
	if p.proj == nil {
		return nil
	}

	docs := p.proj.Docs
	subdir := "docs"
	if len(p.proj.Docs) == 0 {
		if p.proj.Readme == nil {
			return fmt.Errorf("project: missing readme")
		}
		docs = []*github.RepositoryFile{p.proj.Readme}
		subdir = ""
	}
	p.subdir = subdir

	// must be filled before initializing pages!
	p.pageResolvers = nil
	for idx, doc := range docs {
		res, err := newPageResolver(doc, doc == p.proj.Readme, idx == 0)
		if err != nil {
			return err
		}
		p.pageResolvers = append(p.pageResolvers, res)
	}

	p.pages = nil
	for idx, doc := range docs {
		pp, err := newPage(p, doc, doc == p.proj.Readme, idx == 0)
		if err != nil {
			return err
		}
		p.pages = append(p.pages, pp)
	}

	slices.SortFunc(p.pages, func(a *ProjectPage, b *ProjectPage) int {
		if a == nil || b == nil {
			return 0
		}
		return a.idx - b.idx
	})
	return nil
}

func (p *Project) handleImageUrl(img string, currentPage string) (string, string, error) {
	if img == "" {
		return "", "", nil
	}

	u, err := url.Parse(img)
	if err != nil {
		return "", "", err
	}

	if u.IsAbs() || u.Host != "" {
		return "", "", nil
	}

	if after, ok := strings.CutPrefix(img, "@@"); ok {
		return "", after, nil
	}

	v := strings.TrimPrefix(u.Path, "/")
	if !path.IsAbs(u.Path) {
		v = path.Join(p.subdir, u.Path)
		if v == ".." || strings.HasPrefix(v, "../") {
			return "", "", fmt.Errorf("project: path traversal not allowed: %s", img)
		}
	}

	f, err := filepath.Rel(currentPage, filepath.FromSlash(v))
	if err != nil {
		return "", "", err
	}
	return v, f, nil
}

func (p *Project) handleLinkUrl(link string, currentPage string) (bool, string, error) {
	if link == "" {
		return false, "", nil
	}

	u, err := url.Parse(link)
	if err != nil {
		return false, "", err
	}

	if u.IsAbs() || u.Host != "" {
		return false, "", nil
	}

	if after, ok := strings.CutPrefix(link, "@@"); ok {
		return false, after, nil
	}

	v := path.Join(p.subdir, u.Path)
	if path.IsAbs(u.Path) {
		v = strings.TrimPrefix(u.Path, "/")
	}
	v = path.Clean(v)

	for _, pg := range p.pageResolvers {
		if pg.src == v {
			return false, pg.resolveUrl(currentPage), nil
		}
	}
	return true, v, nil
}
