package main

import (
	"github.com/bmatcuk/doublestar/v2"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

func createWatcher(runtime *Runtime, watchGlob string, onChange chan bool) func() {
	runtime.logger.Debug("creating watcher for file glob", zap.String("glob", watchGlob))
	files, err := doublestar.Glob(watchGlob)
	if err != nil {
		runtime.logger.Fatal("unable to parse file glob", zap.Error(err))
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		runtime.logger.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				runtime.logger.Debug("watcher event", zap.String("event", event.String()))
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Chmod == fsnotify.Chmod {
					onChange <- true
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				runtime.logger.Error("watch error", zap.Error(err))
			}
		}
	}()

	for _, file := range files {
		runtime.logger.Debug("watching", zap.String("file", file))
		err = watcher.Add(file)
		if err != nil {
			runtime.logger.Fatal("failed to add file to watch",
				zap.String("file", file),
				zap.Error(err))
		}
	}

	return func() {
		err := watcher.Close()
		runtime.logger.Error("closing wathcer", zap.Error(err))
	}
}
