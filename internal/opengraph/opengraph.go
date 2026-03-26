package opengraph

import (
	"image/color"
	"os"
	"path"

	"rafaelmartins.com/p/website/internal/hexcolor"
	"rafaelmartins.com/p/website/internal/runner"
)

type Config struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Image       string `yaml:"image"`
	ImageGen    struct {
		Generate *bool    `yaml:"generate"`
		Color    *string  `yaml:"color"`
		DPI      *float64 `yaml:"dpi"`
		Size     *float64 `yaml:"size"`
	} `yaml:"image-gen"`
}

type OpenGraph struct {
	ogi          *OpenGraphImageGen
	pregenerated bool
	baseurl      string
	title        string
	description  string
	image        string
	generate     bool
	c            color.Color
	dpi          float64
	size         float64
}

func New(ogi *OpenGraphImageGen, pregenerated bool, baseurl string, t string, d string, config *Config, ft string, fd string, metadata *Config) (*OpenGraph, error) {
	rv := &OpenGraph{
		ogi:          ogi,
		pregenerated: pregenerated,
		baseurl:      baseurl,
		title:        t,
		description:  d,

		generate: true,
	}

	if ogi != nil {
		rv.c = ogi.dcolor
		rv.dpi = ogi.ddpi
		rv.size = ogi.dsize
	}

	if err := rv.apply(config); err != nil {
		return nil, err
	}

	if ft != "" {
		rv.title = ft
	}
	if fd != "" {
		rv.description = fd
	}
	if err := rv.apply(metadata); err != nil {
		return nil, err
	}

	if pregenerated {
		rv.generate = false
	}
	return rv, nil
}

func (og *OpenGraph) apply(cfg *Config) error {
	if cfg == nil {
		return nil
	}

	if cfg.Title != "" {
		og.title = cfg.Title
	}
	if cfg.Description != "" {
		og.description = cfg.Description
	}
	if cfg.Image != "" {
		og.image = cfg.Image
	}
	if cfg.ImageGen.Generate != nil {
		og.generate = *cfg.ImageGen.Generate
	}
	if cfg.ImageGen.Color != nil {
		var err error
		og.c, err = hexcolor.ToRGBA(*cfg.ImageGen.Color)
		if err != nil {
			return err
		}
	}
	if cfg.ImageGen.DPI != nil {
		og.dpi = *cfg.ImageGen.DPI
	}
	if cfg.ImageGen.Size != nil {
		og.size = *cfg.ImageGen.Size
	}
	return nil
}

func (og *OpenGraph) GenerateByProduct(ch chan *runner.GeneratorByProduct, basedir string) {
	if ch == nil || og.ogi == nil || !(og.ogi.template != "" && og.generate) || og.pregenerated {
		return
	}

	if og.image != "" {
		img := og.image
		if !path.IsAbs(img) {
			img = path.Join(basedir, img)
		}

		rd, err := os.Open(img)
		if err != nil {
			ch <- &runner.GeneratorByProduct{Err: err}
		} else {
			ch <- &runner.GeneratorByProduct{
				Filename: "opengraph.png",
				Reader:   rd,
			}
		}
		return
	}

	rd, err := og.ogi.Generate(og)
	if err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
	} else {
		ch <- &runner.GeneratorByProduct{
			Filename: "opengraph.png",
			Reader:   rd,
		}
	}
}
