package webserver

import (
	"errors"
	"io/fs"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type rwWrapper struct {
	statusCode int
	rw         http.ResponseWriter
}

func (w *rwWrapper) Header() http.Header {
	return w.rw.Header()
}

func (w *rwWrapper) Write(b []byte) (int, error) {
	return w.rw.Write(b)
}

func (w *rwWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.rw.WriteHeader(statusCode)
}

type fsWrapper struct {
	dir http.FileSystem
}

func (f *fsWrapper) Open(name string) (http.File, error) {
	fp, err := f.dir.Open(name)
	if err == nil {
		return fp, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	efp, eerr := f.dir.Open("404.html")
	if eerr != nil {
		efp, eerr = f.dir.Open("404/index.html")
		if eerr != nil {
			return nil, err // original error
		}
	}
	return efp, nil
}

func ListenAndServeWithReloader(addr string, dir string, cb func() error) error {
	exit := make(chan error)
	reload := make(chan bool)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(&fsWrapper{
		dir: http.Dir(dir),
	}))

	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &rwWrapper{rw: w}
			mux.ServeHTTP(rw, r)
			log.Printf("[HTTP] %s - %s %q %s - %d", r.RemoteAddr, r.Method, r.URL, r.Proto, rw.statusCode)
		}),
	}

	go func() {
		log.Printf("Listening on %s ...", addr)
		if err := server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			exit <- err
		}
	}()

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		go watchExec(func() {
			server.Close()
			close(reload)
		})
	}

	return backoff.Retry(func() error {
		if err := func() error {
			for {
				select {
				case <-reload:
					return backoff.Permanent(reExec())
				case err := <-exit:
					return backoff.Permanent(err)
				default:
				}

				if cb != nil {
					if err := cb(); err != nil {
						return err
					}
				}

				time.Sleep(time.Second)
			}
		}(); err != nil {
			log.Printf("error: %s", err)
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
}
