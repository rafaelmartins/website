package hardware

import (
	"bytes"
	"encoding/json"
	"io"

	"rafaelmartins.com/p/website/internal/hardware/hconfig"
	"rafaelmartins.com/p/website/internal/hardware/kicad"
	"rafaelmartins.com/p/website/internal/hardware/tools"
	"rafaelmartins.com/p/website/internal/meta"
	"rafaelmartins.com/p/website/internal/runner"
)

func GetTasksGroups(c *hconfig.Config) ([]*runner.TaskGroup, error) {
	cli, err := tools.NewKicadCli()
	if err != nil {
		return nil, err
	}

	ibom, err := tools.NewInteractiveHtmlBom()
	if err != nil {
		return nil, err
	}

	rv := []*runner.TaskGroup{}
	for _, proj := range c.Projects {
		p, err := kicad.NewKicadProject(proj.File)
		if err != nil {
			return nil, err
		}

		rv = append(rv, runner.NewTaskGroup(&Task{
			Project:            p,
			Config:             &proj,
			KicadCli:           cli,
			InteractiveHtmlBom: ibom,
			IsSingle:           len(c.Projects) == 1,
		}))
	}
	return rv, nil
}

type Task struct {
	Project            *kicad.KicadProject
	Config             *hconfig.ProjectConfig
	KicadCli           *tools.KicadCli
	InteractiveHtmlBom *tools.InteractiveHtmlBom
	IsSingle           bool
}

func (t *Task) GetBaseDestination() string {
	if t.Config != nil && t.Config.BaseDestination != "" {
		return t.Config.BaseDestination
	}

	if t.Project != nil && !t.IsSingle {
		return t.Project.GetName()
	}

	return ""
}

func (t *Task) GetTasks() ([]*runner.Task, error) {
	return []*runner.Task{runner.NewTask(t, t)}, nil
}

func (t *Task) GetDestination() string {
	return "index.json"
}

func (t *Task) GetGenerator() (runner.Generator, error) {
	return t, nil
}

func (*Task) GetID() string {
	return "KICAD"
}

func (t *Task) GetReader() (io.ReadCloser, error) {
	md, err := meta.GetMetadata()
	if err != nil {
		return nil, err
	}

	data := struct {
		Version      int                               `json:"version"`
		Name         string                            `json:"name"`
		Revision     string                            `json:"revision"`
		PcbRender    map[string][]*kicad.PcbRenderFile `json:"pcb-render"`
		PcbIbom      string                            `json:"pcb-ibom"`
		PcbGerber    string                            `json:"pcb-gerber"`
		SchExportPdf string                            `json:"sch-export-pdf"`
		Tools        map[string]string                 `json:"tools"`
	}{
		Version:      1,
		Name:         t.Project.Name(),
		Revision:     t.Project.Revision(),
		PcbRender:    t.Project.PcbRenderFiles(t.Config.PcbRender),
		PcbIbom:      t.Project.PcbIbomFilename(t.Config.PcbIbom),
		PcbGerber:    t.Project.PcbGerberFilename(t.Config.PcbGerber),
		SchExportPdf: t.Project.SchExportPdfFilename(t.Config.SchExportPdf),
		Tools: map[string]string{
			"Kicad":              t.KicadCli.Version(),
			"InteractiveHtmlBom": t.InteractiveHtmlBom.Version(),
			md.Name:              md.Version,
		},
	}

	preset, err := kicad.Patch3dLayers(t.KicadCli, t.Config.PcbRender.PresetFile, t.Config.PcbRender.IncludeDNP)
	if err != nil {
		return nil, err
	}
	t.Project.SetPreset(preset)

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (t *Task) GetPaths() ([]string, error) {
	return t.Project.GetPaths()
}

func (*Task) GetImmutable() bool {
	return false
}

func (t *Task) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch == nil {
		return
	}

	t.Project.PcbRender(ch, t.KicadCli, t.Config.PcbRender)
	t.Project.PcbIbom(ch, t.InteractiveHtmlBom, t.Config.PcbIbom)
	t.Project.PcbGerber(ch, t.Config.PcbGerber)
	t.Project.SchExportPdf(ch, t.KicadCli, t.Config.SchExportPdf)

	close(ch)
}
