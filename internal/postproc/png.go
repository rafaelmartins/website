package postproc

import (
	"io"
	"os/exec"
)

type PNG struct{}

func (PNG) Supported(ext string) bool {
	return ext == ".png"
}

func (p *PNG) Run(dstFn string, dst io.Writer, src io.Reader) error {
	cmd := exec.Command("pngnq")
	cmd.Stdin = src
	cmd.Stdout = dst
	return cmd.Run()
}
