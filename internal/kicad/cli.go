package kicad

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type KicadCli struct {
	bin     string
	version string
}

func NewKicadCli() (*KicadCli, error) {
	if runtime.GOOS == "darwin" {
		if err := os.Setenv("PATH", "/Applications/KiCad/KiCad.app/Contents/MacOS:/Applications/KiCad.app/Contents/MacOS:"+os.Getenv("PATH")); err != nil {
			return nil, err
		}
	}

	bin, err := exec.LookPath("kicad-cli")
	if err != nil {
		return nil, err
	}

	rv := &KicadCli{
		bin: bin,
	}

	version, err := rv.Run("version")
	if err != nil {
		return nil, fmt.Errorf("kicad-cli: failed to detect version: %w", err)
	}
	rv.version = strings.TrimSpace(version)
	return rv, nil
}

func (k *KicadCli) Run(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	cmd := exec.Command(k.bin, args...)
	cmd.Stderr = buf
	cmd.Stdout = buf
	err := cmd.Run()
	return buf.String(), err
}

func (k *KicadCli) Version() string {
	return k.version
}
