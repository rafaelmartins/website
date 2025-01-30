package tasks

import (
	"errors"
	"path/filepath"

	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/runner"
)

type qrcodeTask struct {
	sourceFile          string
	sourceContent       string
	destinationFilename string
	size                int
	foregroundColor     *uint32
	backgroundColor     *uint32
	withoutBorders      bool
}

func (t *qrcodeTask) GetDestination() string {
	return filepath.FromSlash(t.destinationFilename)
}

func (t qrcodeTask) GetGenerator() (runner.Generator, error) {
	return &generators.QRCode{
		File:            t.sourceFile,
		Content:         t.sourceContent,
		Size:            t.size,
		ForegroundColor: t.foregroundColor,
		BackgroundColor: t.backgroundColor,
		WithoutBorders:  t.withoutBorders,
	}, nil
}

type QRCode struct {
	SourceFile      string
	SourceContent   string
	DestinationFile string
	Size            int
	ForegroundColor *uint32
	BackgroundColor *uint32
	WithoutBorders  bool
}

func (q *QRCode) GetBaseDestination() string {
	return filepath.Dir(q.DestinationFile)
}

func (q *QRCode) GetTasks() ([]*runner.Task, error) {
	if q.DestinationFile == "" {
		return nil, errors.New("qrcode: destination file not defined")
	}

	return []*runner.Task{
		runner.NewTask(q,
			&qrcodeTask{
				sourceFile:          q.SourceFile,
				sourceContent:       q.SourceContent,
				destinationFilename: filepath.Base(q.DestinationFile),
				size:                q.Size,
				foregroundColor:     q.ForegroundColor,
				backgroundColor:     q.BackgroundColor,
				withoutBorders:      q.WithoutBorders,
			},
		),
	}, nil
}
