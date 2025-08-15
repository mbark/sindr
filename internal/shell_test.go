package internal_test

import (
	"testing"
)

func TestShell(t *testing.T) {
	t.Run("executes basic shell command", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('echo "hello world"')
    if result.stdout != 'hello world':
        print('ERROR: expected "hello world", got: ' + str(result.stdout))
        return

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
        print('ERROR: expected: ' + expected + ', got: ' + str(result.stdout))
        return

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
        print('ERROR: expected "test value", got: ' + str(result.stdout))
        return

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
        print('ERROR: expected "prefixed output", got: ' + str(result.stdout))
        return

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
        print('ERROR: expected "content with spaces", got: ' + str(result.stdout))
        return
    
    # Test trailing newline is trimmed
    result2 = shmake.shell('printf "no newline here"')
    if result2.stdout != 'no newline here':
        print('ERROR: expected "no newline here", got: ' + str(result2.stdout))
        return

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
        print('ERROR: expected empty string, got: ' + str(result.stdout))
        return

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
        print('ERROR: expected "test content", got: ' + str(result.stdout))
        return
    
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
        print('ERROR: expected "error message" in stderr, got: ' + str(result.stderr))
        return
    if result.stdout != '':
        print('ERROR: expected empty stdout, got: ' + str(result.stdout))
        return

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
        print('ERROR: expected exit code 0, got: ' + str(result.exit_code))
        return
    if not result.success:
        print('ERROR: expected success to be True')
        return

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
        print('ERROR: expected exit code 1, got: ' + str(result.exit_code))
        return
    if result.success:
        print('ERROR: expected success to be False')
        return

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
        print('ERROR: expected "stdout message", got: ' + str(result.stdout))
        return
    if result.stderr != 'stderr message':
        print('ERROR: expected "stderr message", got: ' + str(result.stderr))
        return
    if not result.success:
        print('ERROR: expected success to be True')
        return

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
        print('ERROR: successful result should be truthy')
        return
    
    fail_result = shmake.shell('exit 1')
    if fail_result:
        print('ERROR: failed result should be falsy')
        return

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
        print('ERROR: expected string representation to be "test output", got: ' + str_result)
        return

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
        print('ERROR: expected empty stdout when no_output=True, got: ' + str(result.stdout))
        fail('stdout should be empty with no_output=True')
    if result.stderr != '':
        print('ERROR: expected empty stderr when no_output=True, got: ' + str(result.stderr))
        fail('stderr should be empty with no_output=True')
    # Exit code and success should still be captured
    if result.exit_code != 0:
        print('ERROR: expected exit code 0, got: ' + str(result.exit_code))
        fail('exit code should be 0')
    if not result.success:
        print('ERROR: expected success to be True')
        fail('success should be True')

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}
