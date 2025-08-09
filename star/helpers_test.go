package star_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mbark/shmake/star"
	"github.com/stretchr/testify/require"
)

func setupStarlarkRuntime(t *testing.T) {
	dir := t.TempDir()
	err := os.Chdir(dir)
	require.NoError(t, err)
}

func withMainStar(t *testing.T, contents string) {
	dir, err := os.Getwd()
	require.NoError(t, err)

	err = os.RemoveAll(filepath.Join(dir, "main.star"))
	require.NoError(t, err)

	f, err := os.Create(filepath.Join(dir, "main.star"))
	require.NoError(t, err)

	t.Cleanup(func() {
		err := f.Close()
		require.NoError(t, err)
	})

	_, err = f.WriteString(contents)
	require.NoError(t, err)

	t.Log("=== main.star ===")
	for i, line := range strings.Split(contents, "\n") {
		t.Logf("%3d: %s", i+1, line)
	}
	t.Log()
}

func runStarlarkRuntime(t *testing.T, testName string) {
	star.RunStar([]string{testName, "test"})
}
