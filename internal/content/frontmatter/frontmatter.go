package frontmatter

import (
	"bufio"
	"bytes"
	"errors"
	"time"

	"gopkg.in/yaml.v3"
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

func appendLine(buf []byte, scanner *bufio.Scanner) []byte {
	// this changes the line ending to `\n`, if the file had windows/mac line endings,
	// which is not something bad at all :)
	return append(buf, append(scanner.Bytes(), '\n')...)
}

func Parse(src []byte) (*FrontMatter, []byte, error) {
	fm := []byte{}
	rest := []byte{}
	level := 0

	s := bufio.NewScanner(bytes.NewBuffer(src))
	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, nil, err
		}

		switch level {
		case 0:
			if s.Text() != "---" {
				rest = appendLine(rest, s)
				level = 2
				continue
			}
			fm = appendLine(fm, s)
			level = 1

		case 1:
			if s.Text() == "---" {
				level = 2
				continue
			}
			fm = appendLine(fm, s)

		case 2:
			rest = appendLine(rest, s)
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
