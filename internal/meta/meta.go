package meta

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
)

var baseUrl = "https://github.com/rafaelmartins/website"

type MetadataGo struct {
	Version string
	Cgo     bool
	Arch    string
	OS      string
}

func (m *MetadataGo) String() string {
	rv := ""
	if m.Version != "" {
		rv += m.Version
		if m.OS != "" && m.Arch != "" {
			rv += " " + m.OS + "/" + m.Arch
		}
		if m.Cgo {
			rv += " cgo"
		}
	}
	return rv
}

type Metadata struct {
	Name    string
	Version string
	URL     string

	Git struct {
		Revision string
		URL      string
	}

	Go MetadataGo
}

func (m *Metadata) String() string {
	rv := m.Name + " " + m.Version
	if v := m.Go.String(); v != "" {
		rv += " (" + v + ")"
	}
	return rv
}

func GetMetadata() (*Metadata, error) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, errors.New("version: build info not available")
	}

	isGit := false
	rv := &Metadata{}

	for _, s := range bi.Settings {
		switch s.Key {

		case "vcs":
			if s.Value == "git" {
				isGit = true
			} else {
				return nil, errors.New("meta: vcs is not git")
			}

		case "vcs.revision":
			rv.Git.Revision = s.Value

		case "CGO_ENABLED":
			m, err := strconv.ParseBool(s.Value)
			if err != nil {
				return nil, err
			}
			rv.Go.Cgo = m

		case "GOARCH":
			rv.Go.Arch = s.Value

		case "GOOS":
			rv.Go.OS = s.Value
		}
	}

	if !isGit {
		// assume that we are on a different worktree during development
		rv.Git.Revision = "0000000"
	}

	rv.Name = bi.Path
	rv.Version = bi.Main.Version
	rv.URL = baseUrl
	rv.Git.URL = fmt.Sprintf("%s/commit/%s", rv.URL, rv.Git.Revision)
	rv.Go.Version = bi.GoVersion

	return rv, nil
}
