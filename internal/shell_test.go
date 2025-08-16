package internal_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mbark/shmake/internal"
)

func TestShell(t *testing.T) {
	t.Run("executes basic shell command", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('echo "hello world"')
    if result.stdout != 'hello world':
        fail('expected "hello world", got: ' + str(result.stdout))

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("captures command output", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('printf "line1\\nline2\\nline3"')
    expected = 'line1\\nline2\\nline3'
    if result.stdout != expected:
        fail('expected: ' + expected + ', got: ' + str(result.stdout))

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("works with shell variables", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('VAR="test value" && echo $VAR')
    if result.stdout != 'test value':
        fail('expected "test value", got: ' + str(result.stdout))

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles command with options prefix", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('echo "prefixed output"', prefix='TEST')
    if result.stdout != 'prefixed output':
        fail('expected "prefixed output", got: ' + str(result.stdout))

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("trims whitespace from output", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('echo "  content with spaces  "')
    if result.stdout != 'content with spaces':
        fail('expected "content with spaces", got: ' + str(result.stdout))
    
    # Test trailing newline is trimmed
    result2 = shmake.shell('printf "no newline here"')
    if result2.stdout != 'no newline here':
        fail('expected "no newline here", got: ' + str(result2.stdout))

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles empty command output", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('true')  # command that produces no output
    if result.stdout != '':
        fail('expected empty string, got: ' + str(result.stdout))

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("works with complex commands", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    # Create a test file and read it back
    shmake.shell('echo "test content" > test.txt')
    result = shmake.shell('cat test.txt')
    if result.stdout != 'test content':
        fail('expected "test content", got: ' + str(result.stdout))
    
    # Clean up
    shmake.shell('rm test.txt')

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("captures stderr output", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('echo "error message" >&2')
    if result.stderr != 'error message':
        fail('expected "error message" in stderr, got: ' + str(result.stderr))
    if result.stdout != '':
        fail('expected empty stdout, got: ' + str(result.stdout))

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles successful command exit code", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('exit 0')
    if result.exit_code != 0:
        fail('expected exit code 0, got: ' + str(result.exit_code))
    if not result.success:
        fail('expected success to be True')

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles failed command exit code", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('exit 1')
    if result.exit_code != 1:
        fail('expected exit code 1, got: ' + str(result.exit_code))
    if result.success:
        fail('expected success to be False')

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles command with both stdout and stderr", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('echo "stdout message" && echo "stderr message" >&2')
    if result.stdout != 'stdout message':
        fail('expected "stdout message", got: ' + str(result.stdout))
    if result.stderr != 'stderr message':
        fail('expected "stderr message", got: ' + str(result.stderr))
    if not result.success:
        fail('expected success to be True')

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("result truthiness matches success status", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    success_result = shmake.shell('exit 0')
    if not success_result:
        fail('successful result should be truthy')
    
    fail_result = shmake.shell('exit 1')
    if fail_result:
        fail('failed result should be falsy')

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("string representation returns stdout", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('echo "test output"')
    str_result = str(result)
    if str_result != 'test output':
        fail('expected string representation to be "test output", got: ' + str_result)

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("no_output parameter prevents capturing output", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('echo "should not be captured" && echo "error output" >&2', no_output=True)
    print('DEBUG: stdout=' + repr(result.stdout))
    print('DEBUG: stderr=' + repr(result.stderr))
    if result.stdout != '':
        fail('stdout should be empty with no_output=True, got: ' + str(result.stdout))
    if result.stderr != '':
        fail('stderr should be empty with no_output=True, got: ' + str(result.stderr))
    # Exit code and success should still be captured
    if result.exit_code != 0:
        fail('exit code should be 0, got: ' + str(result.exit_code))
    if not result.success:
        fail('success should be True')

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("stdout and stderr capture regression test", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    # Test stdout capture
    stdout_result = shmake.shell('echo "stdout test"')
    if stdout_result.stdout != 'stdout test':
        fail('REGRESSION: stdout not captured correctly - expected "stdout test", got "' + str(stdout_result.stdout) + '"')
    
    # Test stderr capture
    stderr_result = shmake.shell('echo "stderr test" >&2')
    if stderr_result.stderr != 'stderr test':
        fail('REGRESSION: stderr not captured correctly - expected "stderr test", got "' + str(stderr_result.stderr) + '"')
    
    # Test both stdout and stderr capture
    both_result = shmake.shell('echo "out" && echo "err" >&2')
    if both_result.stdout != 'out':
        fail('REGRESSION: stdout not captured when both present - expected "out", got "' + str(both_result.stdout) + '"')
    if both_result.stderr != 'err':
        fail('REGRESSION: stderr not captured when both present - expected "err", got "' + str(both_result.stderr) + '"')
    
    # Test multiline output capture
    multiline_result = shmake.shell('printf "line1\\nline2\\nline3"')
    expected_multiline = 'line1\\nline2\\nline3'
    if multiline_result.stdout != expected_multiline:
        fail('REGRESSION: multiline stdout not captured - expected "' + expected_multiline + '", got "' + str(multiline_result.stdout) + '"')

shmake.cli(name="TestShellRegression", usage="Test shell capture regression")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}

// TestStartShellCmd tests the StartShellCmd function directly to ensure stdout/stderr capture works correctly.
// This is a regression test for the strings.Builder pointer bug.
func TestStartShellCmd(t *testing.T) {
	t.Run("captures stdout correctly", func(t *testing.T) {
		cmd := exec.CommandContext(context.Background(), "echo", "test stdout")
		result, err := internal.StartShellCmd(cmd, "", false)
		require.NoError(t, err)
		require.Equal(t, "test stdout", result.Stdout)
		require.Equal(t, "", result.Stderr)
		require.True(t, result.Success)
		require.Equal(t, 0, result.ExitCode)
	})

	t.Run("captures stderr correctly", func(t *testing.T) {
		cmd := exec.CommandContext(context.Background(), "sh", "-c", "echo 'test stderr' >&2")
		result, err := internal.StartShellCmd(cmd, "", false)
		require.NoError(t, err)
		require.Equal(t, "", result.Stdout)
		require.Contains(t, result.Stderr, "test stderr")
		require.True(t, result.Success)
		require.Equal(t, 0, result.ExitCode)
	})

	t.Run("captures both stdout and stderr", func(t *testing.T) {
		cmd := exec.CommandContext(
			context.Background(),
			"sh",
			"-c",
			"echo 'stdout msg' && echo 'stderr msg' >&2",
		)
		result, err := internal.StartShellCmd(cmd, "", false)
		require.NoError(t, err)
		require.Equal(t, "stdout msg", result.Stdout)
		require.Contains(t, result.Stderr, "stderr msg")
		require.True(t, result.Success)
		require.Equal(t, 0, result.ExitCode)
	})

	t.Run("captures multiline output", func(t *testing.T) {
		cmd := exec.CommandContext(
			context.Background(),
			"sh",
			"-c",
			"printf 'line1\\nline2\\nline3'",
		)
		result, err := internal.StartShellCmd(cmd, "", false)
		require.NoError(t, err)
		require.Equal(t, "line1\nline2\nline3", result.Stdout)
		// Stderr may contain shell initialization warnings, so just check it's captured
		require.True(t, result.Success)
		require.Equal(t, 0, result.ExitCode)
	})

	t.Run("handles no_output parameter correctly", func(t *testing.T) {
		cmd := exec.CommandContext(
			context.Background(),
			"sh",
			"-c",
			"echo 'should not capture' && echo 'stderr not capture' >&2",
		)
		result, err := internal.StartShellCmd(cmd, "", true)
		require.NoError(t, err)
		require.Equal(t, "", result.Stdout) // Should be empty with no_output=true
		require.Equal(t, "", result.Stderr) // Should be empty with no_output=true
		require.True(t, result.Success)
		require.Equal(t, 0, result.ExitCode)
	})

	t.Run("handles command failure", func(t *testing.T) {
		cmd := exec.CommandContext(
			context.Background(),
			"sh",
			"-c",
			"echo 'output before exit' && exit 42",
		)
		result, err := internal.StartShellCmd(cmd, "", false)
		require.NoError(t, err) // StartShellCmd should not return error for non-zero exit
		require.Equal(t, "output before exit", result.Stdout)
		require.False(t, result.Success)
		require.Equal(t, 42, result.ExitCode)
	})
}
