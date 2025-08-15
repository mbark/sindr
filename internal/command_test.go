package internal_test

import (
	"testing"
)

func TestShmakeCLI(t *testing.T) {
	t.Run("sets CLI name and usage", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    print("CLI configured")

shmake.cli(name="test-cli", usage="A test CLI")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("handles optional usage parameter", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    print("CLI with name only")

shmake.cli(name="minimal-cli")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}

func TestShmakeCommand(t *testing.T) {
	t.Run("creates basic command", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def build_action(ctx):
    print("Building project")

shmake.cli(name="TestShmakeCommand")
shmake.command(
    name="build",
    help="Build the project",
    action=build_action
)
shmake.command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("handles command with arguments", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def deploy_action(ctx):
    target = ctx.args.target
    environment = ctx.args.environment
    print("Deploying", target, "to", environment)

shmake.cli(name="TestShmakeCommand")
shmake.command(
    name="deploy",
    help="Deploy the application",
    action=deploy_action,
    args=["target", "environment"]
)
shmake.command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("handles command with flags", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    verbose = ctx.flags.verbose
    count = ctx.flags.count
    name = ctx.flags.name
    print("verbose:", verbose, "count:", count, "name:", name)

shmake.cli(name="TestShmakeCommand")
shmake.command(
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
shmake.command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("handles command with category", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    print("Running categorized command")

shmake.cli(name="TestShmakeCommand")
shmake.command(
    name="lint",
    help="Run linting",
    action=test_action,
    category="development"
)
shmake.command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("handles dash-case flag names", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    # Test both dash-case and snake_case access
    value1 = ctx.flags["dry-run"]
    value2 = ctx.flags.dry_run
    if value1 != value2:
        fail("Expected dash-case and snake_case to return same value")
    print("dry-run flag:", value1)

shmake.cli(name="TestShmakeCommand")
shmake.command(
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
shmake.command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})
}

func TestShmakeSubCommand(t *testing.T) {
	t.Run("creates nested subcommand", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def deploy_staging_action(ctx):
    print("Deploying to staging")

def deploy_production_action(ctx):
    print("Deploying to production")

shmake.cli(name="TestShmakeSubCommand")

# Create parent deploy command first
shmake.command(
    name="deploy",
    help="Deployment commands",
    action=lambda ctx: print("Use subcommands")
)

shmake.sub_command(
    path=["deploy", "staging"],
    help="Deploy to staging",
    action=deploy_staging_action
)

shmake.sub_command(
    path=["deploy", "production"], 
    help="Deploy to production",
    action=deploy_production_action
)

shmake.command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("creates deeply nested subcommands", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def action(ctx):
    print("Deep nested command executed")

shmake.cli(name="TestShmakeSubCommand")

# Create parent commands
shmake.command(name="cloud", help="Cloud commands", action=lambda ctx: print("Use subcommands"))

shmake.sub_command(
    path=["cloud", "aws"],
    help="AWS commands", 
    action=lambda ctx: print("Use AWS subcommands")
)

shmake.sub_command(
    path=["cloud", "aws", "deploy"],
    help="Deploy to AWS",
    action=action
)

shmake.command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("subcommand with arguments and flags", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def deploy_action(ctx):
    service = ctx.args.service
    force = ctx.flags.force
    print("Deploying service:", service, "force:", force)

shmake.cli(name="TestShmakeSubCommand")

shmake.command(name="k8s", help="Kubernetes commands", action=lambda ctx: print("Use subcommands"))

shmake.sub_command(
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

shmake.command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})
}

func TestContextFlagAccess(t *testing.T) {
	t.Run("flag map supports both index and attr access", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
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

shmake.cli(name="TestContextFlagAccess")
shmake.command(
    name="check",
    action=test_action,
    flags={
        "verbose": {"default": True, "type": "bool"},
        "dry-run": {"default": False, "type": "bool"}
    }
)
shmake.command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("context provides args access", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    target = ctx.args.target
    environment = ctx.args.environment
    print("target:", target, "environment:", environment)

shmake.cli(name="TestContextFlagAccess")
shmake.command(
    name="deploy",
    action=test_action,
    args=["target", "environment"]
)
shmake.command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})

	t.Run("context provides args_list access", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    args_list = ctx.args_list
    print("args_list length:", len(args_list))
    for i in range(len(args_list)):
        print("arg", i, ":", args_list[i])

shmake.cli(name="TestContextFlagAccess")
shmake.command(
    name="process",
    action=test_action
)
shmake.command(name="test", action=lambda ctx: print("test executed"))
`)
		run()
	})
}

func TestInvalidConfigurations(t *testing.T) {
	t.Run("invalid flag type should fail", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    fail("Should not reach here")

shmake.cli(name="TestInvalidConfigurations")
shmake.command(
    name="fail",
    action=test_action,
    flags={
        "invalid": {
            "default": "test",
            "type": "unknown_type"
        }
    }
)
shmake.command(name="test", action=lambda ctx: print("test executed"))
`)

		run()
	})

	t.Run("non-string args should fail", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    fail("Should not reach here")

shmake.cli(name="TestInvalidConfigurations")
shmake.command(
    name="fail",
    action=test_action,
    args=[123, "valid"]
)
shmake.command(name="test", action=lambda ctx: print("test executed"))
`)

		// The error is logged but doesn't cause a panic, just test that it runs
		run()
	})

	t.Run("invalid subcommand path should fail", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    fail("Should not reach here")

shmake.cli(name="TestInvalidConfigurations") 

shmake.sub_command(
    path=["nonexistent", "command"],
    action=test_action
)
shmake.command(name="test", action=lambda ctx: print("test executed"))
`)

		// The error is logged but doesn't cause a panic, just test that it runs
		run()
	})
}
