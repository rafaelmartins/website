package kicad

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type KicadProject struct {
	name     string
	revision string
	pro      string
	sch      string
	pcb      string

	pcbPreset string
}

var reRevision = regexp.MustCompile(`rev "([0-9.-]+)"`)

func getRevision(f string) (string, error) {
	data, err := os.ReadFile(f)
	if err != nil {
		return "", err
	}

	if m := reRevision.FindSubmatch(data); len(m) == 2 {
		return string(m[1]), nil
	}
	return "", errors.New("kicad: revision not found")
}

func NewKicadProject(pro string) (*KicadProject, error) {
	pro, err := filepath.Abs(pro)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(pro)
	if ext != ".kicad_pro" {
		return nil, fmt.Errorf("kicad: invalid project file: %s", pro)
	}
	if _, err := os.Stat(pro); err != nil {
		return nil, err
	}

	base := strings.TrimSuffix(pro, ext)
	rv := &KicadProject{
		name: filepath.Base(base),
		pro:  pro,
	}

	sch := base + ".kicad_sch"
	if _, err := os.Stat(sch); err == nil {
		rev, err := getRevision(sch)
		if err != nil {
			return nil, err
		}
		rv.revision = rev
		rv.sch = sch
	}

	pcb := base + ".kicad_pcb"
	if _, err := os.Stat(pcb); err == nil {
		rev, err := getRevision(pcb)
		if err != nil {
			return nil, err
		}
		if rv.revision != "" && rv.revision != rev {
			return nil, fmt.Errorf("kicad: revision mismatch: %q != %q", rv.revision, rev)
		}
		rv.revision = rev
		rv.pcb = pcb
	}
	return rv, nil
}

func (k *KicadProject) GetName() string {
	return k.name
}

func (k *KicadProject) SetPreset(preset string) {
	k.pcbPreset = preset
}

func (k *KicadProject) GetPaths() ([]string, error) {
	rv := []string{k.pro}
	if k.sch != "" {
		rv = append(rv, k.sch)
	}
	if k.pcb != "" {
		rv = append(rv, k.pcb)
	}
	return rv, nil
}

func (k *KicadProject) Name() string {
	return k.name
}

func (k *KicadProject) Revision() string {
	return k.revision
}
