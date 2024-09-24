package utils

import (
	"os"
	"path/filepath"
	"time"
)

func ExecutableTimestamp() (time.Time, error) {
	lexe, err := os.Executable()
	if err != nil {
		return time.Time{}, err
	}

	exe, err := filepath.EvalSymlinks(lexe)
	if err != nil {
		return time.Time{}, err
	}

	st, err := os.Stat(exe)
	if err != nil {
		return time.Time{}, err
	}
	return st.ModTime().UTC(), nil
}

func ExecutableTimestamps() ([]time.Time, error) {
	ts, err := ExecutableTimestamp()
	if err != nil {
		return nil, err
	}
	return []time.Time{ts}, nil
}
