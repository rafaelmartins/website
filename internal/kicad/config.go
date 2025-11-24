package kicad

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type XYZ struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
	Z float64 `yaml:"z"`
}

func (x XYZ) String() string {
	return fmt.Sprintf("'%f,%f,%f'", x.X, x.Y, x.Z)
}

type PcbRenderConfig struct {
	Width      int      `yaml:"width"`
	Height     int      `yaml:"height"`
	Zoom       *float64 `yaml:"zoom"`
	Pan        *XYZ     `yaml:"pan"`
	Rotate     *XYZ     `yaml:"rotate"`
	Scales     []int    `yaml:"scales"`
	Sides      []string `yaml:"sides"`
	PresetFile string   `yaml:"preset-file"`
	IncludeDNP bool     `yaml:"include-dnp"`
	preset     string
}

type PcbIbomConfig struct {
	Enable    bool   `yaml:"enable"`
	Blacklist string `yaml:"blacklist"`
}

type PcbGerberConfig struct {
	Enable      bool   `yaml:"enable"`
	CopyPattern string `yaml:"copy-pattern"`
}

type SchExportPdfConfig struct {
	Enable bool `yaml:"enable"`
}

type ProjectConfig struct {
	BaseDestination string              `yaml:"base-destination"`
	File            string              `yaml:"file"`
	PcbRender       *PcbRenderConfig    `yaml:"pcb-render"`
	PcbIbom         *PcbIbomConfig      `yaml:"pcb-ibom"`
	PcbGerber       *PcbGerberConfig    `yaml:"pcb-gerber"`
	SchExportPdf    *SchExportPdfConfig `yaml:"sch-export-pdf"`
}

type Config struct {
	Projects []ProjectConfig `yaml:"projects"`

	file string
	ts   time.Time
}

func (c *Config) GetTimeStamp() (time.Time, error) {
	st, err := os.Stat(c.file)
	if err != nil {
		return time.Time{}, err
	}
	return st.ModTime().UTC(), nil
}

func (c *Config) IsUpToDate() bool {
	ts, err := c.GetTimeStamp()
	return err == nil && ts.Compare(c.ts) <= 0
}

func NewConfig(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	st, err := os.Stat(file)
	if err != nil {
		return nil, err
	}

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	rv := &Config{
		file: file,
		ts:   st.ModTime().UTC(),
	}
	if err := dec.Decode(rv); err != nil {
		return nil, err
	}
	return rv, nil
}
