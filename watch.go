package shmake

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gobwas/glob"
	"github.com/radovskyb/watcher"
)

func startWatching(_ *Runtime, watchGlob string, onChange chan bool) (func(), error) {
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
