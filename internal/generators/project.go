package generators

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"golang.org/x/net/html"
	"rafaelmartins.com/p/website/internal/github"
	"rafaelmartins.com/p/website/internal/ogimage"
	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/templates"
)

var reHideComments = regexp.MustCompile(`<!--\s*(/?)website-hide\s*-->`)

type Project struct {
	Owner   string
	Repo    string
	SubPage string

	SubPages []string
	DocLinks []*templates.ProjectContentDocLink

	GoImport string
	GoRepo   string

	CDocsURL string

	KicadProjects []string

	URL       string
	Template  string
	LayoutCtx *templates.LayoutContext
	Immutable bool

	OpenGraphTitle         string
	OpenGraphDescription   string
	OpenGraphImage         string
	OpenGraphImageGenColor *uint32
	OpenGraphImageGenDPI   *float64
	OpenGraphImageGenSize  *float64

	contentCtx github.RequestContext
	otitle     string
	images     []string
}

func (*Project) GetID() string {
	return "PROJECT"
}

func (p *Project) GetReader() (io.ReadCloser, error) {
	mks := ""
	baseurl := ""
	if p.SubPage != "" {
		var err error
		var rd io.ReadCloser
		rd, baseurl, err = github.Contents(&p.contentCtx, p.Owner, p.Repo, p.SubPage+".md", true)
		if err != nil {
			return nil, err
		}
		defer rd.Close()

		data, err := io.ReadAll(rd)
		if err != nil {
			return nil, err
		}
		mks = string(data)
	} else {
		var err error
		mks, baseurl, err = github.Readme(&p.contentCtx, p.Owner, p.Repo)
		if err != nil {
			return nil, err
		}
	}

	mkr, err := github.Markdown(nil, p.Owner, p.Repo, handleHideComments(mks))
	if err != nil {
		return nil, err
	}

	title, body, images, err := p.processHtml(baseurl, mkr)
	if err != nil {
		return nil, err
	}

	p.images = images

	proj := &templates.ProjectContentEntry{
		Owner:    p.Owner,
		Repo:     p.Repo,
		DocLinks: p.DocLinks,
		GoImport: p.GoImport,
		GoRepo:   p.GoRepo,
		CDocsURL: p.CDocsURL,
	}

	for _, url := range p.KicadProjects {
		prj, err := getKicadProject(url)
		if err != nil {
			return nil, err
		}
		proj.KicadProjects = append(proj.KicadProjects, prj)
	}

	rbody, err := github.Request(nil, "GET", "repos/"+p.Owner+"/"+p.Repo, nil)
	if err != nil {
		return nil, err
	}

	v := struct {
		Description      string `json:"description"`
		Homepage         string `json:"homepage"`
		ForksCount       int    `json:"forks_count"`
		StargazersCount  int    `json:"stargazers_count"`
		SubscribersCount int    `json:"subscribers_count"`
	}{}
	if err := json.Unmarshal(rbody, &v); err != nil {
		return nil, err
	}

	proj.URL = v.Homepage
	proj.Description = v.Description
	proj.Stars = v.StargazersCount
	proj.Watching = v.SubscribersCount
	proj.Forks = v.ForksCount

	withLicense := true
	lbody, err := github.Request(nil, "GET", "repos/"+p.Owner+"/"+p.Repo+"/license", nil)
	if err != nil {
		if herr, ok := err.(*github.HttpError); !ok || herr.StatusCode != 404 {
			return nil, err
		}
		withLicense = false
	}

	if withLicense {
		vl := struct {
			HtmlUrl string `json:"html_url"`
			License struct {
				SpdxId string `json:"spdx_id"`
			} `json:"license"`
		}{}
		if err := json.Unmarshal(lbody, &vl); err != nil {
			return nil, err
		}

		if vl.License.SpdxId != "NOASSERTION" {
			proj.License.SPDX = vl.License.SpdxId
		}
		proj.License.URL = vl.HtmlUrl
	}

	withRelease := true
	lrbody, err := github.Request(nil, "GET", "repos/"+p.Owner+"/"+p.Repo+"/releases/latest", nil)
	if err != nil {
		if herr, ok := err.(*github.HttpError); !ok || herr.StatusCode != 404 {
			return nil, err
		}
		withRelease = false
	}

	if withRelease {
		vlr := struct {
			Status  string `json:"status"`
			Name    string `json:"name"`
			TagName string `json:"tag_name"`
			HtmlUrl string `json:"html_url"`
			Body    string `json:"body"`
			Assets  []struct {
				Name               string `json:"name"`
				BrowserDownloadURL string `json:"browser_download_url"`
			} `json:"assets"`
		}{}
		if err := json.Unmarshal(lrbody, &vlr); err != nil {
			return nil, err
		}

		if vlr.Status == "" {
			mkd, err := github.Markdown(nil, p.Owner, p.Repo, vlr.Body)
			if err != nil {
				return nil, err
			}

			// FIXME: handle images, relative to the tag instead of the default branch
			// right now none of my projects have images in the releases... ¯\_(ツ)_/¯
			_, rbody, _, err := p.processHtml("https://github.com/"+p.Owner+"/"+p.Repo+"/blob/"+vlr.TagName+"/dummy", mkd)
			if err != nil {
				return nil, err
			}

			proj.LatestRelease = &templates.ProjectContentLatestRelease{
				Name: vlr.Name,
				Tag:  vlr.TagName,
				Body: rbody,
				URL:  vlr.HtmlUrl,
			}
			for _, asset := range vlr.Assets {
				proj.LatestRelease.Files = append(proj.LatestRelease.Files,
					&templates.ProjectContentLatestReleaseFile{
						File: asset.Name,
						URL:  asset.BrowserDownloadURL,
					},
				)
			}
		}
	}

	proj.Date = time.Now().UTC()

	p.otitle = title
	if p.OpenGraphTitle != "" {
		p.otitle = p.OpenGraphTitle
	}
	odesc := proj.Description
	if p.OpenGraphDescription != "" {
		odesc = p.OpenGraphDescription
	}

	buf := &bytes.Buffer{}
	if err := templates.Execute(buf, p.Template, nil, p.LayoutCtx, &templates.ContentContext{
		Title:       title,
		Description: proj.Description,
		URL:         p.URL,
		OpenGraph: templates.OpenGraphEntry{
			Title:       p.otitle,
			Description: odesc,
			Image:       ogimage.URL(p.URL),
		},
		Entry: &templates.ContentEntry{
			Title:   title,
			Body:    body,
			Project: proj,
		},
	}); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (p *Project) GetTimeStamps() ([]time.Time, error) {
	// we would be safe to just run this method frequently, as we support cache with
	// etag/last-modified, but it is easier to just disable this manually when adding
	// a new project than spam github servers for no good reason.
	if p.Immutable {
		return nil, nil
	}

	rv, err := templates.GetTimestamps(p.Template, !p.Immutable)
	if err != nil {
		return nil, err
	}

	og, err := ogimage.GetTimeStamps()
	if err != nil {
		return nil, err
	}
	rv = append(rv, og...)

	if p.SubPage != "" {
		_, _, err = github.Contents(&p.contentCtx, p.Owner, p.Repo, p.SubPage+".md", false)
	} else {
		_, _, err = github.Readme(&p.contentCtx, p.Owner, p.Repo)
	}
	if err != nil {
		return nil, err
	}

	return append(rv, p.contentCtx.LastModifiedTime), nil
}

func (p *Project) GetImmutable() bool {
	return p.Immutable
}

func (p *Project) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch == nil {
		return
	}

	for _, img := range p.images {
		rd, _, err := github.Contents(nil, p.Owner, p.Repo, strings.TrimPrefix(img, "images/"), true)
		if err != nil {
			ch <- &runner.GeneratorByProduct{Err: err}
			break
		}

		ch <- &runner.GeneratorByProduct{
			Filename: filepath.FromSlash(img),
			Reader:   rd,
		}
	}

	ogimage.GenerateByProduct(ch, p.otitle, true, p.OpenGraphImage, p.OpenGraphImageGenColor, p.OpenGraphImageGenDPI, p.OpenGraphImageGenSize)
	close(ch)
}

