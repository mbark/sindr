package shmake

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/radovskyb/watcher"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func shmakeRunAsync(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if args.Len() != 1 {
		return nil, errors.New("run_async() requires exactly 1 argument (a function)")
	}

	callable, ok := args.Index(0).(*starlark.Function)
	if !ok {
		return nil, errors.New("run_async() argument must be a callable function")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		newThread := &starlark.Thread{Name: "async"}
		_, err := starlark.Call(newThread, callable, starlark.Tuple{}, nil)
		if err != nil {
			slog.Error("async function failed", "error", err)
		}
	}()

	return starlark.None, nil
}

func shmakeWait(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	wg.Wait()
	return starlark.None, nil
}

func shmakeWatch(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if args.Len() != 2 {
		return nil, errors.New("watch() requires exactly 2 arguments (glob pattern and function)")
	}

	globVal, ok := args.Index(0).(starlark.String)
	if !ok {
		return nil, errors.New("watch() first argument must be a string (glob pattern)")
	}
	glob := string(globVal)

	callable, ok := args.Index(1).(*starlark.Function)
	if !ok {
		return nil, errors.New("watch() second argument must be a callable function")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		log := slog.With(slog.String("glob", glob))

		onChange := make(chan bool)
		close, err := startWatching(glob, onChange)
		defer close()
		if err != nil {
			log.With(slog.Any("err", err)).Error("starting watcher failed")
			return
		}

		for {
			newThread := &starlark.Thread{Name: "watch"}
			go func() {
				_, err := starlark.Call(newThread, callable, starlark.Tuple{}, nil)
				var serr *starlark.EvalError
				if errors.As(err, &serr) {
					slog.Debug("watch function canceled")
				} else if err != nil {
					slog.Error("watch function failed", "error", err)
				}
			}()

			slog.Debug("waiting for changes")
			<-onChange
			newThread.Cancel("watched file changed")
			log.Info("changes detected, rerunning function")
		}
	}()

	return starlark.None, nil
}

func shmakePool(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	pool := &Pool{wg: sync.WaitGroup{}}

	poolMethods := starlark.StringDict{
		"run":  starlark.NewBuiltin("pool.run", makePoolRun(pool)),
		"wait": starlark.NewBuiltin("pool.wait", makePoolWait(pool)),
	}

	return starlarkstruct.FromStringDict(starlark.String("pool"), poolMethods), nil
}

func makePoolRun(pool *Pool) func(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return func(
		thread *starlark.Thread,
		fn *starlark.Builtin,
		args starlark.Tuple,
		kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		if args.Len() != 1 {
			return nil, errors.New("pool.run() requires exactly 1 argument (a function)")
		}

		callable, ok := args.Index(0).(*starlark.Function)
		if !ok {
			return nil, errors.New("pool.run() argument must be a callable function")
		}

		pool.wg.Add(1)
		go func() {
			defer pool.wg.Done()

			newThread := &starlark.Thread{Name: "pool"}
			_, err := starlark.Call(newThread, callable, starlark.Tuple{}, nil)
			if err != nil {
				slog.Error("pool function failed", "error", err)
			}
		}()

		return starlark.None, nil
	}
}

func makePoolWait(pool *Pool) func(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return func(
		thread *starlark.Thread,
		fn *starlark.Builtin,
		args starlark.Tuple,
		kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		pool.wg.Wait()
		return starlark.None, nil
	}
}

type Pool struct {
	wg sync.WaitGroup
}

func runWatchFnOnce(callable *starlark.Function, onChange chan bool) {
}

func startWatching(watchGlob string, onChange chan bool) (func(), error) {
	w := watcher.New()
	w.SetMaxEvents(1)

	cwd, err := os.Getwd()
	if err != nil {
		return w.Close, err
	}

	g := glob.MustCompile(watchGlob)

	w.AddFilterHook(func(info os.FileInfo, fullPath string) error {
		relPath, err := filepath.Rel(cwd, fullPath)
		if err != nil {
			return err
		}

		// If it doesn't match, try it with the relative path appended
		if g.Match(relPath) || g.Match("./"+relPath) {
			return nil
		} else {
			return watcher.ErrSkip
		}
	})

	log := slog.With(slog.String("glob", watchGlob))

	go func() {
		for {
			select {
			case event := <-w.Event:
				log.With(slog.String("event", event.String())).Debug("watcher event")
				onChange <- true
			case err := <-w.Error:
				log.With(slog.Any("err", err)).Error("watcher failed")
			case <-w.Closed:
				return
			}
		}
	}()
	if err := w.AddRecursive("."); err != nil {
		return w.Close, err
	}

	var paths []string
	for path := range w.WatchedFiles() {
		paths = append(paths, path)
	}

	log.With(slog.Any("files", paths)).Debug("watching files")

	go func() {
		if err := w.Start(time.Millisecond * 100); err != nil {
			log.With(slog.Any("err", err)).Error("starting watcher failed")
		}
	}()

	return w.Close, nil
}
