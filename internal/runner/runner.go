package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"
	"rafaelmartins.com/p/website/internal/github"
	"rafaelmartins.com/p/website/internal/pagefind"
	"rafaelmartins.com/p/website/internal/postproc"
	"rafaelmartins.com/p/website/internal/utils"
	"rafaelmartins.com/p/website/internal/webserver"
)

var mayReload = false

type GeneratorByProduct struct {
	Filename string
	Reader   io.ReadCloser
	Err      error
}

type Generator interface {
	GetID() string
	GetReader() (io.ReadCloser, error)
	GetPaths() ([]string, error)
	GetImmutable() bool
	GetByProducts(chan *GeneratorByProduct)
}

type TaskImpl interface {
	GetDestination() string
	GetGenerator() (Generator, error)
}

type Task struct {
	group TaskGroupImpl
	impl  TaskImpl
	gen   Generator
}

func NewTask(group TaskGroupImpl, impl TaskImpl) *Task {
	return &Task{
		group: group,
		impl:  impl,
	}
}

func (t *Task) destination(basedir string) string {
	return filepath.Join(basedir, t.group.GetBaseDestination(), t.impl.GetDestination())
}

func (t *Task) generator() (Generator, error) {
	if t.impl == nil {
		return nil, errors.New("task missing implementation")
	}
	if t.gen != nil {
		return t.gen, nil
	}

	gen, err := t.impl.GetGenerator()
	if err != nil {
		return nil, err
	}
	t.gen = gen
	return t.gen, nil
}

type tst struct {
	t time.Time
	e bool
}

func (t *Task) outdated(basedir string, cfg Config, force bool) (bool, bool, error) {
	if force {
		return true, false, nil
	}

	gen, err := t.generator()
	if err != nil {
		return false, false, err
	}

	dts := time.Time{}
	if st, err := os.Stat(t.destination(basedir)); err == nil {
		if gen.GetImmutable() {
			return false, false, nil
		}
		dts = st.ModTime().UTC()
	} else {
		return true, false, nil
	}

	ts := []*tst{}
	cts, err := cfg.GetTimeStamp()
	if err != nil {
		return false, false, err
	}
	ts = append(ts, &tst{
		t: cts,
	})

	paths, err := gen.GetPaths()
	if err != nil {
		return false, false, err
	}
	slices.Sort(paths)

	for _, p := range slices.Compact(paths) {
		st, err := os.Stat(p)
		if err != nil {
			return false, false, err
		}

		if st.IsDir() {
			if err := filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {
				ts = append(ts, &tst{
					t: info.ModTime().UTC(),
					e: p == utils.Executable(),
				})
				return nil
			}); err != nil {
				return false, false, err
			}
			continue
		}
		ts = append(ts, &tst{
			t: st.ModTime().UTC(),
			e: p == utils.Executable(),
		})
	}

	outdated := false
	isExe := false
	for _, e := range ts {
		if e.t.After(dts) {
			outdated = true
			if e.e {
				isExe = true
			}
		}
	}
	return outdated, isExe, nil
}

func (t *Task) run(basedir string) error {
	if t.group == nil {
		return errors.New("task group is nil")
	}

	gen, err := t.generator()
	if err != nil {
		return err
	}

	dest := t.destination(basedir)
	log.Printf("  %-8s  %s", gen.GetID(), dest)

	rd, err := gen.GetReader()
	if err != nil {
		return err
	}

	if err := postproc.PostProc(dest, rd); err != nil {
		return err
	}

	tmp := filepath.Base(t.impl.GetDestination())
	tmp = strings.TrimSuffix(tmp, filepath.Ext(tmp))

	relDir := ""
	if tmp != "index" {
		relDir = tmp
	}

	ch := make(chan *GeneratorByProduct)
	go gen.GetByProducts(ch)

	bpDir := filepath.Join(basedir, t.group.GetBaseDestination(), filepath.Dir(t.impl.GetDestination()), relDir)
	for bp := range ch {
		if bp.Err != nil {
			return bp.Err
		}

		bpDest := filepath.Join(bpDir, bp.Filename)

		log.Printf("  %-8s  %s [%s]", gen.GetID(), dest, bpDest)

		if err := postproc.PostProc(bpDest, bp.Reader); err != nil {
			return err
		}
	}

	return nil
}

