package github

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type RequestContext struct {
	LastModified     string
	LastModifiedTime time.Time
	ETag             string
	Body             []byte
}

type HttpError struct {
	StatusCode int
	Status     string
	Message    string
}

func (e *HttpError) Error() string {
	s := "github: http: " + e.Status
	if e.Message != "" {
		s += ": " + e.Message
	}
	return s
}

func Request(ctx *RequestContext, method string, path string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, "https://api.github.com/"+path, body)
	if err != nil {
		return nil, err
	}

	if token, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		req.Header.Add("authorization", "Bearer "+token)
	}
	req.Header.Add("accept", "application/vnd.github+json")
	req.Header.Add("x-github-api-version", "2022-11-28")
	if ctx != nil {
		if ctx.ETag != "" {
			req.Header.Set("if-none-match", ctx.ETag)
		}
		if ctx.LastModified != "" {
			req.Header.Set("if-modified-since", ctx.LastModified)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if ctx != nil && resp.StatusCode == 304 {
		return ctx.Body, nil
	}

	if resp.StatusCode >= 400 {
		herr := &HttpError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}

		v := &struct {
			Message string `json:"message"`
		}{}
		if err := json.NewDecoder(resp.Body).Decode(v); err == nil && v.Message != "" {
			herr.Message = v.Message
		}
		return nil, herr
	}
	rv, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		ctx.Body = rv

		if etag := resp.Header.Get("etag"); etag != "" {
			ctx.ETag = strings.TrimPrefix(etag, "W/")
		}

		if lastModified := resp.Header.Get("last-modified"); lastModified != "" {
			lastModifiedTime, err := time.Parse(time.RFC1123, lastModified)
			if err != nil {
				return nil, err
			}

			ctx.LastModified = lastModified
			ctx.LastModifiedTime = lastModifiedTime.UTC()
		}
	}
	return rv, nil
}

func Markdown(ctx *RequestContext, owner string, repo string, src string) (string, error) {
	in := &bytes.Buffer{}
	if err := json.NewEncoder(in).Encode(map[string]string{
		"text":    src,
		"mode":    "gfm",
		"context": owner + "/" + repo,
	}); err != nil {
		return "", err
	}

	body, err := Request(ctx, "POST", "markdown", in)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func readContents(body []byte) (io.ReadCloser, string, error) {
	v := struct {
		Message     string `json:"message"`
		Type        string `json:"type"`
		Encoding    string `json:"encoding"`
		Content     string `json:"content"`
		HtmlUrl     string `json:"html_url"`
		DownloadUrl string `json:"download_url"`
	}{}
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, "", err
	}

	if v.Type != "file" {
		if v.Message != "" {
			return nil, "", fmt.Errorf("github: %s", v.Message)
		}
		return nil, "", errors.New("github: response is not a file")
	}
	if v.HtmlUrl == "" {
		return nil, "", errors.New("github: invalid response html url")
	}

	switch v.Encoding {
	case "base64":
		if v.Content == "" {
			return nil, "", errors.New("github: invalid response base64 data")
		}
		content, err := base64.StdEncoding.DecodeString(v.Content)
		return io.NopCloser(bytes.NewBuffer(content)), v.HtmlUrl, err

	case "none":
		// TODO: cache
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

func Readme(ctx *RequestContext, owner string, repo string) (string, string, error) {
	body, err := Request(ctx, "GET", "repos/"+owner+"/"+repo+"/readme", nil)
	if err != nil {
		return "", "", err
	}

	resp, u, err := readContents(body)
	if err != nil {
		return "", "", err
	}
	defer resp.Close()

	rv, err := io.ReadAll(resp)
	if err != nil {
		return "", "", err
	}
	return string(rv), u, nil
}

func Contents(ctx *RequestContext, owner string, repo string, path string) (io.ReadCloser, error) {
	body, err := Request(ctx, "GET", "repos/"+owner+"/"+repo+"/contents/"+path, nil)
	if err != nil {
		return nil, err
	}

	rv, _, err := readContents(body)
	return rv, err
}
