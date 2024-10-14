package runner

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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

func (t *Task) Run(basedir string, cfg Config, force bool) error {
	if t.impl == nil {
		return errors.New("task missing implementation")
	}
	if t.group == nil {
		return errors.New("task group is nil")
	}

	var gen Generator
	if t.gen == nil {
		var err error
		gen, err = t.impl.GetGenerator()
		if err != nil {
			return err
		}
		t.gen = gen
	} else {
		gen = t.gen
	}

	dest := t.destination(basedir)
	if !force {
		timestamps, err := gen.GetTimeStamps()
		if err != nil {
			return err
		}

		if cfg != nil && !gen.GetImmutable() {
			ctimestamp, err := cfg.GetTimeStamp()
			if err != nil {
				return err
			}
			timestamps = append(timestamps, ctimestamp)
		}

		if st, err := os.Stat(dest); err == nil {
			destts := st.ModTime().UTC()
			found := false
			for _, ts := range timestamps {
				if ts.Compare(destts) > 0 {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}
	}

	log.Printf("  %-8s  %s", gen.GetID(), dest)

	if err := func() error {
		rd, err := gen.GetReader()
		if err != nil {
			return err
		}
		defer rd.Close()

		if err := os.MkdirAll(filepath.Dir(dest), 0777); err != nil {
			return err
		}

		fp, err := os.Create(dest)
		if err != nil {
			return err
		}
		defer fp.Close()

		_, err = io.Copy(fp, rd)
		return err
	}(); err != nil {
		return err
	}

	ch := make(chan *GeneratorByProduct)
	go gen.GetByProducts(ch)

	for bp := range ch {
		if bp.Err != nil {
			return bp.Err
		}

		dest := filepath.Join(basedir, t.group.GetBaseDestination(), filepath.Dir(t.impl.GetDestination()), bp.Filename)

		log.Printf("  %-8s  %s", strings.Repeat("-", len(gen.GetID())), dest)

		if err := func() error {
			defer bp.Reader.Close()

			if err := os.MkdirAll(filepath.Dir(dest), 0777); err != nil {
				return err
			}

			fp, err := os.Create(dest)
			if err != nil {
				return err
			}
			defer fp.Close()

			_, err = io.Copy(fp, bp.Reader)
			return err
		}(); err != nil {
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

func Run(groups []*TaskGroup, basedir string, cfg Config, force bool) error {
	allTasks := []*Task{}
	for _, group := range groups {
		if group == nil {
			continue
		}

		if group.impl == nil {
			continue
		}

		var tasks []*Task
		if group.tasks == nil {
			var err error
			tasks, err = group.impl.GetTasks()
			if err != nil {
				return err
			}
			group.tasks = tasks
		} else {
			tasks = group.tasks
		}
		allTasks = append(allTasks, tasks...)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(allTasks))

	taskCh := []chan *Task{}
	failures := atomic.Int32{}

	for i := 0; i < runtime.NumCPU(); i++ {
		ch := make(chan *Task)
		taskCh = append(taskCh, ch)

		go func(c chan *Task) {
			for task := range c {
				if err := task.Run(basedir, cfg, force); err != nil {
					failures.Add(1)
					log.Printf("  ERROR     %s", task.destination(basedir))
					log.Printf("  -----     error: %s", err)
				}
				wg.Done()
			}
		}(ch)
	}

	for idx, task := range allTasks {
		taskCh[idx%len(taskCh)] <- task
	}

	wg.Wait()

	if f := failures.Load(); f > 0 {
		return fmt.Errorf("runner: %d tasks failed", f)
	}
	return nil
}
