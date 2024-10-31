package meta

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type Metadata struct {
	Name    string
	Version string
	URL     string

	Git struct {
		Revision string
		Date     time.Time
		Dirty    bool
		URL      string
	}

	Go struct {
		Version string
		Cgo     bool
		Arch    string
		OS      string
	}
}

func (m *Metadata) String() string {
	rv := m.Name + " " + m.Version
	if m.Go.Version != "" {
		rv += " (" + m.Go.Version
		if m.Go.OS != "" && m.Go.Arch != "" {
			rv += " " + m.Go.OS + "/" + m.Go.Arch
		}
		rv += ")"
	}
	return rv
}

func GetMetadata() (*Metadata, error) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, errors.New("version: build info not available")
	}

	rv := &Metadata{}

	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			rv.Git.Revision = s.Value

		case "vcs.time":
			t, err := time.Parse(time.RFC3339, s.Value)
			if err != nil {
				return nil, err
			}
			rv.Git.Date = t

		case "vcs.modified":
			m, err := strconv.ParseBool(s.Value)
			if err != nil {
				return nil, err
			}
			rv.Git.Dirty = m

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

	rv.Name = bi.Path
	rv.Version = "unknown"
	rv.Go.Version = bi.GoVersion

	pathp := strings.Split(rv.Name, "/")
	if len(rv.Git.Revision) < 7 || len(pathp) < 3 || pathp[0] != "github.com" {
		return rv, nil
	}

	rv.URL = fmt.Sprintf("https://%s/%s/%s", pathp[0], pathp[1], pathp[2])
	rv.Git.URL = fmt.Sprintf("%s/commit/%s", rv.URL, rv.Git.Revision)
	rv.Version = rv.Git.Date.Format("2006010215") + "-" + rv.Git.Revision[:7]
	if rv.Git.Dirty {
		rv.Version += "-dirty"
	}
	return rv, nil
}
