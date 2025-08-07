package shmake_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mbark/shmake"
)

func setupRuntime(t *testing.T) *shmake.Runtime {
	dir := t.TempDir()

	cacheDir := filepath.Join(dir, "cache")
	logFile, err := os.Create(filepath.Join(dir, "logs"))
	require.NoError(t, err)
	t.Cleanup(func() {
		err := logFile.Close()
		require.NoError(t, err)
	})

	err = os.Chdir(dir)
	require.NoError(t, err)

	r, err := shmake.NewRuntime(cacheDir, logFile)
	require.NoError(t, err)

	// set this by default
	r.Args = []string{t.Name(), "test"}
	return r
}

func withMainLua(t *testing.T, contents string) {
	dir, err := os.Getwd()
	require.NoError(t, err)

	err = os.RemoveAll(filepath.Join(dir, "main.lua"))
	require.NoError(t, err)

	f, err := os.Create(filepath.Join(dir, "main.lua"))
	require.NoError(t, err)

	t.Cleanup(func() {
		err := f.Close()
		require.NoError(t, err)
	})

	_, err = f.WriteString(contents)
	require.NoError(t, err)

	t.Log("=== main.lua ===")
	for i, line := range strings.Split(contents, "\n") {
		t.Logf("%3d: %s", i+1, line)
	}
	t.Log()
}
