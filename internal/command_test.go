package internal_test

import (
	"testing"

	"github.com/mbark/sindr/internal/sindrtest"
)

func TestSindrCLI(t *testing.T) {
	t.Run("sets CLI name and usage", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    print("CLI configured")

cli(name="test-cli", usage="A test CLI")
command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles optional usage parameter", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    print("CLI with name only")

cli(name="minimal-cli")
command(name="test", action=test_action)
`)
		run()
	})
}

func TestSindrCommand(t *testing.T) {
	t.Run("creates basic command", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def build_action(ctx):
    print("Building project")

cli(name="TestSindrCommand")
command(
    name="build",
    help="Build the project",
    action=build_action
)
command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("handles command with arguments", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def deploy_action(ctx):
    target = ctx.args.target
    environment = ctx.args.environment
    print("Deploying", target, "to", environment)

cli(name="TestSindrCommand")
command(
    name="deploy",
    help="Deploy the application",
    action=deploy_action,
    args=["target", "environment"]
)
command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("handles command with flags", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    verbose = ctx.flags.verbose
    count = ctx.flags.count
    name = ctx.flags.name
    print("verbose:", verbose, "count:", count, "name:", name)

cli(name="TestSindrCommand")
command(
    name="run",
    help="Run with flags",
    action=test_action,
    flags={
        "verbose": {
            "default": False,
            "type": "bool",
            "help": "Enable verbose output"
        },
        "count": {
            "default": 1,
            "type": "int", 
            "help": "Number of iterations"
        },
        "name": {
            "default": "default",
            "type": "string",
            "help": "Name parameter"
        }
    }
)
command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("handles command with strings flag type", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    strings_list = ctx.flags.strings
    if type(strings_list) != "list":
        fail("Expected strings flag to be a list, got: " + str(type(strings_list)))
    if len(strings_list) != 2:
        fail("Expected strings flag to have 2 items, got: " + str(len(strings_list)))
    if strings_list[0] != "default1" or strings_list[1] != "default2":
        fail("Expected default strings ['default1', 'default2'], got: " + str(strings_list))
    print("strings flag test passed:", strings_list)

cli(name="TestSindrCommand")
command(
    name="strings-test",
    help="Test strings flag type",
    action=test_action,
    flags={
        "strings": {
            "default": ["default1", "default2"],
            "type": "strings",
            "help": "List of strings"
        }
    }
)
command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("handles command with ints flag type", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    ints_list = ctx.flags.ints
    if type(ints_list) != "list":
        fail("Expected ints flag to be a list, got: " + str(type(ints_list)))
    if len(ints_list) != 3:
        fail("Expected ints flag to have 3 items, got: " + str(len(ints_list)))
    if ints_list[0] != 10 or ints_list[1] != 20 or ints_list[2] != 30:
        fail("Expected default ints [10, 20, 30], got: " + str(ints_list))
    print("ints flag test passed:", ints_list)

cli(name="TestSindrCommand")
command(
    name="ints-test",
    help="Test ints flag type",
    action=test_action,
    flags={
        "ints": {
            "default": [10, 20, 30],
            "type": "ints",
            "help": "List of integers"
        }
    }
)
command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("handles command with category", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    print("Running categorized command")

cli(name="TestSindrCommand")
command(
    name="lint",
    help="Run linting",
    action=test_action,
    category="development"
)
command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("handles dash-case flag names", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    # Test both dash-case and snake_case access
    value1 = ctx.flags["dry-run"]
    value2 = ctx.flags.dry_run
    if value1 != value2:
        fail("Expected dash-case and snake_case to return same value")
    print("dry-run flag:", value1)

cli(name="TestSindrCommand")
command(
    name="deploy",
    help="Deploy with dash-case flag",
    action=test_action,
    flags={
        "dry-run": {
            "default": True,
            "type": "bool"
        }
    }
)
command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})
}

func TestSindrSubCommand(t *testing.T) {
	t.Run("creates nested subcommand", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def deploy_staging_action(ctx):
    print("Deploying to staging")

def deploy_production_action(ctx):
    print("Deploying to production")

cli(name="TestSindrSubCommand")

# Create parent deploy command first
command(
    name="deploy",
    help="Deployment commands",
    action=lambda ctx: print("Use subcommands")
)

sub_command(
    path=["deploy", "staging"],
    help="Deploy to staging",
    action=deploy_staging_action
)

sub_command(
    path=["deploy", "production"], 
    help="Deploy to production",
    action=deploy_production_action
)

command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("creates deeply nested subcommands", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def action(ctx):
    print("Deep nested command executed")

cli(name="TestSindrSubCommand")

# Create parent commands
command(name="cloud", help="Cloud commands", action=lambda ctx: print("Use subcommands"))

sub_command(
    path=["cloud", "aws"],
    help="AWS commands", 
    action=lambda ctx: print("Use AWS subcommands")
)

sub_command(
    path=["cloud", "aws", "deploy"],
    help="Deploy to AWS",
    action=action
)

command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("subcommand with arguments and flags", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def deploy_action(ctx):
    service = ctx.args.service
    force = ctx.flags.force
    print("Deploying service:", service, "force:", force)

cli(name="TestSindrSubCommand")

command(name="k8s", help="Kubernetes commands", action=lambda ctx: print("Use subcommands"))

sub_command(
    path=["k8s", "deploy"],
    help="Deploy to Kubernetes",
    action=deploy_action,
    args=["service"],
    flags={
        "force": {
            "default": False,
            "type": "bool",
            "help": "Force deployment"
        }
    }
)

command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})
}

func TestContextFlagAccess(t *testing.T) {
	t.Run("flag map supports both index and attr access", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    # Test index access
    verbose_index = ctx.flags["verbose"]
    dry_run_index = ctx.flags["dry-run"]
    
    # Test attr access  
    verbose_attr = ctx.flags.verbose
    dry_run_attr = ctx.flags.dry_run
    
    if verbose_index != verbose_attr:
        fail("verbose flag access mismatch")
    if dry_run_index != dry_run_attr:
        fail("dry-run flag access mismatch")
        
    print("Flag access works correctly")

cli(name="TestContextFlagAccess")
command(
    name="check",
    action=test_action,
    flags={
        "verbose": {"default": True, "type": "bool"},
        "dry-run": {"default": False, "type": "bool"}
    }
)
command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("context provides args access", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    target = ctx.args.target
    environment = ctx.args.environment
    print("target:", target, "environment:", environment)

cli(name="TestContextFlagAccess")
command(
    name="deploy",
    action=test_action,
    args=["target", "environment"]
)
command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("context provides args_list access", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    args_list = ctx.args_list
    print("args_list length:", len(args_list))
    for i in range(len(args_list)):
        print("arg", i, ":", args_list[i])

cli(name="TestContextFlagAccess")
command(
    name="process",
    action=test_action
)
command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})
}

func TestInvalidConfigurations(t *testing.T) {
	t.Run("invalid flag type should fail", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    fail("Should not reach here")

cli(name="TestInvalidConfigurations")
command(
    name="fail",
    action=test_action,
    flags={
        "invalid": {
            "default": "test",
            "type": "unknown_type"
        }
    }
)
command(name="test", action=lambda ctx: print("test executed"))
`)

		run(false)
	})

	t.Run("non-string args should fail", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    fail("Should not reach here")

cli(name="TestInvalidConfigurations")
command(
    name="fail",
    action=test_action,
    args=[123, "valid"]
)
command(name="test", action=lambda ctx: print("test executed"))
`)

		run(false)
	})

	t.Run("invalid subcommand path should fail", func(t *testing.T) {
		run := sindrtest.SetupStarlarkRuntime(t)
		sindrtest.WithMainStar(t, `
def test_action(ctx):
    fail("Should not reach here")

cli(name="TestInvalidConfigurations") 

sub_command(
    path=["nonexistent", "command"],
    action=test_action
)
command(name="test", action=lambda ctx: print("test executed"))
`)

		run(false)
	})
}
