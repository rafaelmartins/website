package tasks

import (
	"errors"
	"path/filepath"

	"rafaelmartins.com/p/website/internal/generators"
	"rafaelmartins.com/p/website/internal/runner"
)

type jsonTask struct {
	data                any
	destinationFilename string
}

func (t *jsonTask) GetDestination() string {
	return filepath.FromSlash(t.destinationFilename)
}

func (t *jsonTask) GetGenerator() (runner.Generator, error) {
	return &generators.Json{
		Data: t.data,
	}, nil
}

type Json struct {
	Data            any
	DestinationFile string
}

func (j *Json) GetBaseDestination() string {
	return filepath.Dir(j.DestinationFile)
}

func (j *Json) GetTasks() ([]*runner.Task, error) {
	if j.DestinationFile == "" {
		return nil, errors.New("json: destination file not defined")
	}

	return []*runner.Task{
		runner.NewTask(j,
			&jsonTask{
				data:                j.Data,
				destinationFilename: filepath.Base(j.DestinationFile),
			},
		),
	}, nil
}
