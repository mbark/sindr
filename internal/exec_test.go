package internal_test

import (
	"testing"

	"github.com/mbark/sindr/internal/sindrtest"
)

func TestExec(t *testing.T) {
	t.Run("executes basic command with python3", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('python3', 'print("hello from python")')
    assert_equals('hello from python', result.stdout, 'expected hello from python')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("executes command with shell interpreter", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "shell output"')
    assert_equals('shell output', result.stdout, 'expected shell output')

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
    result = exec('sh', command)
    expected = '''line1\nline2\nline3'''
    assert_equals(expected, result.stdout, 'expected multiline output')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("captures stderr output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "error message" >&2')
    assert_equals('error message', result.stderr, 'expected error message in stderr')
    assert_equals('', result.stdout, 'expected empty stdout')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles successful command exit code", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'exit 0')
    assert_equals(0, result.exit_code, 'expected exit code 0')
    assert_equals(True, result.success, 'expected success to be True')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles failed command exit code", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'exit 1')
    assert_equals(1, result.exit_code, 'expected exit code 1')
    assert_equals(False, result.success, 'expected success to be False')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles command with both stdout and stderr", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "stdout message" && echo "stderr message" >&2')
    assert_equals('stdout message', result.stdout, 'expected stdout message')
    assert_equals('stderr message', result.stderr, 'expected stderr message')
    assert_equals(True, result.success, 'expected success to be True')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles command with prefix", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "prefixed output"', prefix='EXEC')
    assert_equals('prefixed output', result.stdout, 'expected prefixed output')

cli(name="TestExec", usage="Test exec functionality")  
command(name="test", action=test_action)
`)
	})

	t.Run("no_output parameter prevents capturing output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "should not be captured" && echo "error output" >&2', no_output=True)
    assert_equals('', result.stdout, 'stdout should be empty with no_output=True')
    assert_equals('', result.stderr, 'stderr should be empty with no_output=True')
    # Exit code and success should still be captured
    assert_equals(0, result.exit_code, 'exit code should be 0')
    assert_equals(True, result.success, 'success should be True')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("trims whitespace from output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "  content with spaces  "')
    assert_equals('content with spaces', result.stdout, 'expected content with spaces')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles empty command output", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'true')  # command that produces no output
    assert_equals('', result.stdout, 'expected empty string')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("result truthiness matches success status", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    success_result = exec('sh', 'exit 0')
    assert_true(success_result, 'successful result should be truthy')
    
    fail_result = exec('sh', 'exit 1')
    assert_true(not fail_result, 'failed result should be falsy')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("string representation returns stdout", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "test output"')
    str_result = str(result)
    assert_equals('test output', str_result, 'expected string representation to be test output')

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
    result = exec('python3', python_code)
    assert_equals('Project sindr version 1.0', result.stdout, 'expected Project sindr version 1.0')

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
    result = exec('sh', shell_script)
    assert_equals('test content', result.stdout, 'expected test content')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("handles different interpreters", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test with awk
    awk_command = 'print "Hello from perl\n";'
    result = exec('perl', awk_command)
    assert_equals('Hello from perl', result.stdout, 'expected "Hello from perl"')

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
    result = exec('python3', python_code)
    assert_equals(42, result.exit_code, 'expected exit code 42')
    assert_true(not result.success, 'expected success to be False')
    assert_equals('This will fail', result.stdout, 'expected "This will fail"')

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
    result = exec('python3', python_code)
    assert_true(not result.success, 'expected command to fail due to syntax error')
    assert_non_zero(result.exit_code, 'expected non-zero exit code for syntax error')
    # stderr should contain error information
    assert_not_empty(result.stderr, 'expected error output in stderr')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("works with valid parameters", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test that exec works when all required parameters are provided
    result = exec('sh', 'echo "validation test"')
    assert_equals('validation test', result.stdout, 'expected "validation test"')
    assert_true(result.success, 'expected successful execution')

cli(name="TestExec", usage="Test exec functionality")
command(name="test", action=test_action)
`)
	})
}

