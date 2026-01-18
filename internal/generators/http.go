package generators

import (
	"io"

	"rafaelmartins.com/p/website/internal/http"
	"rafaelmartins.com/p/website/internal/runner"
)

type HTTP struct {
	Url       string
	Header    map[string]string
	Immutable bool

	ctx http.RequestContext
}

func (*HTTP) GetID() string {
	return "HTTP"
}

func (h *HTTP) GetReader() (io.ReadCloser, error) {
	return http.RequestWithContext(&h.ctx, "GET", h.Url, h.Header, nil)
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
