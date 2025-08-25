package cache_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mbark/shmake"
)

func setupStarlarkRuntime(t *testing.T) func() {
	t.Helper()
	dir := t.TempDir()
	err := os.Chdir(dir)
	require.NoError(t, err)

	return func() {
		t.Helper()
		err := shmake.Run(t.Context(), []string{t.Name(), "test"}, shmake.WithCacheDir(dir))
		require.NoError(t, err)
	}
}

func withMainStar(t *testing.T, contents string) {
	t.Helper()
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

func TestDiff(t *testing.T) {
	t.Run("with diff expected", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	c = cache()
	if not c.diff(name='version', version='1'):
		fail('unexpected diff')

cli(name="TestDiff", usage="Test diff functionality")
command(name="test", action=test_action)
`)
		run()
	})

	t.Run("with no diff expected", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	c = cache()
	c.set_version(name='version', version='1')
	if c.diff(name='version', version='1'):
		fail('expected no diff')

cli(name="TestDiff", usage="Test diff functionality")
command(name="test", action=test_action)
`)
		run()
	})
}

func TestStore(t *testing.T) {
	t.Run("set_version version successfully", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	c = cache()
	c.set_version(name='test-key', version='v1.0.0')
	
	# Verify it was stored by checking with get_version
	stored = c.get_version('test-key')
	if stored != 'v1.0.0':
		fail('expected stored version to be v1.0.0, got: ' + str(stored))

cli(name="TestStore", usage="Test set_version functionality")
command(name="test", action=test_action)
`)
		run()
	})

	t.Run("set_version with int version", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	c = cache()
	c.set_version(name='test-int', version=42)
	
	# Verify it was stored by checking with get_version
	stored = c.get_version('test-int')
	if stored != '42':
		fail('expected stored version to be 42, got: ' + str(stored))

cli(name="TestStore", usage="Test set_version functionality")
command(name="test", action=test_action)
`)
		run()
	})
}

func TestWithVersion(t *testing.T) {
	t.Run("runs function when version differs", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	c = cache()
	def version_func():
		print('executing')
		return "executed"
	
	ran = c.with_version(version_func, name='test-version', version='v2.0.0')
	
	if not ran:
		fail('expected with_version to return true when function runs')
	
	print('Function executed successfully')

cli(name="TestWithVersion", usage="Test with_version functionality")
command(name="test", action=test_action)
`)
		run()
	})

	t.Run("skips function when version matches", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	c = cache()
	def version_func():
		fail('function should not be called when versions match')
	
	# First set_version a version
	c.set_version(name='skip-test', version='v1.5.0')
	
	# Then try to run with_version with same version
	ran = c.with_version(version_func, name='skip-test', version='v1.5.0')
	
	if ran:
		fail('expected with_version to return false when versions match')
	
	print('Version matching test passed - function was correctly skipped')
	
cli(name="TestWithVersion", usage="Test with_version functionality")
command(name="test", action=test_action)
`)
		run()
	})

	t.Run("runs function with int version", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	c = cache()
	def version_func():
		print('executing int version test')
		return True

	ran = c.with_version(version_func, name='int-version', version=123)
	
	if not ran:
		fail('expected with_version to return true when function runs')
	
	print('Int version function executed successfully')
	
	# Verify version was stored
	stored = c.get_version('int-version')
	if stored != '123':
		fail('expected stored version to be 123, got: ' + str(stored))

cli(name="TestWithVersion", usage="Test with_version functionality")
command(name="test", action=test_action)
`)
		run()
	})
}

func TestCacheWithCustomDir(t *testing.T) {
	t.Run("create cache with custom directory", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
	c = cache(cache_dir="/tmp/test-cache")
	c.set_version(name='custom-dir-test', version='v1.0.0')
	
	# Verify it was stored
	stored = c.get_version('custom-dir-test')
	if stored != 'v1.0.0':
		fail('expected stored version to be v1.0.0, got: ' + str(stored))

cli(name="TestCacheWithCustomDir", usage="Test cache with custom directory")
command(name="test", action=test_action)
`)
		run()
	})
}
