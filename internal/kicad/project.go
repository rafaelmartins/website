package kicad

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type KicadProject struct {
	cli  *KicadCli
	name string
	pro  string
	sch  string
	pcb  string
}

func NewKicadProject(cli *KicadCli, pro string) (*KicadProject, error) {
	ext := filepath.Ext(pro)
	if ext != ".kicad_pro" {
		return nil, fmt.Errorf("kicad: invalid project file: %s", pro)
	}
	if _, err := os.Stat(pro); err != nil {
		return nil, err
	}

	base := strings.TrimSuffix(pro, ext)
	rv := &KicadProject{
		cli:  cli,
		name: filepath.Base(base),
		pro:  pro,
	}

	sch := base + ".kicad_sch"
	if _, err := os.Stat(sch); err == nil {
		rv.sch = sch
	}

	pcb := base + ".kicad_pcb"
	if _, err := os.Stat(pcb); err == nil {
		rv.pcb = pcb
	}
	return rv, nil
}

func (k *KicadProject) KicadVersion() string {
	return k.cli.Version()
}
