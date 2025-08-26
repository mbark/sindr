package internal_test

import (
	"testing"

	"github.com/mbark/sindr/internal/sindrtest"
)

func TestExec(t *testing.T) {
	t.Run("executes basic command with python3", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='python3', command='print("hello from python")')
    if result.stdout != 'hello from python':
        fail('expected "hello from python", got: ' + str(result.stdout))

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("executes command with shell interpreter", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='echo "shell output"')
    if result.stdout != 'shell output':
        fail('expected "shell output", got: ' + str(result.stdout))

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles multiline commands", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    command = '''echo "line1"
echo "line2"
echo "line3"'''
    result = exec(bin='sh', command=command)
    expected = 'line1\\nline2\\nline3'
    if result.stdout != expected:
        fail('expected "' + expected + '", got: ' + str(result.stdout))

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("captures stderr output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='echo "error message" >&2')
    if result.stderr != 'error message':
        fail('expected "error message" in stderr, got: ' + str(result.stderr))
    if result.stdout != '':
        fail('expected empty stdout, got: ' + str(result.stdout))

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles successful command exit code", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='exit 0')
    if result.exit_code != 0:
        fail('expected exit code 0, got: ' + str(result.exit_code))
    if not result.success:
        fail('expected success to be True')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles failed command exit code", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='exit 1')
    if result.exit_code != 1:
        fail('expected exit code 1, got: ' + str(result.exit_code))
    if result.success:
        fail('expected success to be False')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles command with both stdout and stderr", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='echo "stdout message" && echo "stderr message" >&2')
    if result.stdout != 'stdout message':
        fail('expected "stdout message", got: ' + str(result.stdout))
    if result.stderr != 'stderr message':
        fail('expected "stderr message", got: ' + str(result.stderr))
    if not result.success:
        fail('expected success to be True')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles command with prefix", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='echo "prefixed output"', prefix='EXEC')
    if result.stdout != 'prefixed output':
        fail('expected "prefixed output", got: ' + str(result.stdout))

cli(name="TestExec", usage="Test exec functionality")  
command(name="test", action=test_action)
`)
	})

	t.Run("no_output parameter prevents capturing output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='echo "should not be captured" && echo "error output" >&2', no_output=True)
    if result.stdout != '':
        fail('stdout should be empty with no_output=True, got: ' + str(result.stdout))
    if result.stderr != '':
        fail('stderr should be empty with no_output=True, got: ' + str(result.stderr))
    # Exit code and success should still be captured
    if result.exit_code != 0:
        fail('exit code should be 0, got: ' + str(result.exit_code))
    if not result.success:
        fail('success should be True')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("trims whitespace from output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='echo "  content with spaces  "')
    if result.stdout != 'content with spaces':
        fail('expected "content with spaces", got: ' + str(result.stdout))

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles empty command output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='true')  # command that produces no output
    if result.stdout != '':
        fail('expected empty string, got: ' + str(result.stdout))

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("result truthiness matches success status", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    success_result = exec(bin='sh', command='exit 0')
    if not success_result:
        fail('successful result should be truthy')
    
    fail_result = exec(bin='sh', command='exit 1')
    if fail_result:
        fail('failed result should be falsy')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("string representation returns stdout", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='echo "test output"')
    str_result = str(result)
    if str_result != 'test output':
        fail('expected string representation to be "test output", got: ' + str_result)

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("executes python script with variables", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    python_code = '''
name = "sindr"
version = "1.0"
print(f"Project {name} version {version}")
'''
    result = exec(bin='python3', command=python_code)
    if result.stdout != 'Project sindr version 1.0':
        fail('expected "Project sindr version 1.0", got: ' + str(result.stdout))

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("works with complex shell commands", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    shell_script = '''#!/bin/sh
# Create a test file and read it back
echo "test content" > test_exec.txt
cat test_exec.txt
# Clean up
rm test_exec.txt
'''
    result = exec(bin='sh', command=shell_script)
    if result.stdout != 'test content':
        fail('expected "test content", got: ' + str(result.stdout))

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles different interpreters", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test with awk
    awk_command = 'BEGIN { print "Hello from awk" }'
    result = exec(bin='awk', command=awk_command)
    if result.stdout != 'Hello from awk':
        fail('expected "Hello from awk", got: ' + str(result.stdout))

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles command execution failure", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    python_code = '''
import sys
print("This will fail")
sys.exit(42)
'''
    result = exec(bin='python3', command=python_code)
    if result.exit_code != 42:
        fail('expected exit code 42, got: ' + str(result.exit_code))
    if result.success:
        fail('expected success to be False')
    if result.stdout != 'This will fail':
        fail('expected "This will fail", got: ' + str(result.stdout))

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("captures syntax errors from interpreter", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    python_code = '''
# Invalid Python syntax
print("unclosed string
'''
    result = exec(bin='python3', command=python_code)
    if result.success:
        fail('expected command to fail due to syntax error')
    if result.exit_code == 0:
        fail('expected non-zero exit code for syntax error')
    # stderr should contain error information
    if result.stderr == '':
        fail('expected error output in stderr')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("works with valid parameters", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test that exec works when all required parameters are provided
    result = exec(bin='sh', command='echo "validation test"')
    if result.stdout != 'validation test':
        fail('expected "validation test", got: ' + str(result.stdout))
    if not result.success:
        fail('expected successful execution')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})
}
