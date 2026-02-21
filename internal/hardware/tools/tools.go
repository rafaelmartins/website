package tools

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type ToolError struct {
	Err error
	Out string
}

func (t *ToolError) Error() string {
	if t.Err == nil {
		return ""
	}

	rv := t.Err.Error()
	if t.Out != "" {
		rv += "\n\n" + t.Out
	}
	return rv
}

func (t *ToolError) Unwrap() error {
	return t.Err
}

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
	if err := cmd.Run(); err != nil {
		return "", &ToolError{
			Err: err,
			Out: buf.String(),
		}
	}
	return buf.String(), nil
}

func (k *KicadCli) Version() string {
	return k.version
}

type InteractiveHtmlBom struct {
	bin     string
	version string
}

func NewInteractiveHtmlBom() (*InteractiveHtmlBom, error) {
	if runtime.GOOS == "darwin" {
		if err := os.Setenv("PATH", "/Applications/KiCad/KiCad.app/Contents/Frameworks/Python.framework/Versions/Current/bin:/Applications/KiCad.app/Contents/Frameworks/Python.framework/Versions/Current/bin:"+os.Getenv("PATH")); err != nil {
			return nil, err
		}
	}

	bin, err := exec.LookPath("generate_interactive_bom.py")
	if err != nil {
		return nil, err
	}

	rv := &InteractiveHtmlBom{
		bin: bin,
	}

	version, err := rv.Run("--version")
	if err != nil {
		return nil, fmt.Errorf("generate_interactive_bom.py: failed to detect version: %w", err)
	}
	rv.version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	return rv, nil
}

func (k *InteractiveHtmlBom) Run(args ...string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("python3", append([]string{k.bin}, args...)...)
	} else {
		cmd = exec.Command(k.bin, args...)
	}

	buf := &bytes.Buffer{}
	cmd.Stderr = buf
	cmd.Stdout = buf
	cmd.Env = []string{
		"INTERACTIVE_HTML_BOM_NO_DISPLAY=1",
	}
	if err := cmd.Run(); err != nil {
		return "", &ToolError{
			Err: err,
			Out: buf.String(),
		}
	}
	return buf.String(), nil
}

func (k *InteractiveHtmlBom) Version() string {
	return k.version
}
