package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"rafaelmartins.com/p/website/internal/generators"
	"rafaelmartins.com/p/website/internal/http"
	"rafaelmartins.com/p/website/internal/runner"
)

type kicadTaskImpl struct {
	repo                string
	destination         string
	includeNameRevision bool
	immutable           bool

	name     string
	revision string
	url      string
	path     string
}

func (k *kicadTaskImpl) GetDestination() string {
	if k.includeNameRevision {
		return filepath.Join(k.repo, k.destination, k.name, k.revision, k.path)
	}
	return filepath.Join(k.repo, k.destination, k.path)
}

func (k *kicadTaskImpl) GetGenerator() (runner.Generator, error) {
	return &generators.HTTP{
		Url:       k.url,
		Immutable: k.immutable,
	}, nil
}

type Kicad struct {
	Owner    string
	Repo     string
	UrlOrTag string

	BaseDestination     string
	Destination         string
	IncludeNameRevision bool
	Immutable           bool

	ctx   http.RequestContext
	tasks []*runner.Task
}

func (k *Kicad) GetBaseDestination() string {
	// FIXME: deduplicate this with internal/project
	if k.BaseDestination == "" {
		return "projects"
	}
	return k.BaseDestination
}

func (k *Kicad) GetTasks() ([]*runner.Task, error) {
	if k.tasks != nil {
		return k.tasks, nil
	}

	destination := k.Destination
	if destination == "" {
		destination = "kicad"
	}

	baseUrl := k.UrlOrTag
	if !strings.HasPrefix(baseUrl, "http://") && !strings.HasPrefix(baseUrl, "https://") {
		baseUrl = "https://github.com/" + k.Owner + "/" + k.Repo + "/releases/download/" + k.UrlOrTag
	}
	baseUrl = strings.TrimSuffix(baseUrl, "/index.json")

	body, err := http.RequestWithContext(&k.ctx, "GET", baseUrl+"/index.json", nil, nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	data := map[string]any{}
	if err := json.NewDecoder(body).Decode(&data); err != nil {
		return nil, err
	}

	version, ok := data["version"]
	if !ok {
		return nil, errors.New("tasks: kicad: version field not found")
	}

	fversion, ok := version.(float64)
	if !ok {
		return nil, errors.New("tasks: kicad: version is not an integer")
	}
	iversion := int(fversion)

	rv := []*runner.Task{}

	switch iversion {
	case 1:
		name := data["name"].(string)
		revision := data["revision"].(string)

		if d, ok := data["sch-export-pdf"].(string); ok && d != "" {
			sch, err := url.JoinPath(baseUrl, d)
			if err != nil {
				return nil, err
			}
			rv = append(rv, runner.NewTask(k, &kicadTaskImpl{
				repo:                k.Repo,
				destination:         destination,
				includeNameRevision: k.IncludeNameRevision,
				immutable:           k.Immutable,
				name:                name,
				revision:            revision,
				url:                 sch,
				path:                d,
			}))
		}

		if d, ok := data["pcb-ibom"].(string); ok && d != "" {
			ibom, err := url.JoinPath(baseUrl, d)
			if err != nil {
				return nil, err
			}
			rv = append(rv, runner.NewTask(k, &kicadTaskImpl{
				repo:                k.Repo,
				destination:         destination,
				includeNameRevision: k.IncludeNameRevision,
				immutable:           k.Immutable,
				name:                name,
				revision:            revision,
				url:                 ibom,
				path:                d,
			}))
		}

		if d, ok := data["pcb-gerber"].(string); ok && d != "" {
			gerber, err := url.JoinPath(baseUrl, d)
			if err != nil {
				return nil, err
			}
			rv = append(rv, runner.NewTask(k, &kicadTaskImpl{
				repo:                k.Repo,
				destination:         destination,
				includeNameRevision: k.IncludeNameRevision,
				immutable:           k.Immutable,
				name:                name,
				revision:            revision,
				url:                 gerber,
				path:                d,
			}))
		}

		if d, ok := data["pcb-render"].(map[string]any); ok && d != nil {
			for _, v := range d {
				if l, ok := v.([]any); ok {
					for _, m := range l {
						if mm, ok := m.(map[string]any); ok {
							img, err := url.JoinPath(baseUrl, mm["file"].(string))
							if err != nil {
								return nil, err
							}
							rv = append(rv, runner.NewTask(k, &kicadTaskImpl{
								repo:                k.Repo,
								destination:         destination,
								includeNameRevision: k.IncludeNameRevision,
								immutable:           k.Immutable,
								name:                name,
								revision:            revision,
								url:                 img,
								path:                mm["file"].(string),
							}))
						}
					}
				}
			}
		}

	default:
		return nil, fmt.Errorf("project: kicad: unsupported version: %d", iversion)
	}

	k.tasks = rv
	return rv, nil
}
