package webserver

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
)

var (
	exe string
	ts  time.Time
)

func init() {
	lexe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exe, err = filepath.EvalSymlinks(lexe)
	if err != nil {
		panic(err)
	}

	st, err := os.Stat(exe)
	if err != nil {
		panic(err)
	}
	ts = st.ModTime().UTC()
}

func watchExec(fn func()) {
	for {
		if st, err := os.Stat(exe); err == nil {
			if ts.Compare(st.ModTime().UTC()) < 0 {
				if fn != nil {
					fn()
				}
				break
			}
		}
		time.Sleep(time.Millisecond * 200)
	}
}

func reExec() error {
	time.Sleep(500 * time.Millisecond) // wait a bit for new exec to settle down
	return syscall.Exec(exe, os.Args, os.Environ())
}
