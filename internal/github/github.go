package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var token = func() string {
	if rv, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		return rv
	}

	gh, err := exec.LookPath("gh")
	if err != nil {
		return ""
	}

	out, err := exec.Command(gh, "auth", "token").Output()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			log.Printf("warning: failed to request gh auth token: %s", err.Stderr)
		} else {
			log.Printf("warning: failed to request gh auth token: %s", err)
		}
		return ""
	}
	return strings.TrimSpace(string(out))
}()

var (
	rlMutex   sync.Mutex
	rlGraphql *int
	rlRest    *int
)

type Error struct {
	StatusCode int
	Status     string
	Message    string
}

func (e *Error) Error() string {
	s := "github: http: " + e.Status
	if e.Message != "" {
		s += ": " + e.Message
	}
	return s
}

type RequestContext struct {
	LastModified     string
	LastModifiedTime time.Time
	ETag             string
	Body             []byte
}

func RequestWithContext(ctx *RequestContext, method string, path string, headers map[string]string, body io.Reader) (io.ReadCloser, error) {
	uu, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := "https://api.github.com/" + path
	if uu.IsAbs() {
		u = path
	}

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("authorization", "Bearer "+token)
	}

	if !uu.IsAbs() {
		req.Header.Set("accept", "application/vnd.github+json")
		req.Header.Set("x-github-api-version", "2022-11-28")
	}

	if ctx != nil {
		if ctx.ETag != "" {
			req.Header.Set("if-none-match", ctx.ETag)
		}
		if ctx.LastModified != "" {
			req.Header.Set("if-modified-since", ctx.LastModified)
		}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	resource := resp.Header.Get("x-ratelimit-resource")
	remaining := resp.Header.Get("x-ratelimit-remaining")
	if remaining != "" {
		r, err := strconv.ParseInt(remaining, 10, 32)
		if err != nil {
			return nil, err
		}

		rlMutex.Lock()
		defer rlMutex.Unlock()

		rr := int(r)
		switch resource {
		case "core":
			rlRest = &rr
		case "graphql":
			rlGraphql = &rr
		default:
			return nil, fmt.Errorf("github: unknown ratelimit resource: %s", resource)
		}
	}

	if ctx != nil && resp.StatusCode == 304 {
		defer resp.Body.Close()
		return io.NopCloser(bytes.NewReader(ctx.Body)), nil
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		err := &Error{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
		v := struct {
			Message string `json:"message"`
		}{}
		if json.NewDecoder(resp.Body).Decode(&v) == nil && v.Message != "" {
			err.Message = v.Message
		}
		return nil, err
	}

	if ctx == nil {
		return resp.Body, nil
	}

	defer resp.Body.Close()

	v, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	ctx.Body = v

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
	return io.NopCloser(bytes.NewReader(ctx.Body)), nil
}

func Request(method string, path string, headers map[string]string, body io.Reader) (io.ReadCloser, error) {
	return RequestWithContext(nil, method, path, headers, body)
}

func DumpRatelimit() {
	rlMutex.Lock()
	defer rlMutex.Unlock()

	s := "ratelimit:"
	if rlGraphql != nil {
		s += fmt.Sprintf(" graphql=%d;", *rlGraphql)
	}
	if rlRest != nil {
		s += fmt.Sprintf(" rest=%d;", *rlRest)
	}
	log.Print(s)
}
