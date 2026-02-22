package kicad

import (
	"bytes"
	"errors"
	"fmt"
	"image"
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

	fn := k.name
	if k.revision != "" {
		fn += "_" + k.revision
	}
	if side != "" {
		fn += "_" + side
	}
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
	if k.pcb == "" || config == nil || !config.Enable {
		return map[string][]*PcbRenderFile{}
	}

	rv := map[string][]*PcbRenderFile{}
	for _, side := range append([]string{""}, config.Sides...) {
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
	if k.pcb == "" || ch == nil || cli == nil || config == nil || !config.Enable {
		return
	}

	tmpd, err := os.MkdirTemp("", "website")
	if err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
		return
	}
	defer os.RemoveAll(tmpd)

	zoom := 1.
	if config.Zoom != nil {
		zoom = *config.Zoom
	}

	montageSrc := []image.Image{}

	for _, side := range config.Sides {
		args := []string{
			"pcb", "render",
			"--output", filepath.Join(tmpd, "out.png"),
			"--background", "opaque",
			"--width", fmt.Sprint(config.Width),
			"--height", fmt.Sprint(config.Height),
			"--zoom", fmt.Sprint(zoom),
			"--side", side,
		}

		if config.preset != "" {
			// this does not works at all as of kicad 9.0.5
			// see https://gitlab.com/kicad/code/kicad/-/issues/21950
			args = append(args,
				"--preset", config.preset,
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

		src, _, err := image.Decode(fp)
		if err != nil {
			ch <- &runner.GeneratorByProduct{Err: err}
			return
		}

		montageSrc = append(montageSrc, src)

		for _, scale := range config.Scales {
			buf := &bytes.Buffer{}
			if err := resize(buf, src, scale); err != nil {
				ch <- &runner.GeneratorByProduct{Err: err}
			} else {
				ch <- &runner.GeneratorByProduct{
					Filename: k.PcbRenderFilename(side, scale),
					Reader:   io.NopCloser(buf),
				}
			}
		}
	}

	m, err := montage(montageSrc)
	if err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
		return
	}

	for _, scale := range config.Scales {
		buf := &bytes.Buffer{}
		if err := resize(buf, m, scale); err != nil {
			ch <- &runner.GeneratorByProduct{Err: err}
		} else {
			ch <- &runner.GeneratorByProduct{
				Filename: k.PcbRenderFilename("", scale),
				Reader:   io.NopCloser(buf),
			}
		}
	}
}

func (k *KicadProject) PcbIbomFilename(config *PcbIbomConfig) string {
	if k.pcb == "" || config == nil || !config.Enable {
		return ""
	}

	fn := k.name
	if k.revision != "" {
		fn += "_" + k.revision
	}
	return fn + "_ibom.html"
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

	fp, err := os.Open(filepath.Join(tmpd, k.name+".html"))
	if err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
		return
	}

	ch <- &runner.GeneratorByProduct{
		Filename: k.PcbIbomFilename(config),
		Reader:   fp,
	}
}

func (k *KicadProject) PcbGerberFilename(config *PcbGerberConfig) string {
	if k.pcb == "" || config == nil || !config.Enable {
		return ""
	}

	fn := k.name
	if k.revision != "" {
		fn += "_" + k.revision
	}
	return fn + "_gerber.zip"
}

func (k *KicadProject) PcbGerber(ch chan *runner.GeneratorByProduct, config *PcbGerberConfig) {
	if k.pcb == "" || ch == nil || config == nil || !config.Enable {
		return
	}
	if config.CopyPattern == "" {
		ch <- &runner.GeneratorByProduct{Err: errors.New("kicad: gerber: missing copy pattern")}
		return
	}
	if filepath.IsAbs(config.CopyPattern) {
		ch <- &runner.GeneratorByProduct{Err: errors.New("kicad: gerber: copy pattern must be relative to project folder")}
		return
	}

	m, err := filepath.Glob(filepath.Join(filepath.Dir(k.pcb), config.CopyPattern))
	if err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
		return
	}
	if len(m) == 0 {
		ch <- &runner.GeneratorByProduct{Err: errors.New("kicad: gerber-copy: no gerber file found")}
		return
	}
	if len(m) > 1 {
		ch <- &runner.GeneratorByProduct{Err: errors.New("kicad: gerber-copy: multiple gerber files found")}
		return
	}

	fp, err := os.Open(m[0])
	if err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
		return
	}

	ch <- &runner.GeneratorByProduct{
		Filename: k.PcbGerberFilename(config),
		Reader:   fp,
	}
}
