package webserver

import (
	"os"
	"runtime"
	"syscall"
	"time"

	"rafaelmartins.com/p/website/internal/utils"
)

func ReExec() error {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		time.Sleep(500 * time.Millisecond) // wait a bit for new exec to settle down
		return syscall.Exec(utils.Executable(), os.Args, os.Environ())
	}
	return nil
}
