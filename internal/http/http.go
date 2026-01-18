package http

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"
)

type Error struct {
	StatusCode int
	Status     string
}

func (e *Error) Error() string {
	return "http: " + e.Status
}

type RequestContext struct {
	LastModified     string
	LastModifiedTime time.Time
	ETag             string
	Body             []byte
}

func RequestWithContext(ctx *RequestContext, method string, u string, headers map[string]string, body io.Reader) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
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

	if ctx != nil && resp.StatusCode == 304 {
		defer resp.Body.Close()
		return io.NopCloser(bytes.NewReader(ctx.Body)), nil
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		return nil, &Error{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
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
