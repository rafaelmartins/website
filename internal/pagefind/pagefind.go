package pagefind

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	en        bool
	assetsdir string
)

type PageFindError struct {
	Err error
	Out string
}

func (p *PageFindError) Error() string {
	if p.Err == nil {
		return ""
	}

	rv := p.Err.Error()
	if p.Out != "" {
		rv += "\n\n" + p.Out
	}
	return rv
}

func (p *PageFindError) Unwrap() error {
	return p.Err
}

func SetGlobals(enable bool, assetsDir string) {
	en = enable
	assetsdir = assetsDir
}

func Outdated(outputDir string, outdated int32) bool {
	if !en {
		return false
	}

	if outdated == 0 {
		if _, err := os.Stat(filepath.Join(outputDir, assetsdir, "pagefind")); err == nil {
			return false
		}
	}
	return true
}

func GetID() string {
	if !en {
		return ""
	}
	return "SEARCH"
}

func GetDestination(outputDir string) string {
	if !en {
		return ""
	}
	return filepath.Join(outputDir, assetsdir, "pagefind", "**")
}

func GenerateIndex(outputDir string) error {
	if !en {
		return nil
	}

	output := filepath.Join(outputDir, assetsdir, "pagefind")

	buf := &bytes.Buffer{}
	cmd := exec.Command("pagefind", "--site", outputDir, "--output-path", output)
	cmd.Stdout = buf
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		return &PageFindError{
			Err: err,
			Out: buf.String(),
		}
	}

	for _, f := range []string{
		"pagefind-highlight.js",
		"pagefind-modular-ui.css",
		"pagefind-modular-ui.js",
		"pagefind-ui.css",
		"pagefind-ui.js",
	} {
		if err := os.Remove(filepath.Join(output, f)); err != nil {
			return err
		}
	}
	return nil
}
