package generators

import (
	"io"
	"net/http"
	"time"

	"rafaelmartins.com/p/website/internal/runner"
)

type HttpError struct {
	StatusCode int
	Status     string
}

func (e *HttpError) Error() string {
	return "http: " + e.Status
}

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
	r, err := http.NewRequest("GET", h.Url, nil)
	if err != nil {
		return nil, err
	}

	if h.Header != nil {
		r.Header = h.Header.Clone()
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, &HttpError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}
	return resp.Body, nil
}

func (*HTTP) GetPaths() ([]string, error) {
	return nil, nil
}

func (h *HTTP) GetImmutable() bool {
	return h.Immutable
}

func (*HTTP) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		close(ch)
	}
}
