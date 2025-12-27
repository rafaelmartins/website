package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
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
	GetTimeStamps() ([]time.Time, error)
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

func (t *Task) run(basedir string, cfg Config, force bool) (bool, error) {
	if t.impl == nil {
		return false, errors.New("task missing implementation")
	}
	if t.group == nil {
		return false, errors.New("task group is nil")
	}

	var gen Generator
	if t.gen == nil {
		var err error
		gen, err = t.impl.GetGenerator()
		if err != nil {
			return false, err
		}
		t.gen = gen
	} else {
		gen = t.gen
	}

	dest := t.destination(basedir)
	if !force {
		timestamps, err := gen.GetTimeStamps()
		if err != nil {
			return false, err
		}

		if cfg != nil && !gen.GetImmutable() {
			ctimestamp, err := cfg.GetTimeStamp()
			if err != nil {
				return false, err
			}
			timestamps = append(timestamps, ctimestamp)
		}

		if st, err := os.Stat(dest); err == nil {
			destts := st.ModTime().UTC()
			if !slices.ContainsFunc(timestamps, func(ts time.Time) bool {
				return ts.Compare(destts) > 0
			}) {
				return false, nil
			}
		}
	}

	log.Printf("  %-8s  %s", gen.GetID(), dest)

	rd, err := gen.GetReader()
	if err != nil {
		return true, err
	}

	if err := postproc.PostProc(dest, rd); err != nil {
		return true, err
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
			return true, bp.Err
		}

		bpDest := filepath.Join(bpDir, bp.Filename)

		log.Printf("  %-8s  %s [%s]", gen.GetID(), dest, bpDest)

		if err := postproc.PostProc(bpDest, bp.Reader); err != nil {
			return true, err
		}
	}

	return true, nil
}

type TaskGroupImpl interface {
	GetBaseDestination() string
	GetTasks() ([]*Task, error)
}

type TaskGroup struct {
	impl  TaskGroupImpl
	tasks []*Task
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

			if implf, ok := group.impl.(interface{ GetSkipIfExists() string }); ok && !force {
				if skip := implf.GetSkipIfExists(); skip != "" {
					if _, err := os.Stat(path.Join(basedir, skip)); err == nil {
						continue
					}
				}
			}

			if group.tasks == nil {
				tasks, err := group.impl.GetTasks()
				if err != nil {
					queue <- &taskJob{
						err: err,
					}
					break
				}
				group.tasks = tasks
			}
			for _, task := range group.tasks {
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
			close(queue)
			return job.err
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}

		go func(task *Task) {
			defer sem.Release(1)

			if outd, err := task.run(basedir, cfg, force); err != nil {
				failures.Add(1)
				log.Printf("  %-8s  %s: %s", "[ERROR]", task.destination(basedir), err)
			} else if outd {
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

	if outdated.Load() > 0 {
		log.Printf("--------------------------------------------------------------------------------")
		github.DumpRatelimit()
		log.Printf("--------------------------------------------------------------------------------")
	}

	return nil
}
