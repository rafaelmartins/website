package generators

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rafaelmartins/website/internal/runner"
	"github.com/rafaelmartins/website/internal/templates"
	"golang.org/x/net/html"
)

type Project struct {
	Owner     string
	Repo      string
	URL       string
	Template  string
	LayoutCtx *templates.LayoutContext
	Immutable bool

	ts           time.Time
	etag         string
	lastmodified string
	readme       string
	baseurl      string
	images       []string
}

func (*Project) GetID() string {
	return "PROJECT"
}

func (p *Project) GetReader() (io.ReadCloser, error) {
	mks, baseurl, _, err := p.getReadme()
	if err != nil {
		return nil, err
	}

	mkr, err := p.renderMarkdown(mks)
	if err != nil {
		return nil, err
	}

	title, body, images, err := p.processHtml(baseurl, mkr)
	if err != nil {
		return nil, err
	}

	p.images = images

	buf := &bytes.Buffer{}
	if err := templates.Execute(buf, p.Template, nil, p.LayoutCtx, &templates.ContentContext{
		Title: title,
		URL:   p.URL,
		Entry: &templates.ContentEntry{
			Title: title,
			Body:  body,
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
	if p.Immutable && !p.ts.IsZero() {
		return []time.Time{p.ts}, nil
	}

	rv, err := templates.GetTimestamps(p.Template, !p.Immutable)
	if err != nil {
		return nil, err
	}

	_, _, ts, err := p.getReadme()
	if err != nil {
		return nil, err
	}
	return append(rv, ts), nil
}

func (p *Project) GetImmutable() bool {
	return p.Immutable
}

func (p *Project) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch == nil {
		return
	}

	for _, img := range p.images {
		rd, err := p.getContents(img)
		if err != nil {
			ch <- &runner.GeneratorByProduct{Err: err}
			break
		}

		ch <- &runner.GeneratorByProduct{
			Filename: filepath.FromSlash(img),
			Reader:   rd,
		}
	}
	close(ch)
}

func (p *Project) getReadme() (string, string, time.Time, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/repos/"+p.Owner+"/"+p.Repo+"/readme", nil)
	if err != nil {
		return "", "", time.Time{}, err
	}

	if token, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		req.Header.Add("authorization", "Bearer "+token)
	}
	req.Header.Add("accept", "application/vnd.github+json")
	req.Header.Add("x-github-api-version", "2022-11-28")
	if p.etag != "" {
		req.Header.Set("if-none-match", p.etag)
	}
	if p.lastmodified != "" {
		req.Header.Set("if-modified-since", p.lastmodified)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", time.Time{}, err
	}

	if resp.StatusCode == 304 {
		return p.readme, p.baseurl, p.ts, nil
	}

	if resp.StatusCode != 200 {
		return "", "", time.Time{}, fmt.Errorf("project: http error: %s", resp.Status)
	}

	body, baseurl, err := p.readContents(resp.Body)
	if err != nil {
		return "", "", time.Time{}, err
	}
	defer body.Close()

	readme, err := io.ReadAll(body)
	if err != nil {
		return "", "", time.Time{}, err
	}

	p.readme = string(readme)
	p.baseurl = baseurl

	if etag := resp.Header.Get("etag"); etag != "" {
		p.etag = strings.TrimPrefix(etag, "W/")
	}

	if lastmodified := resp.Header.Get("last-modified"); lastmodified != "" {
		t, err := time.Parse(time.RFC1123, lastmodified)
		if err != nil {
			return "", "", time.Time{}, err
		}

		p.lastmodified = lastmodified
		p.ts = t.UTC()
	}

	return p.readme, p.baseurl, p.ts, nil
}

func (p *Project) request(method string, path string, body io.Reader) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, "https://api.github.com/"+path, body)
	if err != nil {
		return nil, err
	}

	if token, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		req.Header.Add("authorization", "Bearer "+token)
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (p *Project) getContents(path string) (io.ReadCloser, error) {
	jbody, err := p.request("GET", "repos/"+p.Owner+"/"+p.Repo+"/contents/"+path, nil)
	if err != nil {
		return nil, err
	}

	body, _, err := p.readContents(jbody)
	return body, err
}

func (p *Project) readContents(body io.ReadCloser) (io.ReadCloser, string, error) {
	defer body.Close()

	v := struct {
		Message     string `json:"message"`
		Type        string `json:"type"`
		Encoding    string `json:"encoding"`
		Content     string `json:"content"`
		HtmlUrl     string `json:"html_url"`
		DownloadUrl string `json:"download_url"`
	}{}
	if err := json.NewDecoder(body).Decode(&v); err != nil {
		return nil, "", err
	}

	if v.Type != "file" {
		if v.Message != "" {
			return nil, "", fmt.Errorf("project: response error: %s", v.Message)
		}
		return nil, "", errors.New("project: response is not a file")
	}
	if v.HtmlUrl == "" {
		return nil, "", errors.New("project: invalid response html url")
	}

	switch v.Encoding {
	case "base64":
		if v.Content == "" {
			return nil, "", errors.New("project: invalid response base64 data")
		}
		content, err := base64.StdEncoding.DecodeString(v.Content)
		return io.NopCloser(bytes.NewBuffer(content)), v.HtmlUrl, err

	case "none":
		rsp, err := http.Get(v.DownloadUrl)
		if err != nil {
			return nil, "", err
		}
		return rsp.Body, v.HtmlUrl, err

	case "":
		return io.NopCloser(bytes.NewBufferString(v.Content)), v.HtmlUrl, nil

	default:
		return nil, "", errors.New("project: invalid response encoding")
	}
}

func (p *Project) renderMarkdown(src string) (string, error) {
	in := &bytes.Buffer{}
	if err := json.NewEncoder(in).Encode(map[string]string{
		"text":    src,
		"mode":    "gfm",
		"context": p.Owner + "/" + p.Repo,
	}); err != nil {
		return "", err
	}

	body, err := p.request("POST", "markdown", in)
	if err != nil {
		return "", err
	}
	defer body.Close()

	rv, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	return string(rv), nil
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
							if len(u.Path) > 0 && u.Path[0] == '/' {
								u.Path = u.Path[1:]
							}
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
							images = append(images, strings.TrimPrefix(strings.TrimPrefix(u.Path, "./"), "/"))
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
