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
    expected = '''line1\nline2\nline3'''
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
    awk_command = 'print "Hello from perl\n";'
    result = exec(bin='perl', command=awk_command)
    if result.stdout != 'Hello from perl':
        fail('expected "Hello from perl", got: ' + str(result.stdout))

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

func TestExecTemplating(t *testing.T) {
	t.Run("automatic string template expansion with global variables", func(t *testing.T) {
		sindrtest.Test(t, `
# Define global variables
service_name = "api-server"
port = 8080

def test_action(ctx):
    result = exec(bin='sh', command='echo "Starting {{.service_name}} on port {{.port}}"')
    if result.stdout != 'Starting api-server on port 8080':
        fail('expected "Starting api-server on port 8080", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("automatic string template expansion with context flags", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='echo "Debug mode: {{.debug}}"')
    if result.stdout != 'Debug mode: true':
        fail('expected "Debug mode: true", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action, flags=[
    {
        "name": "debug",
        "type": "bool",
        "default": True,
    }
])
`)
	})

	t.Run("automatic string template expansion with context args", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='echo "Processing environment: {{.environment}}"')
    if result.stdout != 'Processing environment: development':
        fail('expected "Processing environment: development", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action, args=["environment"])
`)
	})

	t.Run("automatic string template expansion with python command", func(t *testing.T) {
		sindrtest.Test(t, `
# Global variables
app_version = "2.1.0"
build_number = 42

def test_action(ctx):
    python_code = '''print("App version {{.app_version}} build {{.build_number}} with verbose={{.verbose}}")'''
    result = exec(bin='python3', command=python_code)
    if result.stdout != 'App version 2.1.0 build 42 with verbose=false':
        fail('expected "App version 2.1.0 build 42 with verbose=false", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action, flags=[
    {
        "name": "verbose",
        "type": "bool", 
        "default": False,
    }
])
`)
	})

	t.Run("automatic string template expansion with multiline command", func(t *testing.T) {
		sindrtest.Test(t, `
# Global variables
database = "postgres"
host = "localhost"

def test_action(ctx):
    shell_script = '''echo "Database: {{.database}}"
echo "Host: {{.host}}"
echo "Mode: {{.mode}}"'''
    result = exec(bin='sh', command=shell_script)
    expected = '''Database: postgres\nHost: localhost\nMode: production'''
    if result.stdout != expected:
        fail('expected "' + expected + '", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action, args=["mode"])
`)
	})

	t.Run("automatic string template expansion with prefix option", func(t *testing.T) {
		sindrtest.Test(t, `
config_file = "app.config"

def test_action(ctx):
    result = exec(bin='sh', command='echo "Loading config: {{.config_file}}"', prefix='CONFIG')
    if result.stdout != 'Loading config: app.config':
        fail('expected "Loading config: app.config", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("template expansion handles commands without templates", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='echo "no template variables here"')
    if result.stdout != 'no template variables here':
        fail('expected "no template variables here", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating with additional variables", func(t *testing.T) {
		sindrtest.Test(t, `
# Global variables
service_name = "web-server"

def test_action(ctx):
    result = exec(bin='sh', command='echo "{{.service_name}} {{.instance}} {{.region}}"', instance="web-01", region="us-west-2")
    if result.stdout != 'web-server web-01 us-west-2':
        fail('expected "web-server web-01 us-west-2", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating overrides global variables", func(t *testing.T) {
		sindrtest.Test(t, `
# Global variables
service_name = "original-service"
port = 8080

def test_action(ctx):
    result = exec(bin='sh', command='echo "{{.service_name}} {{.port}}"', service_name="overridden-service")
    if result.stdout != 'overridden-service 8080':
        fail('expected "overridden-service 8080", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating with context flags and args", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='echo "{{.environment}} {{.debug}} {{.deployment_id}}"', deployment_id="deploy-123")
    if result.stdout != 'development true deploy-123':
        fail('expected "development true deploy-123", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action, args=["environment"], flags=[
    {
        "name": "debug",
        "type": "bool",
        "default": True,
    }
])
`)
	})

	t.Run("kwargs templating with python command", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    python_code = '''print("Processing {{.task_name}} with ID {{.task_id}} status {{.status}}")'''
    result = exec(bin='python3', command=python_code, task_name="backup", task_id="task-456", status="running")
    if result.stdout != 'Processing backup with ID task-456 status running':
        fail('expected "Processing backup with ID task-456 status running", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating with prefix option", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = exec(bin='sh', command='echo "Executing {{.operation}} on {{.resource}}"', prefix='EXEC', operation="deploy", resource="cluster")
    if result.stdout != 'Executing deploy on cluster':
        fail('expected "Executing deploy on cluster", got: ' + str(result.stdout))

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
    result = exec(bin='sh', command=shell_script, database="production_db", user="admin", backup_id="backup-789")
    expected = '''Database: production_db\nUser: admin\nBackup ID: backup-789'''
    if result.stdout != expected:
        fail('expected "' + expected + '", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})

	t.Run("kwargs templating with complex data types", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    python_code = '''print(f"Count: {{.count}} Enabled: {{.enabled}} Rate: {{.rate}}")'''
    result = exec(bin='python3', command=python_code, count=100, enabled=False, rate=1.5)
    if result.stdout != 'Count: 100 Enabled: false Rate: 1.5':
        fail('expected "Count: 100 Enabled: false Rate: 1.5", got: ' + str(result.stdout))

cli(name="TestExecTemplating", usage="Test exec automatic templating")
command(name="test", action=test_action)
`)
	})
}