func (p *Project) processHtml(baseUrl string, data string) (string, string, []string, error) {
	burl, err := url.Parse(baseUrl)
	if err != nil {
		return "", "", nil, err
	}

	tk := html.NewTokenizer(bytes.NewBufferString(data))

	buf := &bytes.Buffer{}

	title := ""
	capturingTitleTag := ""
	images := []string{}

	for {
		typ := tk.Next()

		if typ == html.ErrorToken {
			if err := tk.Err(); err != nil && err != io.EOF {
				return "", "", nil, err
			}
			return title, strings.TrimSpace(buf.String()), images, nil
		}

		tok := tk.Token()
		tag := tok.DataAtom.String()

		switch typ {
		case html.StartTagToken:
			switch tag {
			case "a":
				for idx := range tok.Attr {
					if strings.ToLower(tok.Attr[idx].Key) == "href" {
						u, err := url.Parse(tok.Attr[idx].Val)
						if err != nil {
							return "", "", nil, err
						}

						if !u.IsAbs() && (len(u.Fragment) == 0 || len(u.Path) != 0) {
							if rv := fixSubPageHtmlLink(u.Path, p.SubPage, p.SubPages); rv != "" {
								tok.Attr[idx].Val = rv
								continue
							}
							u.Path = strings.TrimPrefix(u.Path, "/")
							tok.Attr[idx].Val = burl.ResolveReference(u).String()
						}
					}
				}

			case "img":
				for idx := range tok.Attr {
					if strings.ToLower(tok.Attr[idx].Key) == "src" {
						u, err := url.Parse(tok.Attr[idx].Val)
						if err != nil {
							return "", "", nil, err
						}

						if !u.IsAbs() {
							if rv := fixSubPageHtmlImg(u.Path, p.SubPage); rv != "" {
								tok.Attr[idx].Val = rv
								if !slices.Contains(images, rv) {
									images = append(images, rv)
								}
							}
						}
					}
				}

			case "h1", "h2", "h3", "h4", "h5", "h6":
				if capturingTitleTag == "" && title == "" {
					capturingTitleTag = tag
					continue
				}
			}

		case html.TextToken:
			if capturingTitleTag != "" {
				if v := strings.TrimSpace(tok.Data); v != "" {
					title = v
				}
				continue
			}

		case html.EndTagToken:
			if capturingTitleTag != "" && capturingTitleTag == tag {
				capturingTitleTag = ""
				continue
			}
		}

		if capturingTitleTag == "" {
			fmt.Fprint(buf, tok)
		}
	}
}

