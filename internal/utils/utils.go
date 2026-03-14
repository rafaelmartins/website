package utils

import (
	"os"
	"path/filepath"
)

var exe = func() string {
	lexe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exe, err := filepath.EvalSymlinks(lexe)
	if err != nil {
		panic(err)
	}

	aexe, err := filepath.Abs(exe)
	if err != nil {
		panic(err)
	}
	return aexe
}()

func Executable() string {
	return exe
}

func Executables() ([]string, error) {
	return []string{exe}, nil
}
