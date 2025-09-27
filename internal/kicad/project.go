package kicad

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type KicadProject struct {
	name string
	pro  string
	sch  string
	pcb  string
}

func NewKicadProject(pro string) (*KicadProject, error) {
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
		rv.sch = sch
	}

	pcb := base + ".kicad_pcb"
	if _, err := os.Stat(pcb); err == nil {
		rv.pcb = pcb
	}
	return rv, nil
}

func (k *KicadProject) GetTimeStamps() ([]time.Time, error) {
	st, err := os.Stat(k.pro)
	if err != nil {
		return nil, err
	}
	rv := []time.Time{st.ModTime().UTC()}

	if k.sch != "" {
		st, err := os.Stat(k.sch)
		if err != nil {
			return nil, err
		}
		rv = append(rv, st.ModTime().UTC())
	}

	if k.pcb != "" {
		st, err := os.Stat(k.pcb)
		if err != nil {
			return nil, err
		}
		rv = append(rv, st.ModTime().UTC())
	}
	return rv, nil
}