func TestExecTemplating(t *testing.T) {
	t.Run("automatic string template expansion with context flags", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "Debug mode: {{.debug}}"')
    assert_equals('Debug mode: true', result.stdout, 'expected "Debug mode: true"')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action, flags=[
    bool_flag("debug", default=True)
])
`)
	})

	t.Run("automatic string template expansion with context args", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "Processing environment: {{.environment}}"')
    assert_equals('Processing environment: development', result.stdout, 'expected "Processing environment: development"')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action, args=[string_arg("environment")])
`, sindrtest.WithArgs("test", "development"))
	})

	t.Run("automatic string template expansion with python command", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    python_code = '''print("App with verbose={{.verbose}}")'''
    result = exec('python3', python_code)
    assert_equals('App with verbose=false', result.stdout, 'expected "App with verbose=false"')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action, flags=[
	bool_flag("verbose")
])
`)
	})

	t.Run("automatic string template expansion with multiline command", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    shell_script = '''echo "Database: {{.database}}"
echo "Host: {{.host}}"
echo "Mode: {{.mode}}"'''
    result = exec('sh', shell_script)
    expected = '''Database: postgres\nHost: localhost\nMode: production'''
    assert_equals(expected, result.stdout, 'expected multiline output to match')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action, args=[string_arg("database"), string_arg("host"), string_arg("mode")])
`, sindrtest.WithArgs("test", "postgres", "localhost", "production"))
	})

	t.Run("automatic string template expansion with prefix option", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "Loading config: {{.config_file}}"', prefix='CONFIG', config_file="app.config")
    assert_equals('Loading config: app.config', result.stdout, 'expected "Loading config: app.config"')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("template expansion handles commands without templates", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "no template variables here"')
    assert_equals('no template variables here', result.stdout, 'expected "no template variables here"')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating with additional variables", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "{{.instance}} {{.region}}"', instance="web-01", region="us-west-2")
    assert_equals('web-01 us-west-2', result.stdout, 'expected "web-01 us-west-2"')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating overrides global variables", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "{{.service_name}}"', service_name="overridden-service")
    assert_equals('overridden-service', result.stdout, 'expected "overridden-service"')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action, args=[string_arg("service_name")])
`, sindrtest.WithArgs("test", "default"))
	})

	t.Run("kwargs templating with context flags and args", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "{{.environment}} {{.debug}} {{.deployment_id}}"', deployment_id="deploy-123")
    assert_equals('development true deploy-123', result.stdout, 'expected "development true deploy-123"')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action, args=[string_arg("environment")], flags=[
    bool_flag("debug", default=True)
])
`, sindrtest.WithArgs("test", "development"))
	})

	t.Run("kwargs templating with python command", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    python_code = '''print("Processing {{.task_name}} with ID {{.task_id}} status {{.status}}")'''
    result = exec('python3', python_code, task_name="backup", task_id="task-456", status="running")
    assert_equals('Processing backup with ID task-456 status running', result.stdout, 'expected "Processing backup with ID task-456 status running"')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating with prefix option", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec('sh', 'echo "Executing {{.operation}} on {{.resource}}"', prefix='EXEC', operation="deploy", resource="cluster")
    assert_equals('Executing deploy on cluster', result.stdout, 'expected "Executing deploy on cluster"')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating with multiline command", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    shell_script = '''echo "Database: {{.database}}"
echo "User: {{.user}}"
echo "Backup ID: {{.backup_id}}"'''
    result = exec('sh', shell_script, database="production_db", user="admin", backup_id="backup-789")
    expected = '''Database: production_db\nUser: admin\nBackup ID: backup-789'''
    assert_equals(expected, result.stdout, 'expected multiline database output to match')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating with complex data types", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    python_code = '''print(f"Count: {{.count}} Enabled: {{.enabled}} Rate: {{.rate}}")'''
    result = exec('python3', python_code, count=100, enabled=False, rate=1.5)
    assert_equals('Count: 100 Enabled: false Rate: 1.5', result.stdout, 'expected "Count: 100 Enabled: false Rate: 1.5"')

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})
}