func fixSubPageHtmlLink(link string, subpage string, subpages []string) string {
	if !strings.HasSuffix(link, ".md") {
		return ""
	}
	link = strings.TrimSuffix(link, ".md")

	skipFolder := ".."
	if subpage == "" {
		subpage = "."
		skipFolder = "."
	}
	absSubpage := path.Clean("/" + subpage)

	absLink := link
	if !path.IsAbs(link) {
		if p := path.Join("/dummy/"+subpage, skipFolder, link); !strings.HasPrefix(p, "/dummy/") {
			return ""
		}
		absLink = path.Join(absSubpage, skipFolder, link)
	}

	for _, sp := range subpages {
		if abssp := path.Clean("/" + sp); abssp == absLink {
			rv, err := filepath.Rel(filepath.FromSlash(absSubpage), filepath.FromSlash(absLink))
			if err != nil {
				return ""
			}
			if rv == "." {
				return ""
			}
			return filepath.ToSlash(rv) + "/"
		}
	}
	return ""
}

func fixSubPageHtmlImg(img string, subpage string) string {
	skipFolder := ".."
	if subpage == "" {
		subpage = "."
		skipFolder = "."
	}
	absSubpage := path.Clean("/" + subpage)

	absImg := img
	if !path.IsAbs(img) {
		if p := path.Join("/dummy/"+subpage, skipFolder, img); !strings.HasPrefix(p, "/dummy/") {
			return ""
		}
		absImg = path.Join(absSubpage, skipFolder, img)
	}
	return path.Join("images", absImg)
}

