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

	err := os.MkdirAll(cacheDir, 0o700)
	if err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	_, err = os.Stat(logFile)
	var buf io.WriteCloser
	switch {
	case errors.Is(err, os.ErrNotExist):
		buf, err = os.Create(filepath.Clean(logFile)) // #nosec G304
	case err != nil:
		return nil, fmt.Errorf("creating cache directory: %w", err)
	default:
		buf, err = os.OpenFile(
			filepath.Clean(logFile),
			os.O_APPEND|os.O_WRONLY,
			0o600,
		) // #nosec G304
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
