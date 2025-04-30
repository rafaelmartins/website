package postproc

import (
	"io"
	"os/exec"
)

type JPEG struct{}

func (JPEG) Supported(ext string) bool {
	return ext == ".jpg" || ext == ".jpeg"
}

func (p *JPEG) Run(dstFn string, dst io.Writer, src io.Reader) error {
	cmd := exec.Command("jpegoptim", "-m60", "-")
	cmd.Stdin = src
	cmd.Stdout = dst
	return cmd.Run()
}
