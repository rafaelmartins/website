package runner

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	impl TaskImpl
	gen  Generator
}

func NewTask(impl TaskImpl) *Task {
	return &Task{
		impl: impl,
	}
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

func Run(basedir string, cfg Config, groups []*TaskGroup, force bool) error {
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

		for _, task := range tasks {
			if task.impl == nil {
				continue
			}

			var gen Generator
			if task.gen == nil {
				var err error
				gen, err = task.impl.GetGenerator()
				if err != nil {
					return err
				}
				task.gen = gen
			} else {
				gen = task.gen
			}

			dest := filepath.Join(basedir, group.impl.GetBaseDestination(), task.impl.GetDestination())
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
						continue
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

				dest := filepath.Join(basedir, group.impl.GetBaseDestination(), filepath.Dir(task.impl.GetDestination()), bp.Filename)

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
		}
	}

	return nil
}
