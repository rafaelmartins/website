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
	if k.pcb == "" {
		return ""
	}

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

func (k *KicadProject) PcbRenderFiles(config *PcbRenderConfig) map[string][]*PcbRenderFile {
	if k.pcb == "" || config == nil {
		return map[string][]*PcbRenderFile{}
	}

	rv := map[string][]*PcbRenderFile{}
	for _, side := range config.Sides {
		if len(config.Scales) == 0 {
			rv[side] = append(rv[side], &PcbRenderFile{
				Scale: 0,
				File:  k.PcbRenderFilename(side, 0),
			})
			continue
		}

		for _, scale := range config.Scales {
			rv[side] = append(rv[side], &PcbRenderFile{
				Scale: scale,
				File:  k.PcbRenderFilename(side, scale),
			})
		}
	}
	return rv
}

func (k *KicadProject) PcbRender(ch chan *runner.GeneratorByProduct, cli *KicadCli, config *PcbRenderConfig) {
	if k.pcb == "" || ch == nil || cli == nil || config == nil {
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
			// this does not works at all as of kicad 9.0.5
			// see https://gitlab.com/kicad/code/kicad/-/issues/21950
			args = append(args,
				"--preset", config.preset,
			)
		} else {
			args = append(args,
				"--no-use-board-stackup-colors",
			)
		}

		if config.Pan != nil {
			args = append(args, "--pan", config.Pan.String())
		}

		if config.Rotate != nil {
			args = append(args, "--rotate", config.Rotate.String())
		}

		if _, err := cli.Run(append(args, k.pcb)...); err != nil {
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

func (k *KicadProject) PcbIbomFilename(config *PcbIbomConfig) string {
	if k.pcb == "" || config == nil || !config.Enable {
		return ""
	}
	return k.name + ".html"
}

func (k *KicadProject) PcbIbom(ch chan *runner.GeneratorByProduct, ibom *InteractiveHtmlBom, config *PcbIbomConfig) {
	if k.pcb == "" || ch == nil || ibom == nil || config == nil || !config.Enable {
		return
	}

	tmpd, err := os.MkdirTemp("", "website")
	if err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
		return
	}
	defer os.RemoveAll(tmpd)

	args := []string{
		"--no-browser",
		"--dest-dir", tmpd,
		"--name-format", "%f",
		"--include-tracks",
		"--include-nets",
	}

	if config.Blacklist != "" {
		args = append(args, "--blacklist", config.Blacklist)
	}

	if _, err := ibom.Run(append(args, k.pcb)...); err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
		return
	}

	fp, err := os.Open(filepath.Join(tmpd, k.PcbIbomFilename(config)))
	if err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
		return
	}

	ch <- &runner.GeneratorByProduct{
		Filename: k.PcbIbomFilename(config),
		Reader:   fp,
	}
}
