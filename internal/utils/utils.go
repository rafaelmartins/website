package utils

import (
	"os"
	"path/filepath"
)

func Executables() ([]string, error) {
	lexe, err := os.Executable()
	if err != nil {
		return nil, err
	}

	exe, err := filepath.EvalSymlinks(lexe)
	if err != nil {
		return nil, err
	}

	aexe, err := filepath.Abs(exe)
	if err != nil {
		return nil, err
	}
	return []string{aexe}, nil
}