func handleHideComments(mkd string) string {
	m := reHideComments.FindAllStringSubmatchIndex(mkd, -1)
	if m == nil {
		return mkd
	}

	idx := 0
	hiding := false
	buf := bytes.Buffer{}
	for _, match := range m {
		if len(match) != 4 {
			continue
		}

		if match[0] > idx && !hiding {
			buf.WriteString(mkd[idx:match[0]])
		}

		idx = match[1]
		hiding = match[2] == match[3]
	}
	if !hiding {
		buf.WriteString(mkd[idx:])
	}
	return buf.String()
}

func getKicadProject(iurl string) (*templates.ProjectContentKicadProject, error) {
	resp, err := http.Get(iurl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("project: kicad: http error: %d - %s", resp.StatusCode, resp.Status)
	}

	data := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	version, ok := data["version"]
	if !ok {
		return nil, errors.New("project: kicad: version field not found")
	}

	fversion, ok := version.(float64)
	if !ok {
		return nil, errors.New("project: kicad: version is not an integer")
	}
	iversion := int(fversion)

	rv := templates.ProjectContentKicadProject{}

	switch iversion {
	case 1:
		rv.Name = data["name"].(string)
		rv.Revision = data["revision"].(string)

		if d, ok := data["sch-export-pdf"].(string); ok && d != "" {
			sch, err := url.JoinPath(iurl, d)
			if err != nil {
				return nil, err
			}
			rv.SchExportPdf = sch
		}

		if d, ok := data["pcb-ibom"].(string); ok && d != "" {
			ibom, err := url.JoinPath(iurl, d)
			if err != nil {
				return nil, err
			}
			rv.PcbIbom = ibom
		}

		if d, ok := data["pcb-gerber"].(string); ok && d != "" {
			gerber, err := url.JoinPath(iurl, d)
			if err != nil {
				return nil, err
			}
			rv.PcbGerber = gerber
		}

		if d, ok := data["pcb-render"].(map[string]any); ok && d != nil {
			for side, v := range d {
				if side != "top" && side != "bottom" {
					continue
				}

				if l, ok := v.([]any); ok {
					for _, m := range l {
						if mm, ok := m.(map[string]any); ok {
							img, err := url.JoinPath(iurl, mm["file"].(string))
							if err != nil {
								return nil, err
							}

							switch side {
							case "":
								rv.PcbRenderMontage = append(rv.PcbRenderMontage, &templates.ProjectContentKicadProjectPcbRenderFile{
									Scale: int(mm["scale"].(float64)),
									File:  img,
								})

							case "top":
								rv.PcbRenderTop = append(rv.PcbRenderTop, &templates.ProjectContentKicadProjectPcbRenderFile{
									Scale: int(mm["scale"].(float64)),
									File:  img,
								})

							case "bottom":
								rv.PcbRenderBottom = append(rv.PcbRenderBottom, &templates.ProjectContentKicadProjectPcbRenderFile{
									Scale: int(mm["scale"].(float64)),
									File:  img,
								})
							}
						}
					}
				}
			}
		}

		if d, ok := data["tools"].(map[string]any); ok {
			tools := map[string]string{}
			for tool, version := range d {
				tools[tool] = version.(string)
			}
			rv.Tools = tools
		}

	default:
		return nil, fmt.Errorf("project: kicad: unsupported version: %d", iversion)
	}

	scale := 0
	for _, rf := range rv.PcbRenderTop {
		if rf.Scale >= scale {
			rv.PcbRenderTopMain = rf.File
			scale = rf.Scale
		}
	}

	scale = 0
	for _, rf := range rv.PcbRenderBottom {
		if rf.Scale >= scale {
			rv.PcbRenderBottomMain = rf.File
			scale = rf.Scale
		}
	}

	fmt.Printf("%+v\n", rv)

	return &rv, nil
}
