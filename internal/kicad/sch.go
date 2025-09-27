package kicad

import (
	"os"
	"path/filepath"

	"rafaelmartins.com/p/website/internal/runner"
)

func (k *KicadProject) SchExportPdfFilename(config *SchExportPdfConfig) string {
	if k.sch == "" || config == nil || !config.Enable {
		return ""
	}
	return k.name + ".pdf"
}

func (k *KicadProject) SchExportPdf(ch chan *runner.GeneratorByProduct, cli *KicadCli, config *SchExportPdfConfig) {
	if k.sch == "" || ch == nil || cli == nil || config == nil || !config.Enable {
		return
	}

	tmpd, err := os.MkdirTemp("", "website")
	if err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
		return
	}
	defer os.RemoveAll(tmpd)

	args := []string{
		"sch", "export", "pdf",
		"--output", filepath.Join(tmpd, "out.pdf"),
		k.sch,
	}

	if _, err := cli.Run(args...); err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
		return
	}

	fp, err := os.Open(filepath.Join(tmpd, "out.pdf"))
	if err != nil {
		ch <- &runner.GeneratorByProduct{Err: err}
		return
	}

	ch <- &runner.GeneratorByProduct{
		Filename: k.SchExportPdfFilename(config),
		Reader:   fp,
	}
}
