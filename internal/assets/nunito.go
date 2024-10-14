package assets

import (
	"embed"

	"github.com/rafaelmartins/website/internal/generators"
	"github.com/rafaelmartins/website/internal/runner"
)

//go:embed nunito/nunito.css nunito/*.ttf nunito/*.woff2
var nunitoFS embed.FS

type nunitoFontTask string

func (t nunitoFontTask) GetDestination() string {
	return string(t)
}

func (t nunitoFontTask) GetGenerator() (runner.Generator, error) {
	return &generators.Embed{
		FS:   &nunitoFS,
		Name: string(t),
	}, nil
}

type NunitoFont struct {
	BaseDestination string
}

func (f *NunitoFont) GetBaseDestination() string {
	return f.BaseDestination
}

func (f *NunitoFont) GetTasks() ([]*runner.Task, error) {
	files, err := generators.EmbedListFiles(&nunitoFS)
	if err != nil {
		return nil, err
	}

	rv := []*runner.Task{}
	for _, fl := range files {
		rv = append(rv, runner.NewTask(f, nunitoFontTask(fl)))
	}
	return rv, nil
}
