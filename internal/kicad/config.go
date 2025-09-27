package kicad

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

type XYZ struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
	Z float64 `yaml:"z"`
}

func (x XYZ) String() string {
	return fmt.Sprintf("%f,%f,%f", x.X, x.Y, x.Z)
}

type PcbRenderConfig struct {
	Width  int      `yaml:"width"`
	Height int      `yaml:"height"`
	Zoom   float64  `yaml:"zoom"`
	Pan    *XYZ     `yaml:"pan"`
	Rotate *XYZ     `yaml:"rotate"`
	Scales []int    `yaml:"scales"`
	Sides  []string `yaml:"sides"`
}

type SchExportPdfConfig struct {
	Enable bool `yaml:"enable"`
}

type ProjectConfig struct {
	BaseDestination string              `yaml:"base-destination"`
	File            string              `yaml:"file"`
	PcbRender       *PcbRenderConfig    `yaml:"pcb-render"`
	SchExportPdf    *SchExportPdfConfig `yaml:"sch-export-pdf"`
}

type Config struct {
	Projects []ProjectConfig `yaml:"projects"`
}

func NewConfig(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	rv := &Config{}
	if err := dec.Decode(rv); err != nil {
		return nil, err
	}
	return rv, nil
}
