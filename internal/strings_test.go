package internal_test

import (
	"testing"
)

func TestTemplateString(t *testing.T) {
	t.Run("renders simple template", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    result = shmake.string('Hello {{.name}}!', name='World')
    if result != 'Hello World!':
        print('ERROR: expected "Hello World!", got: ' + str(result))
        return

shmake.cli(name="TestTemplateString", usage="Test string templating")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles different variable types", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
str_var = "text"
num_var = 42
bool_var = True

def test_action(ctx):
    result = shmake.string('{{.str_var}} {{.num_var}} {{.bool_var}}')
    if result != 'text 42 true':
        print('ERROR: expected "text 42 true", got: ' + str(result))
        return

shmake.cli(name="TestTemplateString", usage="Test string templating")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("renders template without additional context", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
greeting = "Hello"
target = "World"

def test_action(ctx):
    result = shmake.string('{{.greeting}} {{.target}}!')
    if result != 'Hello World!':
        print('ERROR: expected "Hello World!", got: ' + str(result))
        return

shmake.cli(name="TestTemplateString", usage="Test string templating")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles template with conditional logic", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
debug = True

def test_action(ctx):
    result = shmake.string('{{if .debug}}DEBUG: {{end}}message', message='test')
    if result != 'DEBUG: message':
        print('ERROR: expected "DEBUG: message", got: ' + str(result))
        return

shmake.cli(name="TestTemplateString", usage="Test string templating")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles template with range", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    items = ['apple', 'banana', 'cherry']
    result = shmake.string('{{range .items}}{{.}} {{end}}', items=items)
    if result != 'apple banana cherry ':
        print('ERROR: expected "apple banana cherry ", got: ' + str(result))
        return

shmake.cli(name="TestTemplateString", usage="Test string templating")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles complex nested templates", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
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
    
    result = shmake.string(template, config=config)
    expected = """
myapp v1.0.0
Environment: production
Port: 8080
Features: auth logging metrics
"""
    
    if result != expected:
        print('ERROR: template did not render correctly')
        return

shmake.cli(name="TestTemplateString", usage="Test string templating")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}