type TaskGroupImpl interface {
	GetBaseDestination() string
	GetTasks() ([]*Task, error)
}

type TaskGroup struct {
	impl TaskGroupImpl

	m     sync.Mutex
	paths map[string][]string
}

func NewTaskGroup(impl TaskGroupImpl) *TaskGroup {
	return &TaskGroup{
		impl: impl,
	}
}

type Config interface {
	GetTimeStamp() (time.Time, error)
}

type taskJob struct {
	task     *Task
	outdated bool
	err      error
}

func Run(groups []*TaskGroup, basedir string, cfg Config, runserver bool, force bool) error {
	defer func() {
		mayReload = true
	}()

	queue := make(chan *taskJob, 100)

	go func() {
		defer close(queue)

		prequeue := []*taskJob{}
		for _, group := range groups {
			if group == nil || group.impl == nil {
				continue
			}

			if implf, ok := group.impl.(interface{ GetSkipIfExists() *string }); ok && !force {
				if skip := implf.GetSkipIfExists(); skip != nil {
					if _, err := os.Stat(path.Join(basedir, *skip)); err == nil {
						continue
					}
				}
			}

			tasks, err := group.impl.GetTasks()
			if err != nil {
				queue <- &taskJob{
					err: err,
				}
				return
			}

			for _, task := range tasks {
				if mayReload {
					outd, isExe, err := task.outdated(basedir, cfg, force)
					if err != nil {
						queue <- &taskJob{
							err: err,
						}
						return
					}
					if !outd {
						continue
					}

					if isExe && runserver {
						if err := webserver.ReExec(); err != nil {
							queue <- &taskJob{
								err: err,
							}
						}
						return
					}

					prequeue = append(prequeue, &taskJob{
						task:     task,
						outdated: true,
					})
					continue
				}

				queue <- &taskJob{
					task: task,
				}
			}
		}

		for _, task := range prequeue {
			queue <- task
		}
	}()

	ctx := context.Background()
	nworkers := int64(runtime.NumCPU())
	sem := semaphore.NewWeighted(nworkers)
	failures := atomic.Int32{}
	outdated := atomic.Int32{}

	for job := range queue {
		if job.err != nil {
			return job.err
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}

		go func(task *Task) {
			defer sem.Release(1)

			if !job.outdated {
				outd, _, err := task.outdated(basedir, cfg, force)
				if err != nil {
					failures.Add(1)
					log.Printf("  %-8s  %s: %s", "[ERROR]", task.destination(basedir), err)
					return
				}
				if !outd {
					return
				}
			}

			if err := task.run(basedir); err != nil {
				failures.Add(1)
				log.Printf("  %-8s  %s: %s", "[ERROR]", task.destination(basedir), err)
			} else {
				outdated.Add(1)
			}
		}(job.task)
	}

	if err := sem.Acquire(ctx, nworkers); err != nil {
		return err
	}

	if f := failures.Load(); f > 0 {
		return fmt.Errorf("runner: %d tasks failed", f)
	}

	od := outdated.Load()

	if pagefind.Outdated(basedir, od) {
		log.Printf("  %-8s  %s", pagefind.GetID(), pagefind.GetDestination(basedir))

		if err := pagefind.GenerateIndex(basedir); err != nil {
			return err
		}
		od++
	}

	if od > 0 {
		log.Printf("--------------------------------------------------------------------------------")
		if github.DumpRatelimit() {
			log.Printf("--------------------------------------------------------------------------------")
		}
	}
	return nil
}
