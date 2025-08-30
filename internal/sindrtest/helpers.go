package sindrtest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mbark/sindr"
)

type testOptions struct {
	args           []string
	fail           bool
	packageJson    map[string]any
	rawPackageJson string
}

type TestOption func(o *testOptions)

func ShouldFail() TestOption {
	return func(o *testOptions) {
		o.fail = true
	}
}

func WithPackageJson(packageJson map[string]any) TestOption {
	return func(o *testOptions) {
		o.packageJson = packageJson
	}
}

func WithRawPackageJson(packageJson string) TestOption {
	return func(o *testOptions) {
		o.rawPackageJson = packageJson
	}
}

func WithArgs(command ...string) TestOption {
	return func(o *testOptions) {
		o.args = command
	}
}

var fileName = "test.star"

func Test(t *testing.T, contents string, opts ...TestOption) {
	t.Helper()

	var options testOptions
	for _, opt := range opts {
		opt(&options)
	}

	dir := t.TempDir()

	err := os.RemoveAll(filepath.Join(dir, fileName))
	require.NoError(t, err)

	f, err := os.Create(filepath.Join(dir, fileName))
	require.NoError(t, err)
	t.Cleanup(func() {
		err := f.Close()
		require.NoError(t, err)
	})

	contents = strings.ReplaceAll(contents, "\t", "    ")

	_, err = f.WriteString(contents)
	require.NoError(t, err)

	t.Log("Wrote to file", filepath.Join(dir, fileName))
	t.Log("=== main.star ===")
	for i, line := range strings.Split(contents, "\n") {
		t.Logf("%3d: %s", i+1, line)
	}
	t.Log()

	if options.packageJson != nil {
		withPackageJson(t, dir, options.packageJson)
	} else if options.rawPackageJson != "" {
		writePackageJson(t, dir, []byte(options.rawPackageJson))
	}

	args := []string{t.Name()}
	if options.args != nil {
		args = append(args, options.args...)
	}

	err = sindr.Run(t.Context(),
		args,
		sindr.WithFileName(fileName),
		sindr.WithCacheDir(dir+"/cache"),
		sindr.WithDirectory(dir),
		sindr.WithVerboseLogging(true),
	)
	if options.fail {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
	}
}

func withPackageJson(t *testing.T, dir string, data map[string]any) {
	t.Helper()

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	writePackageJson(t, dir, jsonData)
}

func writePackageJson(t *testing.T, dir string, jsonData []byte) {
	t.Helper()

	packageJsonPath := filepath.Join(dir, "package.json")
	err := os.WriteFile(packageJsonPath, jsonData, 0o644)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := os.Remove(packageJsonPath)
		require.NoError(t, err)
	})
}
