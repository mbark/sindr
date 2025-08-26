package internal_test

import (
	"testing"

	"github.com/mbark/sindr/internal/sindrtest"
)

func TestFileTimestamps(t *testing.T) {
	t.Run("newest_ts with single glob", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create test files with different modification times
    shell('echo "content1" > test1.txt')
    shell('sleep 0.1')  # Small delay to ensure different timestamps
    shell('echo "content2" > test2.txt')
    
    # Get the newer timestamp and test it
    result = newest_ts('test*.txt')
    
    # Get timestamp of test2.txt for comparison
    test2_ts_result = shell('stat -c %Y test2.txt 2>/dev/null || stat -f %m test2.txt')
    expected = int(test2_ts_result.stdout)
    
    if result != expected:
        fail('expected ' + str(expected) + ', got: ' + str(result))

cli(name="TestNewestTS")
command(name="test", action=test_action)
`)
	})

	t.Run("oldest_ts with single glob", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create test files with different modification times
    shell('echo "content1" > old1.txt')
    # Get timestamp of old1.txt before creating the second file
    old1_ts_result = shell('stat -c %Y old1.txt 2>/dev/null || stat -f %m old1.txt')
    expected = int(old1_ts_result.stdout)
    
    shell('sleep 0.1')  # Small delay to ensure different timestamps  
    shell('echo "content2" > old2.txt')
    
    result = oldest_ts('old*.txt')
    
    if result != expected:
        fail('expected ' + str(expected) + ', got: ' + str(result))

cli(name="TestOldestTS")
command(name="test", action=test_action)
`)
	})

	t.Run("newest_ts with list of globs", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create test files in different directories
    shell('mkdir -p dir1 dir2')
    shell('echo "content1" > dir1/file.txt')
    shell('sleep 0.1')  # Small delay to ensure different timestamps
    shell('echo "content2" > dir2/file.log')
    
    # Get timestamp of the newer file for comparison
    file2_ts_result = shell('stat -c %Y dir2/file.log 2>/dev/null || stat -f %m dir2/file.log')
    expected = int(file2_ts_result.stdout)
    
    result = newest_ts(['dir1/*.txt', 'dir2/*.log'])
    
    if result != expected:
        fail('expected ' + str(expected) + ', got: ' + str(result))

cli(name="TestNewestTSList")
command(name="test", action=test_action)
`)
	})

	t.Run("oldest_ts with list of globs", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create test files in different directories
    shell('mkdir -p src docs')
    shell('echo "package main" > src/main.go')
    
    # Get timestamp of the first file before creating the second
    file1_ts_result = shell('stat -c %Y src/main.go 2>/dev/null || stat -f %m src/main.go')
    expected = int(file1_ts_result.stdout)
    
    shell('sleep 0.1')  # Small delay to ensure different timestamps
    shell('echo "# README" > docs/readme.md')
    
    result = oldest_ts(['src/*.go', 'docs/*.md'])
    
    if result != expected:
        fail('expected ' + str(expected) + ', got: ' + str(result))

cli(name="TestOldestTSList")
command(name="test", action=test_action)
`)
	})

	t.Run("error when no files match", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    try:
        result = newest_ts('nonexistent*.xyz')
        fail('expected error for non-matching glob')
    except Exception as e:
        if 'no files found' not in str(e):
            fail('expected "no files found" error, got: ' + str(e))

cli(name="TestNoFilesError")
command(name="test", action=test_action)
`, sindrtest.ShouldFail())
	})

	t.Run("skips directories", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create a directory and a file
    shell('mkdir testdir')
    shell('echo "content" > testfile.txt')
    
    # Get timestamp of the file for comparison
    file_ts_result = shell('stat -c %Y testfile.txt 2>/dev/null || stat -f %m testfile.txt')
    expected = int(file_ts_result.stdout)
    
    result = newest_ts('test*')
    
    if result != expected:
        fail('expected ' + str(expected) + ', got: ' + str(result))

cli(name="TestSkipsDirectories")
command(name="test", action=test_action)
`)
	})
}

func TestGlob(t *testing.T) {
	t.Run("glob with single pattern", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create test files
    shell('echo "content1" > glob1.txt')
    shell('echo "content2" > glob2.txt')
    shell('echo "log" > other.log')
    
    result = glob('glob*.txt')
    if len(result) != 2:
        fail('expected 2 files, got: ' + str(len(result)))
    
    # Convert to sorted list for consistent comparison
    files = sorted([str(f) for f in result])
    expected = ['glob1.txt', 'glob2.txt']
    if files != expected:
        fail('expected ' + str(expected) + ', got: ' + str(files))

cli(name="TestGlob")
command(name="test", action=test_action)
`)
	})

	t.Run("glob with list of patterns", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create test files in different directories
    shell('mkdir -p src test')
    shell('echo "package main" > src/main.go')
    shell('echo "package main" > src/utils.go')
    shell('echo "import unittest" > test/test1.py')
    shell('echo "import unittest" > test/test2.py')
    
    result = glob(['src/*.go', 'test/*.py'])
    if len(result) != 4:
        fail('expected 4 files, got: ' + str(len(result)))
    
    # Convert to sorted list for consistent comparison
    files = sorted([str(f) for f in result])
    expected = ['src/main.go', 'src/utils.go', 'test/test1.py', 'test/test2.py']
    if files != expected:
        fail('expected ' + str(expected) + ', got: ' + str(files))

cli(name="TestGlobList")
command(name="test", action=test_action)
`)
	})

	t.Run("glob returns empty list when no matches", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = glob('nonexistent*.xyz')
    if len(result) != 0:
        fail('expected empty list, got: ' + str(len(result)) + ' files')

cli(name="TestGlobEmpty")
command(name="test", action=test_action)
`)
	})

	t.Run("glob skips directories", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create a directory and a file with similar names
    shell('mkdir skipdir')
    shell('echo "content" > skipfile.txt')
    
    result = glob('skip*')
    if len(result) != 1:
        fail('expected 1 file, got: ' + str(len(result)))
    
    if str(result[0]) != 'skipfile.txt':
        fail('expected skipfile.txt, got: ' + str(result[0]))

cli(name="TestGlobSkipsDirectories")
command(name="test", action=test_action)
`)
	})

	t.Run("glob removes duplicates", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create test files
    shell('echo "content1" > dup1.txt')
    shell('echo "content2" > dup2.txt')
    
    # Use overlapping patterns that would match the same files
    result = glob(['dup*.txt', 'dup1.txt', '*.txt'])
    
    # Convert to sorted list for consistent comparison
    files = sorted([str(f) for f in result])
    
    # Should only have each file once, despite overlapping patterns
    unique_files = []
    for f in files:
        if f not in unique_files:
            unique_files.append(f)
    
    if len(files) != len(unique_files):
        fail('found duplicates in result: ' + str(files))

cli(name="TestGlobDuplicates")
command(name="test", action=test_action)
`)
	})

	t.Run("glob with invalid argument type", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    try:
        result = glob(123)
        fail('expected error for invalid argument type')
    except Exception as e:
        if 'must be a string or list' not in str(e):
            fail('expected "must be a string or list" error, got: ' + str(e))

cli(name="TestGlobInvalidArg")
command(name="test", action=test_action)
`, sindrtest.ShouldFail())
	})
}
