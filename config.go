package sindr

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

func cacheHome() string {
	home := os.Getenv("HOME")
	cacheDir := xdgPath("CACHE_HOME", path.Join(home, ".cache"))
	return path.Join(cacheDir, "sindr")
}

func xdgPath(name, defaultPath string) string {
	dir := os.Getenv(fmt.Sprintf("XDG_%s", name))
	if dir != "" && path.IsAbs(dir) {
		return dir
	}

	return defaultPath
}

func findPathUpdwards(search string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, search)
		if _, err := os.Stat(candidate); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", errors.New("file not found: " + search)
}
