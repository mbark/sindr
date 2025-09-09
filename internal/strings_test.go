package internal_test

import (
	"testing"

	"github.com/mbark/sindr/internal/sindrtest"
)

func TestTemplateString(t *testing.T) {
	t.Run("renders simple template", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = string('Hello {{.name}}!', name='World')
    if result != 'Hello World!':
        fail('expected "Hello World!", got: ' + str(result))

cli(name="TestTemplateString", usage="Test string templating")
command(name="test", action=test_action)
`)
	})

	t.Run("handles different variable types", func(t *testing.T) {
		sindrtest.Test(t, `
str_var = "text"
num_var = 42
bool_var = True

def test_action(ctx):
    result = string('{{.str_var}} {{.num_var}} {{.bool_var}}',str_var=str_var, num_var=num_var, bool_var=bool_var)
    print('result is', result)
    if result != 'text 42 true':
        fail('expected "text 42 true", got: ' + str(result))

cli(name="TestTemplateString", usage="Test string templating")
command(name="test", action=test_action)
`)
	})

	t.Run("handles template with conditional logic", func(t *testing.T) {
		sindrtest.Test(t, `
debug = True

def test_action(ctx):
    result = string('{{if .debug}}DEBUG: {{end}}message', message='test',debug=debug)
    if result != 'DEBUG: message':
        fail('expected "DEBUG: message", got: ' + str(result))

cli(name="TestTemplateString", usage="Test string templating")
command(name="test", action=test_action)
`)
	})

	t.Run("handles template with range", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    items = ['apple', 'banana', 'cherry']
    result = string('{{range .items}}{{.}} {{end}}', items=items)
    if result != 'apple banana cherry ':
        fail('expected "apple banana cherry ", got: ' + str(result))

cli(name="TestTemplateString", usage="Test string templating")
command(name="test", action=test_action)
`)
	})

	t.Run("handles complex nested templates", func(t *testing.T) {
		sindrtest.Test(t, `
app_name = "myapp"
version = "1.0.0"

def test_action(ctx):
    config = {
        "env": "production",
        "port": 8080,
        "features": ['auth', 'logging', 'metrics']
    }
    
    template = """
{{.app_name}} v{{.version}}
Environment: {{.config.env}}
Port: {{.config.port}}
Features:{{range .config.features}} {{.}}{{end}}
"""
    
    result = string(template, config=config, app_name=app_name, version=version)
	# use string to trim some whitespaces for the test
    expected = string("""
myapp v1.0.0
Environment: production
Port: 8080
Features: auth logging metrics
""")
    
    if result != expected:
		print('got')
		print(string(result))
		print('expected')
		print(string(expected))
        fail('template did not render correctly')

cli(name="TestTemplateString", usage="Test string templating")
command(name="test", action=test_action)
`)
	})

	t.Run("automatically includes context flags", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = string('Debug: {{.debug}}, Verbose: {{.verbose}}')
    if result != 'Debug: true, Verbose: false':
        fail('expected "Debug: true, Verbose: false", got: ' + str(result))

cli(name="TestTemplateString", usage="Test string templating")
command(name="test", action=test_action, flags=[
    bool_flag("debug", default=True),
    bool_flag("verbose", default=False)
])
`)
	})

	t.Run("automatically includes context args", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = string('Building {{.target}} for {{.environment}}')
    if result != 'Building backend for staging':
        fail('expected "Building backend for staging", got: ' + str(result))

cli(name="TestTemplateString", usage="Test string templating")
command(name="test", action=test_action, args=[string_arg("target"), string_arg("environment")])
`, sindrtest.WithArgs("test", "backend", "staging"))
	})

	t.Run("explicit variables override context variables", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Context flag should override the explicit parameter
    result = string('Mode: {{.mode}}', mode='explicit')
    if result != 'Mode: explicit':
        fail('expected "Mode: explicit", got: ' + str(result))

cli(name="TestTemplateString", usage="Test string templating")
command(name="test", action=test_action, flags=[
    string_flag("mode", default="development")
])
`, sindrtest.WithArgs("test", "--mode", "development"))
	})

	t.Run("explicit parameters can still be added alongside context", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    result = string('{{.verbose}} {{.custom}} {{.target}}', custom='extra')
    if result != 'false extra production':
        fail('expected "false extra production", got: ' + str(result))

cli(name="TestTemplateString", usage="Test string templating")
command(name="test", action=test_action, args=[string_arg("target")], flags=[
    bool_flag("verbose", default=False)
])
`, sindrtest.WithArgs("test", "production"))
	})

	t.Run("context access via direct flag/arg names", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test that context flags/args can be accessed directly by their names
    result = string('Flag some_flag: {{.some_flag}}, Arg some_arg: {{.some_arg}}')
    if result != 'Flag some_flag: true, Arg some_arg: test_value':
        fail('expected "Flag some_flag: true, Arg some_arg: test_value", got: ' + str(result))

cli(name="TestTemplateString", usage="Test string templating")
command(name="test", action=test_action, args=[string_arg("some_arg")], flags=[
    bool_flag("some-flag", default=True)
])
`, sindrtest.WithArgs("test", "--some-flag", "test_value"))
	})
}
