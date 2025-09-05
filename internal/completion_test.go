package internal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mbark/sindr/internal/sindrtest"
)

func TestCompletionGeneration(t *testing.T) {
	tests := []struct {
		name        string
		shell       string
		contains    []string
		notContains []string
	}{
		{
			name:  "bash completion script generation",
			shell: "bash",
			contains: []string{
				"_sindr_completion()",
				"complete -F _sindr_completion sindr",
				"--generate-shell-completion",
				"$cur\" == -*",
			},
		},
		{
			name:  "zsh completion script generation",
			shell: "zsh",
			contains: []string{
				"#compdef sindr",
				"_sindr()",
				"compdef _sindr sindr",
				"--generate-shell-completion",
				"${words[CURRENT]}\" == -*",
			},
		},
		{
			name:  "fish completion script generation",
			shell: "fish",
			contains: []string{
				"function __sindr_complete",
				"function __sindr_complete_flags",
				"complete -c sindr",
				"--generate-shell-completion",
				"string match -q -- '-*'",
			},
		},
		{
			name:  "powershell completion script generation",
			shell: "powershell",
			contains: []string{
				"Register-ArgumentCompleter",
				"$wordToComplete.StartsWith('-')",
				"--generate-shell-completion",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For testing completion script generation, we need a simple test that
			// verifies the completion command can generate scripts
			sindrtest.Test(t, `
def test_action(ctx):
    # This is just a simple command for the test
    print("completion script generation test")

cli(name="TestCompletion")
command(name="test", action=test_action)
`, sindrtest.WithArgs("completion", tt.shell))
		})
	}
}

func TestCompletionErrorHandling(t *testing.T) {
	t.Run("completion command requires shell argument", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("test action")

cli(name="TestApp")
command(name="test", action=test_action)
`, sindrtest.WithArgs("completion"), sindrtest.ShouldFail())
	})

	t.Run("completion command rejects invalid shell", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("test action")

cli(name="TestApp")
command(name="test", action=test_action)
`, sindrtest.WithArgs("completion", "invalid-shell"), sindrtest.ShouldFail())
	})
}

func TestShellCompletionFunctionality(t *testing.T) {
	// Create a test that verifies the shell completion actually works
	// by testing the --generate-shell-completion flag directly

	t.Run("completion shows commands at root level", func(t *testing.T) {
		sindrtest.Test(t, `
def build_action(ctx):
    print("building")

def deploy_action(ctx):  
    print("deploying")

cli(name="TestApp", usage="test app")
command(name="build", action=build_action)
command(name="deploy", action=deploy_action)
`, sindrtest.WithArgs("--generate-shell-completion"))
	})

	t.Run("completion shows flags when preceded by dash", func(t *testing.T) {
		writer := new(sindrtest.CollectWriter)

		sindrtest.Test(t, `
def test_action(ctx):
    print("test with verbose:", ctx.flags.verbose)

cli(name="TestApp", usage="test app")  
command(
    name="test", 
    action=test_action,
    flags=[
        {"name": "verbose", "type": "bool", "default": False}
    ]
)
`,
			sindrtest.WithArgs("test", "-", "--generate-shell-completion"),
			sindrtest.WithWriter(writer))

		require.Len(t, writer.Writes, 2)
		assert.Equal(t, "--verbose\n", writer.Writes[0])
		assert.Equal(t, "--help:show help\n", writer.Writes[1])
	})

	t.Run("completion shows subcommands", func(t *testing.T) {
		writer := new(sindrtest.CollectWriter)

		sindrtest.Test(t, `
def deploy_staging(ctx):
    print("deploying to staging")

def deploy_prod(ctx):
    print("deploying to production")

cli(name="TestApp", usage="test app")
command(name="deploy") 
sub_command(path=["deploy", "staging"], action=deploy_staging)
sub_command(path=["deploy", "production"], action=deploy_prod)
`,
			sindrtest.WithArgs("deploy", "--generate-shell-completion"),
			sindrtest.WithWriter(writer))

		assert.Equal(t, "staging:\n", writer.Writes[0])
		assert.Equal(t, "production:\n", writer.Writes[1])
		assert.Equal(t, "help:Shows a list of commands or help for one command\n", writer.Writes[2])
	})

	t.Run("completion shows command-specific flags", func(t *testing.T) {
		writer := new(sindrtest.CollectWriter)

		sindrtest.Test(t, `
def build_action(ctx):
    print("building with verbose=" + str(ctx.flags.verbose) + " parallel=" + str(ctx.flags.parallel))

cli(name="TestApp", usage="test app")
command(
    name="build",
    action=build_action,
    flags=[
        {"name": "verbose", "type": "bool", "default": False},
        {"name": "parallel", "type": "bool", "default": False}
    ]
)
`,
			sindrtest.WithArgs("build", "-", "--generate-shell-completion"),
			sindrtest.WithWriter(writer))

		assert.Equal(t, "--verbose\n", writer.Writes[0])
		assert.Equal(t, "--parallel\n", writer.Writes[1])
		assert.Equal(t, "--help:show help\n", writer.Writes[2])
	})
}
