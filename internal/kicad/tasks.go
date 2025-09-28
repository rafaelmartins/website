package kicad

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"rafaelmartins.com/p/website/internal/runner"
)

func GetTasksGroups(c *Config) ([]*runner.TaskGroup, error) {
	cli, err := NewKicadCli()
	if err != nil {
		return nil, err
	}

	rv := []*runner.TaskGroup{}
	for _, proj := range c.Projects {
		p, err := NewKicadProject(cli, proj.File)
		if err != nil {
			return nil, err
		}

		rv = append(rv, runner.NewTaskGroup(&Task{
			Project:  p,
			Config:   &proj,
			IsSingle: len(c.Projects) == 1,
		}))
	}
	return rv, nil
}

type Task struct {
	Project  *KicadProject
	Config   *ProjectConfig
	IsSingle bool
}

func (t *Task) GetBaseDestination() string {
	if t.Config != nil && t.Config.BaseDestination != "" {
		return t.Config.BaseDestination
	}

	if t.Project != nil && !t.IsSingle {
		return t.Project.name
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
	data := struct {
		KicadVersion string           `json:"kicad-version"`
		PcbRender    []*PcbRenderFile `json:"pcb-render"`
		SchExportPdf string           `json:"sch-export-pdf"`
	}{
		KicadVersion: t.Project.KicadVersion(),
	}

	if t.Config.PcbRender != nil {
		data.PcbRender = t.Project.PcbRenderFiles(t.Config.PcbRender)
	}
	if t.Config.SchExportPdf != nil && t.Config.SchExportPdf.Enable {
		data.SchExportPdf = t.Project.SchExportPdfFilename(t.Config.SchExportPdf)
	}

	var err error
	t.Config.PcbRender.preset, err = Patch3dPreset(t.Project.cli, t.Config.PcbRender.PresetFile, t.Config.PcbRender.IncludeDNP)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

func (*Task) GetTimeStamps() ([]time.Time, error) {
	return []time.Time{time.Now()}, nil // force rebuild
}

func (*Task) GetImmutable() bool {
	return false
}

func (t *Task) GetByProducts(ch chan *runner.GeneratorByProduct) {
	if ch == nil {
		return
	}

	t.Project.PcbRender(ch, t.Config.PcbRender)
	t.Project.SchExportPdf(ch, t.Config.SchExportPdf)

	close(ch)
}
