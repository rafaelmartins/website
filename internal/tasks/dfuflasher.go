package tasks

import (
	"path"
	"strings"

	"rafaelmartins.com/p/website/internal/generators"
	"rafaelmartins.com/p/website/internal/opengraph"
	"rafaelmartins.com/p/website/internal/runner"
	"rafaelmartins.com/p/website/internal/templates"
)

type dfuFlasherTask struct {
	baseDestination string
	title           string
	description     string
	projects        []string
	template        string
	layoutCtx       *templates.LayoutContext

	openGraph         *opengraph.Config
	openGraphImageGen *opengraph.OpenGraphImageGen
}

func (t *dfuFlasherTask) GetDestination() string {
	return "index.html"
}

func (t *dfuFlasherTask) GetGenerator() (runner.Generator, error) {
	url := path.Join("/", t.baseDestination)
	if url != "/" {
		url += "/"
	}

	ctx := map[string]any{
		"projects": strings.Join(t.projects, ";"),
	}

	return &generators.Content{
		Title:       t.title,
		Description: t.description,
		URL:         url,
		Template:    t.template,
		TemplateCtx: ctx,
		LayoutCtx:   t.layoutCtx,

		OpenGraph:         t.openGraph,
		OpenGraphImageGen: t.openGraphImageGen,
	}, nil
}

type DfuFlasher struct {
	Title           string
	Description     string
	Projects        []string
	BaseDestination string
	Template        string
	WithSidebar     bool

	OpenGraph         *opengraph.Config
	OpenGraphImageGen *opengraph.OpenGraphImageGen
}

func (d *DfuFlasher) GetBaseDestination() string {
	if d.BaseDestination == "" {
		return "dfu-flasher"
	}
	return d.BaseDestination
}

func (d *DfuFlasher) GetTasks() ([]*runner.Task, error) {
	tmpl := d.Template
	if tmpl == "" {
		tmpl = "dfu-flasher.html"
	}

	return []*runner.Task{
		runner.NewTask(d, &dfuFlasherTask{
			baseDestination: d.BaseDestination,
			title:           d.Title,
			description:     d.Description,
			projects:        d.Projects,
			template:        tmpl,
			layoutCtx: &templates.LayoutContext{
				WithSidebar: d.WithSidebar,
			},

			openGraph:         d.OpenGraph,
			openGraphImageGen: d.OpenGraphImageGen,
		}),
	}, nil
}
