package shmake_test

import (
	"testing"
)

func TestShell(t *testing.T) {
	t.Run("executes basic shell command", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.shell('echo "hello world"')
    if result != 'hello world':
        print('ERROR: expected "hello world", got: ' + str(result))
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
    if result != expected:
        print('ERROR: expected: ' + expected + ', got: ' + str(result))
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
    if result != 'test value':
        print('ERROR: expected "test value", got: ' + str(result))
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
    if result != 'prefixed output':
        print('ERROR: expected "prefixed output", got: ' + str(result))
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
    if result != 'content with spaces':
        print('ERROR: expected "content with spaces", got: ' + str(result))
        return
    
    # Test trailing newline is trimmed
    result2 = shmake.shell('printf "no newline here"')
    if result2 != 'no newline here':
        print('ERROR: expected "no newline here", got: ' + str(result2))
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
    if result != '':
        print('ERROR: expected empty string, got: ' + str(result))
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
    if result != 'test content':
        print('ERROR: expected "test content", got: ' + str(result))
        return
    
    # Clean up
    shmake.shell('rm test.txt')

shmake.cli(name="TestShell", usage="Test shell functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}
