package internal_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFileTimestamps(t *testing.T) {
	t.Run("newest_ts with single glob", func(t *testing.T) {
		run := setupStarlarkRuntime(t)

		// Create test files with different modification times
		file1 := "test1.txt"
		file2 := "test2.txt"

		require.NoError(t, os.WriteFile(file1, []byte("content1"), 0o644))
		time.Sleep(10 * time.Millisecond)
		require.NoError(t, os.WriteFile(file2, []byte("content2"), 0o644))

		// Get expected newest timestamp
		info2, err := os.Stat(file2)
		require.NoError(t, err)
		expectedNewest := info2.ModTime().Unix()

		withMainStar(t, `
def test_action(ctx):
    result = shmake.newest_ts('test*.txt')
    expected = `+strconv.FormatInt(expectedNewest, 10)+`
    if result != expected:
        fail('expected ' + str(expected) + ', got: ' + str(result))

shmake.cli(name="TestNewestTS")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("oldest_ts with single glob", func(t *testing.T) {
		run := setupStarlarkRuntime(t)

		// Create test files with different modification times
		file1 := "old1.txt"
		file2 := "old2.txt"

		require.NoError(t, os.WriteFile(file1, []byte("content1"), 0o644))
		info1, err := os.Stat(file1)
		require.NoError(t, err)
		expectedOldest := info1.ModTime().Unix()

		time.Sleep(10 * time.Millisecond)
		require.NoError(t, os.WriteFile(file2, []byte("content2"), 0o644))

		withMainStar(t, `
def test_action(ctx):
    result = shmake.oldest_ts('old*.txt')
    expected = `+strconv.FormatInt(expectedOldest, 10)+`
    if result != expected:
        fail('expected ' + str(expected) + ', got: ' + str(result))

shmake.cli(name="TestOldestTS")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("newest_ts with list of globs", func(t *testing.T) {
		run := setupStarlarkRuntime(t)

		// Create test files in different directories
		require.NoError(t, os.Mkdir("dir1", 0o755))
		require.NoError(t, os.Mkdir("dir2", 0o755))

		file1 := filepath.Join("dir1", "file.txt")
		file2 := filepath.Join("dir2", "file.log")

		require.NoError(t, os.WriteFile(file1, []byte("content1"), 0o644))
		time.Sleep(10 * time.Millisecond)
		require.NoError(t, os.WriteFile(file2, []byte("content2"), 0o644))

		// Get expected newest timestamp
		info2, err := os.Stat(file2)
		require.NoError(t, err)
		expectedNewest := info2.ModTime().Unix()

		withMainStar(t, `
def test_action(ctx):
    result = shmake.newest_ts(['dir1/*.txt', 'dir2/*.log'])
    expected = `+strconv.FormatInt(expectedNewest, 10)+`
    if result != expected:
        fail('expected ' + str(expected) + ', got: ' + str(result))

shmake.cli(name="TestNewestTSList")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("oldest_ts with list of globs", func(t *testing.T) {
		run := setupStarlarkRuntime(t)

		// Create test files in different directories
		require.NoError(t, os.Mkdir("src", 0o755))
		require.NoError(t, os.Mkdir("docs", 0o755))

		file1 := filepath.Join("src", "main.go")
		file2 := filepath.Join("docs", "readme.md")

		require.NoError(t, os.WriteFile(file1, []byte("package main"), 0o644))
		info1, err := os.Stat(file1)
		require.NoError(t, err)
		expectedOldest := info1.ModTime().Unix()

		time.Sleep(10 * time.Millisecond)
		require.NoError(t, os.WriteFile(file2, []byte("# README"), 0o644))

		withMainStar(t, `
def test_action(ctx):
    result = shmake.oldest_ts(['src/*.go', 'docs/*.md'])
    expected = `+strconv.FormatInt(expectedOldest, 10)+`
    if result != expected:
        fail('expected ' + str(expected) + ', got: ' + str(result))

shmake.cli(name="TestOldestTSList")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("error when no files match", func(t *testing.T) {
		run := setupStarlarkRuntime(t)

		withMainStar(t, `
def test_action(ctx):
    try:
        result = shmake.newest_ts('nonexistent*.xyz')
        fail('expected error for non-matching glob')
    except Exception as e:
        if 'no files found' not in str(e):
            fail('expected "no files found" error, got: ' + str(e))

shmake.cli(name="TestNoFilesError")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("skips directories", func(t *testing.T) {
		run := setupStarlarkRuntime(t)

		// Create a directory and a file
		require.NoError(t, os.Mkdir("testdir", 0o755))
		require.NoError(t, os.WriteFile("testfile.txt", []byte("content"), 0o644))

		// Get expected timestamp from the file
		info, err := os.Stat("testfile.txt")
		require.NoError(t, err)
		expectedTS := info.ModTime().Unix()

		withMainStar(t, `
def test_action(ctx):
    result = shmake.newest_ts('test*')
    expected = `+strconv.FormatInt(expectedTS, 10)+`
    if result != expected:
        fail('expected ' + str(expected) + ', got: ' + str(result))

shmake.cli(name="TestSkipsDirectories")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}
