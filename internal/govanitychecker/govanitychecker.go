package govanitychecker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"

	"golang.org/x/sync/semaphore"
	"rafaelmartins.com/p/website/internal/config"
)

func Run(cfgfile string) error {
	cfg, err := config.New(cfgfile)
	if err != nil {
		return err
	}

	tmp, err := os.MkdirTemp("", "website")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	ctx := context.Background()
	nworkers := int64(runtime.NumCPU())
	sem := semaphore.NewWeighted(nworkers)
	failures := atomic.Int32{}

	for _, proj := range cfg.Projects {
		for _, repo := range proj.Repositories {
			if repo == nil || repo.Go.Import == "" {
				continue
			}

			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}

			base := filepath.Join(tmp, repo.Repo)
			src := filepath.Join(base, "src")
			gopath := filepath.Join(base, "go")

			if err := os.MkdirAll(src, 0777); err != nil {
				return err
			}

			cmd := exec.Command("go", "mod", "init", "govanitytest")
			cmd.Dir = src
			if err := cmd.Run(); err != nil {
				return err
			}

			go func(s string, g string, i string) {
				defer sem.Release(1)

				buf := &bytes.Buffer{}
				cmd := exec.Command("go", "get", "-v", fmt.Sprintf("%s@HEAD", i))
				cmd.Stdout = buf
				cmd.Stderr = buf
				cmd.Dir = s
				cmd.Env = []string{
					"GOPROXY=direct",
					"GOSUMDB=off",
					fmt.Sprintf("GOPATH=%s", g),
					fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
				}
				if err := cmd.Run(); err != nil {
					failures.Add(1)
					log.Printf("error: %s: %s", i, err)
					return
				}
				for line := range strings.Lines(buf.String()) {
					if strings.HasPrefix(line, "get ") && strings.Contains(line, i) {
						fmt.Println(strings.TrimSpace(line))
					}
				}
			}(src, gopath, repo.Go.Import)
		}
	}

	if err := sem.Acquire(ctx, nworkers); err != nil {
		return err
	}

	if f := failures.Load(); f > 0 {
		return fmt.Errorf("%d urls failed", f)
	}
	return nil
}
