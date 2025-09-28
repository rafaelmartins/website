package kicad

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"rafaelmartins.com/p/website/internal/runner"
)

func (k *KicadProject) PcbRenderFilename(side string, scale int) string {
	fn := k.name + "_" + side
	if scale != 0 {
		fn += "_" + strconv.Itoa(scale)
	}
	return fn + ".png"
}

type PcbRenderFile struct {
	Scale int    `json:"scale"`
	File  string `json:"file"`
}

func (k *KicadProject) PcbRenderFiles(config *PcbRenderConfig) []*PcbRenderFile {
	rv := []*PcbRenderFile{}
	for _, side := range config.Sides {
		if len(config.Scales) == 0 {
			rv = append(rv, &PcbRenderFile{
				Scale: 0,
				File:  k.PcbRenderFilename(side, 0),
			})
			continue
		}

		for _, scale := range config.Scales {
			rv = append(rv, &PcbRenderFile{
				Scale: scale,
				File:  k.PcbRenderFilename(side, scale),
			})
		}
	}
	return rv
}

func (k *KicadProject) PcbRender(ch chan *runner.GeneratorByProduct, config *PcbRenderConfig) {
	if ch == nil || config == nil {
		return
	}

	tmpd, err := os.MkdirTemp("", "website")
	if err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
		return
	}
	defer os.RemoveAll(tmpd)

	for _, side := range config.Sides {
		args := []string{
			"pcb", "render",
			"--output", filepath.Join(tmpd, "out.png"),
			"--background", "opaque",
			"--width", fmt.Sprint(config.Width),
			"--height", fmt.Sprint(config.Height),
			"--zoom", fmt.Sprint(config.Zoom),
			"--side", side,
		}

		if config.preset != "" {
			args = append(args, "--preset", config.preset)
		}

		if config.Pan != nil {
			args = append(args, "--pan", config.Pan.String())
		}

		if config.Rotate != nil {
			args = append(args, "--rotate", config.Rotate.String())
		}

		if _, err := k.cli.Run(append(args, k.pcb)...); err != nil {
			ch <- &runner.GeneratorByProduct{Err: err}
			return
		}

		fp, err := os.Open(filepath.Join(tmpd, "out.png"))
		if err != nil {
			ch <- &runner.GeneratorByProduct{Err: err}
			return
		}

		if len(config.Scales) == 0 {
			ch <- &runner.GeneratorByProduct{
				Filename: k.PcbRenderFilename(side, 0),
				Reader:   fp,
			}
			continue
		}
		defer fp.Close()

		for _, scale := range config.Scales {
			buf := &bytes.Buffer{}
			if err := resize(buf, fp, scale); err != nil {
				ch <- &runner.GeneratorByProduct{Err: err}
			} else {
				ch <- &runner.GeneratorByProduct{
					Filename: k.PcbRenderFilename(side, scale),
					Reader:   io.NopCloser(buf),
				}
			}
		}
	}
}
