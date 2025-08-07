package shmake

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

func cacheHome() string {
	home := os.Getenv("HOME")
	cacheDir := xdgPath("CACHE_HOME", path.Join(home, ".cache"))
	return path.Join(cacheDir, "shmake")
}

func logPath(cacheDir string) string {
	return path.Join(cacheDir, "shmake.log")
}

func getLogFile() (io.WriteCloser, error) {
	cacheDir := cacheHome()
	logFile := logPath(cacheDir)

	err := os.MkdirAll(cacheDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	_, err = os.Stat(logFile)
	var buf io.WriteCloser
	if errors.Is(err, os.ErrNotExist) {
		buf, err = os.Create(logFile)
	} else if err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	} else {
		buf, err = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0600)
	}
	return buf, err
}

func xdgPath(name, defaultPath string) string {
	dir := os.Getenv(fmt.Sprintf("XDG_%s", name))
	if dir != "" && path.IsAbs(dir) {
		return dir
	}

	return defaultPath
}

func findPathUpdwards(search string) (string, error) {
	dir := "."

	for {
		// If we hit the root, we're done
		if rel, _ := filepath.Rel("/", search); rel == "." {
			break
		}

		_, err := os.Stat(filepath.Join(dir, search))
		if err != nil {
			if os.IsNotExist(err) {
				dir += "/.."
				continue
			}

			return "", err
		}

		return filepath.Abs(dir)
	}

	return "", errors.New("path not found")
}
