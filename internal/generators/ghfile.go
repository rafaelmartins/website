package generators

import (
	"io"

	"rafaelmartins.com/p/website/internal/github"
	"rafaelmartins.com/p/website/internal/runner"
)

type GithubFile struct {
	Owner     string
	Repo      string
	Ref       string
	Path      string
	Immutable bool
}

func (*GithubFile) GetID() string {
	return "GHFILE"
}

func (g *GithubFile) GetReader() (io.ReadCloser, error) {
	return github.GetRepositoryFile(g.Owner, g.Repo, g.Path, g.Ref)
}

func (*GithubFile) GetPaths() ([]string, error) {
	return nil, nil
}

func (g *GithubFile) GetImmutable() bool {
	return g.Immutable
}

func (*GithubFile) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch != nil {
		close(ch)
	}
}
