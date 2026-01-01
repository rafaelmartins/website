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
	"rafaelmartins.com/p/website/internal/postproc"
)

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

func (t *Task) outdated(basedir string, cfg Config, force bool) (bool, error) {
	if force {
		return true, nil
	}

	gen, err := t.generator()
	if err != nil {
		return false, err
	}

	dts := time.Time{}
	if st, err := os.Stat(t.destination(basedir)); err == nil {
		if gen.GetImmutable() {
			return false, nil
		}
		dts = st.ModTime().UTC()
	} else {
		return true, nil
	}

	ts := []time.Time{}
	cts, err := cfg.GetTimeStamp()
	if err != nil {
		return false, err
	}
	ts = append(ts, cts)

	paths, err := gen.GetPaths()
	if err != nil {
		return false, err
	}
	slices.Sort(paths)
	cpaths := slices.Compact(paths)

	for _, p := range cpaths {
		st, err := os.Stat(p)
		if err != nil {
			return false, err
		}

		if st.IsDir() {
			if err := filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {
				ts = append(ts, info.ModTime().UTC())
				return nil
			}); err != nil {
				return false, err
			}
			continue
		}
		ts = append(ts, st.ModTime().UTC())
	}

	for _, e := range ts {
		if e.After(dts) {
			return true, nil
		}
	}
	return false, nil
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
	task *Task
	err  error
}

func Run(groups []*TaskGroup, basedir string, cfg Config, force bool) error {
	queue := make(chan *taskJob, 100)

	go func() {
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
				break
			}

			for _, task := range tasks {
				queue <- &taskJob{
					task: task,
				}
			}
		}
		close(queue)
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

			outd, err := task.outdated(basedir, cfg, force)
			if err != nil {
				queue <- &taskJob{
					err: err,
				}
				return
			}

			if outd {
				if err := task.run(basedir); err != nil {
					failures.Add(1)
					log.Printf("  %-8s  %s: %s", "[ERROR]", task.destination(basedir), err)
				} else {
					outdated.Add(1)
				}
			}
		}(job.task)
	}

	if err := sem.Acquire(ctx, nworkers); err != nil {
		return err
	}

	if f := failures.Load(); f > 0 {
		return fmt.Errorf("runner: %d tasks failed", f)
	}

	if outdated.Load() > 0 {
		log.Printf("--------------------------------------------------------------------------------")
		github.DumpRatelimit()
		log.Printf("--------------------------------------------------------------------------------")
	}
	return nil
}
