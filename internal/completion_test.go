package internal_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mbark/sindr/internal/sindrtest"
)

func TestShellCompletionFunctionality(t *testing.T) {
	t.Run("completion shows commands at root level", func(t *testing.T) {
		writer := new(sindrtest.CollectWriter)
		sindrtest.Test(t, `
def build_action(ctx):
    print("building")

def deploy_action(ctx):  
    print("deploying")

cli(name="TestApp", usage="test app")
command(name="build", action=build_action)
command(name="deploy", action=deploy_action)
`,
			sindrtest.WithArgs("__complete"),
			sindrtest.WithWriter(writer))

		require.Len(t, writer.Writes, 4)
		assert.Equal(t, "build\n", writer.Writes[0])
		assert.Equal(t, "deploy\n", writer.Writes[1])

		helpUsage := "Shows a list of commands or help for one command"
		assert.Equal(t, fmt.Sprintf("help\t%s\n", helpUsage), writer.Writes[2])
		assert.Equal(t, fmt.Sprintf("h\t%s (alias)\n", helpUsage), writer.Writes[3])
	})

	t.Run("completion shows flags at root level", func(t *testing.T) {
		writer := new(sindrtest.CollectWriter)
		sindrtest.Test(t, `
def build_action(ctx):
    print("building")

def deploy_action(ctx):  
    print("deploying")

cli(name="TestApp", usage="test app")
command(name="build", action=build_action)
command(name="deploy", action=deploy_action)
`,
			sindrtest.WithArgs("__complete", "-"),
			sindrtest.WithWriter(writer))

		require.NotEmpty(t, writer.Writes)
		joined := strings.Join(writer.Writes, "")
		assert.Equal(t, strings.TrimSpace(
			`
--cache-dir	--cache-dir string	path to the Starlark config file
--file-name	--file-name string, -f string	path to the Starlark config file
-f	--file-name string, -f string	path to the Starlark config file
--line-numbers	--line-numbers, -l	print logs with Starlark line numbers if possible (default: false)
-l	--line-numbers, -l	print logs with Starlark line numbers if possible (default: false)
--no-cache	--no-cache, -n	ignore stored values in the cache (default: false)
-n	--no-cache, -n	ignore stored values in the cache (default: false)
--verbose	--verbose, -v	print logs to stdout (default: false)
-v	--verbose, -v	print logs to stdout (default: false)
--help	--help, -h	show help
-h	--help, -h	show help
`)+"\n", joined)
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
        bool_flag("testflag", default=False, usage="Testing flag")
    ]
)
`,
			sindrtest.WithArgs("__complete", "test", "-"),
			sindrtest.WithWriter(writer))

		require.NotEmpty(t, writer.Writes)
		var verboseFlag string
		for _, w := range writer.Writes {
			if strings.Contains(w, "--testflag") {
				verboseFlag = w
				break
			}
		}
		assert.NotEmpty(t, verboseFlag)
		assert.Equal(t, "--testflag\t--testflag\tTesting flag (default: false)\n", verboseFlag)
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
sub_command(path=["deploy", "staging"], action=deploy_staging, usage="Deploy to staging")
sub_command(path=["deploy", "production"], action=deploy_prod, usage="Deploy to production")
`,
			sindrtest.WithArgs("__complete", "deploy", "--", ""),
			sindrtest.WithWriter(writer))

		require.NotEmpty(t, writer.Writes)
		var stagingCommand, productionCommand string
		for _, w := range writer.Writes {
			if strings.Contains(w, "staging") {
				stagingCommand = w
			}
			if strings.Contains(w, "production") {
				productionCommand = w
			}
		}
		assert.NotEmpty(t, stagingCommand)
		assert.NotEmpty(t, productionCommand)
		assert.Equal(t, "staging\tDeploy to staging\n", stagingCommand)
		assert.Equal(t, "production\tDeploy to production\n", productionCommand)
	})
}
