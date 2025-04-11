package frontmatter

import (
	"bytes"
	"errors"
	"time"

	"github.com/goccy/go-yaml"
)

type FrontMatterDate struct {
	time.Time
}

func (d *FrontMatterDate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	s := ""
	if err := unmarshal(&s); err != nil {
		return err
	}

	dt, err1 := time.Parse(time.DateTime, s)
	if err1 == nil {
		d.Time = dt
		return nil
	}

	dt, err := time.Parse(time.DateOnly, s)
	if err == nil {
		d.Time = dt
		return nil
	}
	return err1
}

type FrontMatter struct {
	Title       string          `yaml:"title"`
	Description string          `yaml:"description"`
	Date        FrontMatterDate `yaml:"date"`
	Author      struct {
		Name  string `yaml:"name"`
		Email string `yaml:"email"`
	} `yaml:"author"`
	OpenGraph struct {
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
		Image       string `yaml:"image"`
		ImageGen    struct {
			Color *uint32  `yaml:"color"`
			DPI   *float64 `yaml:"dpi"`
			Size  *float64 `yaml:"size"`
		} `yaml:"image-gen"`
	} `yaml:"opengraph"`
	Extra map[string]any `yaml:"extra"`
}

func Parse(src []byte) (*FrontMatter, []byte, error) {
	fm := []byte{}
	rest := []byte{}
	level := 0

	for rawLine := range bytes.Lines(src) {
		line := string(bytes.TrimRight(rawLine, "\r\n"))

		switch level {
		case 0:
			if line != "---" {
				rest = append(rest, rawLine...)
				level = 2
				continue
			}
			fm = append(fm, rawLine...)
			level = 1

		case 1:
			if line == "---" {
				level = 2
				continue
			}
			fm = append(fm, rawLine...)

		case 2:
			rest = append(rest, rawLine...)
		}
	}

	if level == 0 {
		return nil, nil, errors.New("content: frontmatter: file is empty")
	}

	metadata := &FrontMatter{}
	if err := yaml.Unmarshal(fm, metadata); err != nil {
		return nil, nil, err
	}
	return metadata, rest, nil
}
