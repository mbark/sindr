package main

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/gobwas/glob"
	"github.com/radovskyb/watcher"
)

func startWatching(runtime *Runtime, watchGlob string, onChange chan bool) func() {
	w := watcher.New()
	w.SetMaxEvents(1)

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	g := glob.MustCompile(watchGlob)

	w.AddFilterHook(func(info os.FileInfo, fullPath string) error {
		relPath, err := filepath.Rel(cwd, fullPath)
		if err != nil {
			return err
		}

		// If it doesn't match try it with the relative path appended
		if g.Match(relPath) || g.Match("./"+relPath) {
			return nil
		} else {
			return watcher.ErrSkip
		}
	})

	go func() {
		for {
			select {
			case event := <-w.Event:
				runtime.logger.Debug("watcher event", zap.String("event", event.String()))
				onChange <- true
			case err := <-w.Error:
				panic(err)
			case <-w.Closed:
				return
			}
		}
	}()
	if err := w.AddRecursive("."); err != nil {
		panic(err)
	}

	var paths []string
	for path := range w.WatchedFiles() {
		paths = append(paths, path)
	}

	runtime.logger.Debug("watching files", zap.Strings("files", paths))

	go func() {
		if err := w.Start(time.Millisecond * 100); err != nil {
			panic(err)
		}
	}()

	return w.Close
}
