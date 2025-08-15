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

func TestGlob(t *testing.T) {
	t.Run("glob with single pattern", func(t *testing.T) {
		run := setupStarlarkRuntime(t)

		// Create test files
		require.NoError(t, os.WriteFile("glob1.txt", []byte("content1"), 0o644))
		require.NoError(t, os.WriteFile("glob2.txt", []byte("content2"), 0o644))
		require.NoError(t, os.WriteFile("other.log", []byte("log"), 0o644))

		withMainStar(t, `
def test_action(ctx):
    result = shmake.glob('glob*.txt')
    if len(result) != 2:
        fail('expected 2 files, got: ' + str(len(result)))
    
    # Convert to sorted list for consistent comparison
    files = sorted([str(f) for f in result])
    expected = ['glob1.txt', 'glob2.txt']
    if files != expected:
        fail('expected ' + str(expected) + ', got: ' + str(files))

shmake.cli(name="TestGlob")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("glob with list of patterns", func(t *testing.T) {
		run := setupStarlarkRuntime(t)

		// Create test files in different directories
		require.NoError(t, os.Mkdir("src", 0o755))
		require.NoError(t, os.Mkdir("test", 0o755))

		require.NoError(
			t,
			os.WriteFile(filepath.Join("src", "main.go"), []byte("package main"), 0o644),
		)
		require.NoError(
			t,
			os.WriteFile(filepath.Join("src", "utils.go"), []byte("package main"), 0o644),
		)
		require.NoError(
			t,
			os.WriteFile(filepath.Join("test", "test1.py"), []byte("import unittest"), 0o644),
		)
		require.NoError(
			t,
			os.WriteFile(filepath.Join("test", "test2.py"), []byte("import unittest"), 0o644),
		)

		withMainStar(t, `
def test_action(ctx):
    result = shmake.glob(['src/*.go', 'test/*.py'])
    if len(result) != 4:
        fail('expected 4 files, got: ' + str(len(result)))
    
    # Convert to sorted list for consistent comparison
    files = sorted([str(f) for f in result])
    expected = ['src/main.go', 'src/utils.go', 'test/test1.py', 'test/test2.py']
    if files != expected:
        fail('expected ' + str(expected) + ', got: ' + str(files))

shmake.cli(name="TestGlobList")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("glob returns empty list when no matches", func(t *testing.T) {
		run := setupStarlarkRuntime(t)

		withMainStar(t, `
def test_action(ctx):
    result = shmake.glob('nonexistent*.xyz')
    if len(result) != 0:
        fail('expected empty list, got: ' + str(len(result)) + ' files')

shmake.cli(name="TestGlobEmpty")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("glob skips directories", func(t *testing.T) {
		run := setupStarlarkRuntime(t)

		// Create a directory and a file with similar names
		require.NoError(t, os.Mkdir("skipdir", 0o755))
		require.NoError(t, os.WriteFile("skipfile.txt", []byte("content"), 0o644))

		withMainStar(t, `
def test_action(ctx):
    result = shmake.glob('skip*')
    if len(result) != 1:
        fail('expected 1 file, got: ' + str(len(result)))
    
    if str(result[0]) != 'skipfile.txt':
        fail('expected skipfile.txt, got: ' + str(result[0]))

shmake.cli(name="TestGlobSkipsDirectories")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("glob removes duplicates", func(t *testing.T) {
		run := setupStarlarkRuntime(t)

		// Create test files
		require.NoError(t, os.WriteFile("dup1.txt", []byte("content1"), 0o644))
		require.NoError(t, os.WriteFile("dup2.txt", []byte("content2"), 0o644))

		withMainStar(t, `
def test_action(ctx):
    # Use overlapping patterns that would match the same files
    result = shmake.glob(['dup*.txt', 'dup1.txt', '*.txt'])
    
    # Convert to sorted list for consistent comparison
    files = sorted([str(f) for f in result])
    
    # Should only have each file once, despite overlapping patterns
    unique_files = []
    for f in files:
        if f not in unique_files:
            unique_files.append(f)
    
    if len(files) != len(unique_files):
        fail('found duplicates in result: ' + str(files))

shmake.cli(name="TestGlobDuplicates")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("glob with invalid argument type", func(t *testing.T) {
		run := setupStarlarkRuntime(t)

		withMainStar(t, `
def test_action(ctx):
    try:
        result = shmake.glob(123)
        fail('expected error for invalid argument type')
    except Exception as e:
        if 'must be a string or list' not in str(e):
            fail('expected "must be a string or list" error, got: ' + str(e))

shmake.cli(name="TestGlobInvalidArg")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}
