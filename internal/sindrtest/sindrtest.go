package sindrtest

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"

	"github.com/mbark/sindr"
	"github.com/mbark/sindr/internal/logger"
)

type testOptions struct {
	args           []string
	fail           bool
	packageJson    map[string]any
	rawPackageJson string
	logger         logger.Interface
	writer         io.Writer
	envs           map[string]string
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

func WithLogger(logger logger.Interface) TestOption {
	return func(o *testOptions) {
		o.logger = logger
	}
}

func WithWriter(writer io.Writer) TestOption {
	return func(o *testOptions) {
		o.writer = writer
	}
}

func WithEnv(k, v string) TestOption {
	return func(o *testOptions) {
		o.envs[k] = v
	}
}

var fileName = "test.star"

func Test(t *testing.T, contents string, opts ...TestOption) {
	t.Helper()

	options := testOptions{envs: make(map[string]string)}
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

	args := []string{"sindr"}
	if options.args != nil {
		args = append(args, options.args...)
	} else {
		args = append(args, "test")
	}

	var testWriter io.Writer = &CollectWriter{T: t}
	writer := t.Output()
	if options.writer != nil {
		writer = options.writer
		testWriter = options.writer
	}

	var l logger.Interface = testLogger{T: t, writer: testWriter}
	if options.logger != nil {
		l = options.logger
	}

	for k, v := range options.envs {
		t.Setenv(k, v)
	}
	err = sindr.Run(t.Context(),
		args,
		sindr.WithFileName(fileName),
		sindr.WithCacheDir(dir+"/cache"),
		sindr.WithDirectory(dir),
		sindr.WithVerboseLogging(true),
		sindr.WithLogger(l),
		sindr.WithWriter(writer),
		sindr.WithBuiltin("assert_equals", builtinAssertEquals(t, contents)),
	)
	if options.fail {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
	}
}

var _ io.Writer = (*CollectWriter)(nil)

type CollectWriter struct {
	T      *testing.T
	Writes []string
}

func (c *CollectWriter) Write(p []byte) (n int, err error) {
	c.Writes = append(c.Writes, string(p))
	return len(p), nil
}

var _ logger.Interface = testLogger{}

type testLogger struct {
	T      *testing.T
	stack  starlark.CallStack
	writer io.Writer
}

func (t testLogger) Print(message string) {
	t.T.Logf("%s", message)
	_, _ = t.writer.Write([]byte(message))
}

func (t testLogger) WithStack(stack starlark.CallStack) logger.Interface {
	t.stack = stack
	return t
}

func (t testLogger) Log(messages ...string) {
	if len(t.stack) > 0 {
		t.T.Logf("%s %s\n", t.stack[0].Pos.String(), strings.Join(messages, " "))
		_, _ = t.writer.Write(
			[]byte(fmt.Sprintf("%s %s\n", t.stack[0].Pos.String(), strings.Join(messages, " "))),
		)
		return
	}

	t.T.Log(strings.Join(messages, " "))
	_, _ = t.writer.Write([]byte(strings.Join(messages, " ")))
}

func (t testLogger) LogErr(message string, err error) {
	require.NoError(t.T, err, message)
}

func (t testLogger) LogVerbose(messages ...string) {
	t.Log(messages...)
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

func builtinAssertEquals(t *testing.T, contents string) func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var expected, actual starlark.Value
		var message string
		if err := starlark.UnpackArgs("assert_equals", args, kwargs,
			"expected", &expected,
			"actual", &actual,
			"message?", &message,
		); err != nil {
			return nil, err
		}

		at := thread.CallStack().At(1)
		line := strings.Split(contents, "\n")[at.Pos.Line-1]

		assert.Equal(t, expected, actual, "%s\n%s", message, line)
		return starlark.None, nil
	}
}
