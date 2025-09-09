package internal_test

import (
	"testing"

	"github.com/mbark/sindr/internal/sindrtest"
)

func TestShell(t *testing.T) {
	t.Run("executes basic shell command", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "hello world"')
    if result.stdout != 'hello world':
        fail('expected "hello world", got: ' + str(result.stdout))

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("captures command output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('printf "line1\\nline2\\nline3"')
    expected = '''line1\nline2\nline3'''
    if result.stdout != expected:
        fail('expected: ' + expected + ', got: ' + str(result.stdout))

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("works with shell variables", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('VAR="test value" && echo $VAR')
    if result.stdout != 'test value':
        fail('expected "test value", got: ' + str(result.stdout))

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles command with options prefix", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "prefixed output"', prefix='TEST')
    if result.stdout != 'prefixed output':
        fail('expected "prefixed output", got: ' + str(result.stdout))

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("trims whitespace from output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "  content with spaces  "')
    if result.stdout != 'content with spaces':
        fail('expected "content with spaces", got: ' + str(result.stdout))
    
    # Test trailing newline is trimmed
    result2 = shell('printf "no newline here"')
    if result2.stdout != 'no newline here':
        fail('expected "no newline here", got: ' + str(result2.stdout))

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles empty command output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('true')  # command that produces no output
    if result.stdout != '':
        fail('expected empty string, got: ' + str(result.stdout))

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("works with complex commands", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Create a test file and read it back
    shell('echo "test content" > test.txt')
    result = shell('cat test.txt')
    if result.stdout != 'test content':
        fail('expected "test content", got: ' + str(result.stdout))
    
    # Clean up
    shell('rm test.txt')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("captures stderr output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "error message" >&2')
    if result.stderr != 'error message':
        fail('expected "error message" in stderr, got: ' + str(result.stderr))
    if result.stdout != '':
        fail('expected empty stdout, got: ' + str(result.stdout))

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles successful command exit code", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('exit 0')
    if result.exit_code != 0:
        fail('expected exit code 0, got: ' + str(result.exit_code))
    if not result.success:
        fail('expected success to be True')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles failed command exit code", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('exit 1')
    if result.exit_code != 1:
        fail('expected exit code 1, got: ' + str(result.exit_code))
    if result.success:
        fail('expected success to be False')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles command with both stdout and stderr", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "stdout message" && echo "stderr message" >&2')
    if result.stdout != 'stdout message':
        fail('expected "stdout message", got: ' + str(result.stdout))
    if result.stderr != 'stderr message':
        fail('expected "stderr message", got: ' + str(result.stderr))
    if not result.success:
        fail('expected success to be True')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("result truthiness matches success status", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    success_result = shell('exit 0')
    if not success_result:
        fail('successful result should be truthy')
    
    fail_result = shell('exit 1')
    if fail_result:
        fail('failed result should be falsy')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("string representation returns stdout", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "test output"')
    str_result = str(result)
    if str_result != 'test output':
        fail('expected string representation to be "test output", got: ' + str_result)

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("no_output parameter prevents capturing output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "should not be captured" && echo "error output" >&2', no_output=True)
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

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("stdout and stderr capture regression test", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test stdout capture
    stdout_result = shell('echo "stdout test"')
    if stdout_result.stdout != 'stdout test':
        fail('REGRESSION: stdout not captured correctly - expected "stdout test", got "' + str(stdout_result.stdout) + '"')
    
    # Test stderr capture
    stderr_result = shell('echo "stderr test" >&2')
    if stderr_result.stderr != 'stderr test':
        fail('REGRESSION: stderr not captured correctly - expected "stderr test", got "' + str(stderr_result.stderr) + '"')
    
    # Test both stdout and stderr capture
    both_result = shell('echo "out" && echo "err" >&2')
    if both_result.stdout != 'out':
        fail('REGRESSION: stdout not captured when both present - expected "out", got "' + str(both_result.stdout) + '"')
    if both_result.stderr != 'err':
        fail('REGRESSION: stderr not captured when both present - expected "err", got "' + str(both_result.stderr) + '"')
    
    # Test multiline output capture
    multiline_result = shell('printf "line1\\nline2\\nline3"')
    expected_multiline = '''line1\nline2\nline3'''
    if multiline_result.stdout != expected_multiline:
        fail('REGRESSION: multiline stdout not captured - expected "' + expected_multiline + '", got "' + str(multiline_result.stdout) + '"')

cli(name="TestShellRegression", usage="Test shell capture regression")
command(name="test", action=test_action)
`)
	})
}

func TestShellTemplating(t *testing.T) {
	t.Run("automatic string template expansion with context flags", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "Flag value: {{.verbose}}"')
    if result.stdout != 'Flag value: true':
        fail('expected "Flag value: true", got: ' + str(result.stdout))

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action, flags=[
    bool_flag("verbose", default=True)
])
`)
	})

	t.Run("automatic string template expansion with context args", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "Building target: {{.target}}"')
    if result.stdout != 'Building target: production':
        fail('expected "Building target: production", got: ' + str(result.stdout))

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action, args=[string_arg("target")])
`, sindrtest.WithArgs("test", "production"))
	})

	t.Run("automatic string template expansion with mixed variables", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "Deploying debug={{.debug}} target={{.target}}"')
    if result.stdout != 'Deploying debug=false target=backend':
        fail('expected "Deploying debug=false target=backend", got: ' + str(result.stdout))

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action, args=[string_arg("target")], flags=[
	bool_flag("debug"),
])
`, sindrtest.WithArgs("test", "backend"))
	})

	t.Run("template expansion handles empty templates", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "no templates here"')
    if result.stdout != 'no templates here':
        fail('expected "no templates here", got: ' + str(result.stdout))

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating with additional variables", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "{{.custom_var}} {{.another_var}}"', custom_var="test123", another_var="value456")
    if result.stdout != 'test123 value456':
        fail('expected "test123 value456", got: ' + str(result.stdout))

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating overrides global variables", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "{{.project_name}}"', project_name="overridden")
    if result.stdout != 'overridden':
        fail('expected "overridden", got: ' + str(result.stdout))

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action, flags=[string_flag("project_name", default="default")])
`)
	})

	t.Run("kwargs templating with context flags and args", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "{{.target}} {{.debug}} {{.custom_key}}"', custom_key="custom_value")
    if result.stdout != 'production false custom_value':
        fail('expected "production false custom_value", got: ' + str(result.stdout))

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action, args=[string_arg("target")], flags=[
	bool_flag("debug"),
])
`, sindrtest.WithArgs("test", "production"))
	})

	t.Run("kwargs templating with prefix option", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "Building {{.component}} with {{.custom_flag}}"', prefix='BUILD', component="backend", custom_flag="enabled")
    if result.stdout != 'Building backend with enabled':
        fail('expected "Building backend with enabled", got: ' + str(result.stdout))

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating with complex data types", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "Number: {{.number}} Boolean: {{.flag}}"', number=42, flag=True)
    if result.stdout != 'Number: 42 Boolean: true':
        fail('expected "Number: 42 Boolean: true", got: ' + str(result.stdout))

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action)
`)
	})
}
