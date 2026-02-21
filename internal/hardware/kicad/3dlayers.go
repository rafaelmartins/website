package kicad

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"rafaelmartins.com/p/website/internal/hardware/tools"
)

func Patch3dLayers(cli *tools.KicadCli, presetFile string, includeDnp bool) (string, error) {
	if runtime.GOOS != "linux" || os.Getenv("CI") != "true" {
		return "", nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	version := ""
	if v := cli.Version(); v != "" {
		vv := strings.SplitN(v, ".", 3)
		if len(vv) < 2 {
			return "", fmt.Errorf("kicad: 3d_layers.json: failed to parse version: %s", v)
		}
		version = strings.Join(vv[0:2], ".")
	}

	p := filepath.Join(home, ".config", "kicad", version, "3d_viewer.json")
	if _, err := os.Stat(p); !errors.Is(err, fs.ErrNotExist) {
		return "", errors.New("kicad: 3d_viewer.json: already exists")
	}

	cli.Run(
		"pcb", "render",
		"--output", "dummy.png",
		"/dev/null",
	)

	fp, err := os.Open(p)
	if err != nil {
		return "", err
	}

	v := map[string]any{}
	if err := json.NewDecoder(fp).Decode(&v); err != nil {
		fp.Close()
		return "", err
	}
	fp.Close()

	rv := ""
	if presetFile != "" {
		fpf, err := os.Open(presetFile)
		if err != nil {
			return "", err
		}
		defer fpf.Close()

		fv := map[string]any{}
		if err := json.NewDecoder(fpf).Decode(&fv); err != nil {
			return "", err
		}

		lp, ok := v["layer_presets"].([]any)
		if !ok {
			return "", errors.New("kicad: 3d_layers.json: failed to cast layer_presets field")
		}
		v["layer_presets"] = append(lp, fv)

		presetName, ok := fv["name"].(string)
		if !ok {
			return "", fmt.Errorf("kicad: %s: failed to cast name field", presetFile)
		}
		rv = presetName
	}

	render, ok := v["render"].(map[string]any)
	if !ok {
		return "", errors.New("kicad: 3d_layers.json: failed to cast render field")
	}

	render["show_footprints_dnp"] = includeDnp
	render["show_comments"] = false
	render["show_drawings"] = false
	v["render"] = render

	fp, err = os.Create(p)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	if err := json.NewEncoder(fp).Encode(v); err != nil {
		return "", err
	}
	return rv, nil
}
