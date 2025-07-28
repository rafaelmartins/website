package generators

import (
	"io"
	"net/http"
	"strings"
	"time"

	"rafaelmartins.com/p/website/internal/runner"
)

type HTTP struct {
	Url       string
	Header    http.Header
	Immutable bool

	ts           time.Time
	etag         string
	lastmodified string
}

func (*HTTP) GetID() string {
	return "HTTP"
}

func (h *HTTP) GetReader() (io.ReadCloser, error) {
	resp, err := http.Get(h.Url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (h *HTTP) GetTimeStamps() ([]time.Time, error) {
	if h.Immutable {
		return nil, nil
	}

	r, err := http.NewRequest("HEAD", h.Url, nil)
	if err != nil {
		return nil, err
	}

	if h.Header != nil {
		r.Header = h.Header.Clone()
	}
	if h.etag != "" {
		r.Header.Set("if-none-match", h.etag)
	}
	if h.lastmodified != "" {
		r.Header.Set("if-modified-since", h.lastmodified)
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 304 {
		return []time.Time{h.ts}, nil
	}

	if etag := resp.Header.Get("etag"); etag != "" {
		h.etag = strings.TrimPrefix(etag, "W/")
	}

	if lastmodified := resp.Header.Get("last-modified"); lastmodified != "" {
		h.lastmodified = lastmodified

		t, err := time.Parse(time.RFC1123, lastmodified)
		if err != nil {
			return nil, err
		}

		h.ts = t.UTC()
		return []time.Time{h.ts}, nil
	}

	h.ts = time.Now().UTC()
	return []time.Time{h.ts}, nil
}

func (h *HTTP) GetImmutable() bool {
	return h.Immutable
}

func (*HTTP) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		close(ch)
	}
}
