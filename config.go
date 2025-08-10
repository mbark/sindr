package shmake

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
	return path.Join(cacheDir, "shmake")
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
