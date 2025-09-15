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
    assert_equals('hello world', result.stdout, 'expected hello world')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("captures command output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('printf "line1\\nline2\\nline3"')
    expected = '''line1\nline2\nline3'''
    assert_equals(expected, result.stdout, 'expected multiline output')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("works with shell variables", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('VAR="test value" && echo $VAR')
    assert_equals('test value', result.stdout, 'expected test value')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles command with options prefix", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "prefixed output"', prefix='TEST')
    assert_equals('prefixed output', result.stdout, 'expected prefixed output')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("trims whitespace from output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "  content with spaces  "')
    assert_equals('content with spaces', result.stdout, 'expected content with spaces')
    
    # Test trailing newline is trimmed
    result2 = shell('printf "no newline here"')
    assert_equals('no newline here', result2.stdout, 'expected no newline here')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles empty command output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('true')  # command that produces no output
    assert_equals('', result.stdout, 'expected empty string')

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
    assert_equals('test content', result.stdout, 'expected "test content"')
    
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
    assert_equals('error message', result.stderr, 'expected "error message" in stderr')
    assert_empty(result.stdout, 'expected empty stdout')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles successful command exit code", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('exit 0')
    assert_zero(result.exit_code, 'expected exit code 0')
    assert_true(result.success, 'expected success to be True')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles failed command exit code", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('exit 1')
    assert_equals(1, result.exit_code, 'expected exit code 1')
    assert_true(not result.success, 'expected success to be False')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles command with both stdout and stderr", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "stdout message" && echo "stderr message" >&2')
    assert_equals('stdout message', result.stdout, 'expected "stdout message"')
    assert_equals('stderr message', result.stderr, 'expected "stderr message"')
    assert_true(result.success, 'expected success to be True')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("result truthiness matches success status", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    success_result = shell('exit 0')
    assert_true(success_result, 'successful result should be truthy')
    
    fail_result = shell('exit 1')
    assert_true(not fail_result, 'failed result should be falsy')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("string representation returns stdout", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "test output"')
    str_result = str(result)
    assert_equals('test output', str_result, 'expected string representation to be "test output"')

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
    assert_empty(result.stdout, 'stdout should be empty with no_output=True')
    assert_empty(result.stderr, 'stderr should be empty with no_output=True')
    # Exit code and success should still be captured
    assert_zero(result.exit_code, 'exit code should be 0')
    assert_true(result.success, 'success should be True')

cli(name="TestShell", usage="Test shell functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("stdout and stderr capture regression test", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test stdout capture
    stdout_result = shell('echo "stdout test"')
    assert_equals('stdout test', stdout_result.stdout, 'REGRESSION: stdout not captured correctly')
    
    # Test stderr capture
    stderr_result = shell('echo "stderr test" >&2')
    assert_equals('stderr test', stderr_result.stderr, 'REGRESSION: stderr not captured correctly')
    
    # Test both stdout and stderr capture
    both_result = shell('echo "out" && echo "err" >&2')
    assert_equals('out', both_result.stdout, 'REGRESSION: stdout not captured when both present')
    assert_equals('err', both_result.stderr, 'REGRESSION: stderr not captured when both present')
    
    # Test multiline output capture
    multiline_result = shell('printf "line1\\nline2\\nline3"')
    expected_multiline = '''line1\nline2\nline3'''
    assert_equals(expected_multiline, multiline_result.stdout, 'REGRESSION: multiline stdout not captured')

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
    assert_equals('Flag value: true', result.stdout, 'expected "Flag value: true"')

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
    assert_equals('Building target: production', result.stdout, 'expected "Building target: production"')

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action, args=[string_arg("target")])
`, sindrtest.WithArgs("test", "production"))
	})

	t.Run("automatic string template expansion with mixed variables", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "Deploying debug={{.debug}} target={{.target}}"')
    assert_equals('Deploying debug=false target=backend', result.stdout, 'expected "Deploying debug=false target=backend"')

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
    assert_equals('no templates here', result.stdout, 'expected "no templates here"')

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating with additional variables", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "{{.custom_var}} {{.another_var}}"', custom_var="test123", another_var="value456")
    assert_equals('test123 value456', result.stdout, 'expected "test123 value456"')

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating overrides global variables", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "{{.project_name}}"', project_name="overridden")
    assert_equals('overridden', result.stdout, 'expected "overridden"')

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action, flags=[string_flag("project_name", default="default")])
`)
	})

	t.Run("kwargs templating with context flags and args", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "{{.target}} {{.debug}} {{.custom_key}}"', custom_key="custom_value")
    assert_equals('production false custom_value', result.stdout, 'expected "production false custom_value"')

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
    assert_equals('Building backend with enabled', result.stdout, 'expected "Building backend with enabled"')

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating with complex data types", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = shell('echo "Number: {{.number}} Boolean: {{.flag}}"', number=42, flag=True)
    assert_equals('Number: 42 Boolean: true', result.stdout, 'expected "Number: 42 Boolean: true"')

cli(name="TestShellTemplating", usage="Test shell automatic templating")
command(name="test", action=test_action)
`)
	})
}
