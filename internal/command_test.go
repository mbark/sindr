package internal_test

import (
	"testing"

	"go.starlark.net/starlark"

	"github.com/mbark/sindr/internal"
	"github.com/mbark/sindr/internal/sindrtest"
)

func TestSindrCLI(t *testing.T) {
	t.Run("sets CLI name and usage", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("CLI configured")

cli(name="test-cli", usage="A test CLI")
command(name="test", action=test_action)
`)
	})

	t.Run("handles optional usage parameter", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("CLI with name only")

cli(name="minimal-cli")
command(name="test", action=test_action)
`)
	})
}

func TestSindrCommand(t *testing.T) {
	t.Run("creates basic command", func(t *testing.T) {
		sindrtest.Test(t, `
def build_action(ctx):
    print("Building project")

cli(name="TestSindrCommand")
command(
    name="build",
    usage="Build the project",
    action=build_action
)
command(name="test", action=lambda ctx: print("test executed"))
`)
	})

	t.Run("handles command with arguments", func(t *testing.T) {
		sindrtest.Test(t, `
def deploy_action(ctx):
    target = ctx.args.target
    environment = ctx.args.environment
	version = ctx.args.version

	assert_equals(target, "test", "Expected target to be 'test'")
	assert_equals(environment, "prod", "Expected environment to be 'prod'")
	assert_equals(version, 1, "Expected version to be 1")

cli(name="TestSindrCommand")
command(
    name="test",
    usage="Deploy the application",
    action=deploy_action,
    args=[string_arg("target"), string_arg("environment"), int_arg("version")]
)
`, sindrtest.WithArgs("test", "test", "prod", "1"))
	})

	t.Run("handles command with flags", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    verbose = ctx.flags.verbose
    count = ctx.flags.count
    name = ctx.flags.name
    print("verbose:", verbose, "count:", count, "name:", name)

cli(name="TestSindrCommand")
command(
    name="test",
    usage="Run with flags",
    action=test_action,
    flags=[
		bool_flag('verbose',usage="Enable verbose output"),
		int_flag('count',usage="Number of iterations"),
		string_flag('name',usage="Name parameter")
    ]
)
`)
	})

	t.Run("handles command with strings flag type", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    strings_list = ctx.flags.strings
    assert_equals("list", type(strings_list), "Expected strings flag to be a list")
    assert_equals(2, len(strings_list), "Expected strings flag to have 2 items")
    assert_equals("default1", strings_list[0], "Expected first string to be default1")
    assert_equals("default2", strings_list[1], "Expected second string to be default2")
    print("strings flag test passed:", strings_list)

cli(name="TestSindrCommand")
command(
    name="test",
    usage="Test strings flag type",
    action=test_action,
    flags=[
        string_slice_flag('strings', default=["default1", "default2"], usage="List of strings")
    ]
)
`)
	})

	t.Run("handles command with ints flag type", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    ints_list = ctx.flags.ints
    assert_equals("list", type(ints_list), "Expected ints flag to be a list")
    assert_equals(3, len(ints_list), "Expected ints flag to have 3 items")
    assert_equals(10, ints_list[0], "Expected first int to be 10")
    assert_equals(20, ints_list[1], "Expected second int to be 20")
    assert_equals(30, ints_list[2], "Expected third int to be 30")
    print("ints flag test passed:", ints_list)

cli(name="TestSindrCommand")
command(
    name="test",
    usage="Test ints flag type",
    action=test_action,
    flags=[
        int_slice_flag('ints', default=[10, 20, 30], usage="List of integers")
    ]
)
`)
	})

	t.Run("handles command with category", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    print("Running categorized command")

cli(name="TestSindrCommand")
command(
    name="test",
    usage="Run linting",
    action=test_action,
    category="development"
)
`)
	})

	t.Run("handles dash-case flag names", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test both dash-case and snake_case access
    value1 = ctx.flags["dry-run"]
    value2 = ctx.flags.dry_run
    assert_equals(value1, value2, "Expected dash-case and snake_case to return same value")
    print("dry-run flag:", value1)

cli(name="TestSindrCommand")
command(
    name="test",
    usage="Deploy with dash-case flag",
    action=test_action,
    flags=[
        bool_flag('dry-run', default=True)
    ]
)
`)
	})
}

func TestSindrSubCommand(t *testing.T) {
	t.Run("creates nested subcommand", func(t *testing.T) {
		sindrtest.Test(t, `
def deploy_staging_action(ctx):
    print("Deploying to staging")

def deploy_production_action(ctx):
    print("Deploying to production")

cli(name="TestSindrSubCommand")

# Create parent deploy command first
command(
    name="test",
    usage="Deployment commands",
    action=lambda ctx: print("Use subcommands")
)

sub_command(
    path=["test", "staging"],
    usage="Deploy to staging",
    action=deploy_staging_action
)

sub_command(
    path=["test", "production"], 
    usage="Deploy to production",
    action=deploy_production_action
)
`, sindrtest.WithArgs("test", "staging"))
	})

	t.Run("creates deeply nested subcommands", func(t *testing.T) {
		sindrtest.Test(t, `
def action(ctx):
    print("Deep nested command executed")

cli(name="TestSindrSubCommand")

# Create parent commands
command(name="cloud", usage="Cloud commands", action=lambda ctx: print("Use subcommands"))

sub_command(
    path=["cloud", "aws"],
    usage="AWS commands", 
    action=lambda ctx: print("Use AWS subcommands")
)

sub_command(
    path=["cloud", "aws", "deploy"],
    usage="Deploy to AWS",
    action=action
)
`, sindrtest.WithArgs("cloud", "aws", "deploy"))
	})

	t.Run("subcommand with arguments and flags", func(t *testing.T) {
		sindrtest.Test(t, `
def deploy_action(ctx):
    service = ctx.args.service
    force = ctx.flags.force
    print("Deploying service:", service, "force:", force)

cli(name="TestSindrSubCommand")

command(name="k8s", usage="Kubernetes commands", action=lambda ctx: print("Use subcommands"))

sub_command(
    path=["k8s", "deploy"],
    usage="Deploy to Kubernetes",
    action=deploy_action,
    args=[string_arg("service")],
    flags=[
        bool_flag('force', default=False, usage="Force deployment")
    ]
)

command(name="k8s deploy", action=lambda ctx: print("test executed"))
`, sindrtest.WithArgs("k8s", "deploy", "service1", "--force"))
	})
}

func TestContextFlagAccess(t *testing.T) {
	t.Run("flag map supports both index and attr access", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test index access
    verbose_index = ctx.flags["verbose"]
    dry_run_index = ctx.flags["dry-run"]
    
    # Test attr access  
    verbose_attr = ctx.flags.verbose
    dry_run_attr = ctx.flags.dry_run
    
	assert_equals(verbose_index, verbose_attr, "verbose flag access mismatch")
	assert_equals(dry_run_index, dry_run_attr, "dry-run flag access mismatch")

    print("Flag access works correctly")

cli(name="TestContextFlagAccess")
command(
    name="test",
    action=test_action,
    flags=[
        bool_flag('verbose', default=True),
        bool_flag('dry-run', default=False)
    ]
)
`)
	})

	t.Run("context provides args access", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    target = ctx.args.target
    environment = ctx.args.environment
    print("target:", target, "environment:", environment)

cli(name="TestContextFlagAccess")
command(
    name="test",
    action=test_action,
    args=[string_arg("target"), string_arg("environment")]
)
`)
	})

	t.Run("context provides args_list access", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    args_list = ctx.args_list
    print("args_list length:", len(args_list))
    for i in range(len(args_list)):
        print("arg", i, ":", args_list[i])

cli(name="TestContextFlagAccess")
command(
    name="test",
    action=test_action
)
`)
	})

	t.Run("direct context attribute access for flags", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test direct flag access without going through ctx.flags
    verbose = ctx.verbose
    dry_run = ctx.dry_run
    count = ctx.count
    name = ctx.name
    
    # Verify values match expected defaults
    assert_equals(True, verbose, "verbose flag should be True")
    assert_equals(False, dry_run, "dry_run flag should be False")
    assert_equals(10, count, "count flag should be 10")
    assert_equals("default", name, "name flag should be 'default'")
        
    print("Direct flag access works correctly")

cli(name="TestDirectContextAccess")
command(
    name="test",
    action=test_action,
    flags=[
        bool_flag('verbose', default=True),
        bool_flag('dry-run', default=False),
        int_flag('count', default=10),
        string_flag('name', default="default")
    ]
)
`)
	})

	t.Run("direct context attribute access for args", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test direct arg access without going through ctx.args
    target = ctx.target
    environment = ctx.environment
    
    print("Direct arg access - target:", target, "environment:", environment)
    
    # Verify these match the traditional ctx.args access
    assert_equals(ctx.args.target, target, "ctx.target should match ctx.args.target")
    assert_equals(ctx.args.environment, environment, "ctx.environment should match ctx.args.environment")

cli(name="TestDirectContextAccess") 
command(
    name="test",
    action=test_action,
    args=[string_arg("target"), string_arg("environment")]
)
`)
	})

	t.Run("mixed direct and nested context access", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test mixing direct access with traditional nested access
    
    # Direct access
    verbose = ctx.verbose
    target = ctx.target
    
    # Traditional nested access
    verbose_nested = ctx.flags.verbose
    target_nested = ctx.args.target
    
    # Index access
    verbose_index = ctx.flags["verbose"]
    target_index = ctx.args["target"]
    
    # All should be equivalent
    assert_equals(verbose_nested, verbose, "verbose direct and nested access should match")
    assert_equals(verbose_index, verbose, "verbose direct and index access should match")
    assert_equals(target_nested, target, "target direct and nested access should match")
    assert_equals(target_index, target, "target direct and index access should match")
        
    print("Mixed access patterns work correctly")

cli(name="TestMixedContextAccess")
command(
    name="test",
    action=test_action,
    flags=[bool_flag('verbose', default=True)],
    args=[string_arg("target")]
)
`)
	})

	t.Run("direct access with snake_case conversion", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test that dash-case flags can be accessed via snake_case directly
    dry_run = ctx.dry_run
    api_key = ctx.api_key
    
    # Verify these match the nested access patterns
    assert_equals(ctx.flags.dry_run, dry_run, "ctx.dry_run should match ctx.flags.dry_run")
    assert_equals(ctx.flags.api_key, api_key, "ctx.api_key should match ctx.flags.api_key")
        
    # Also verify index access works with original dash-case names
    assert_equals(ctx.flags["dry-run"], dry_run, "ctx.dry_run should match ctx.flags['dry-run']")
    assert_equals(ctx.flags["api-key"], api_key, "ctx.api_key should match ctx.flags['api-key']")
        
    print("Snake case conversion works correctly")

cli(name="TestSnakeCaseConversion")
command(
    name="test",
    action=test_action,
    flags=[
        bool_flag('dry-run', default=False),
        string_flag('api-key', default="secret")
    ]
)
`)
	})

	t.Run("reserved attributes take precedence", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    # Test that reserved attributes (flags, args, args_list) are overridden if set.
    flags_obj = ctx.flags
    args_obj = ctx.args
    args_list_obj = ctx.args_list

    # These should be the container objects, not direct values
    if str(type(flags_obj)) == "FlagMap":
        fail("ctx.flags should not be FlagMap type")
    if str(type(args_obj)) == "FlagMap":
        fail("ctx.args should not be FlagMap type")
    if str(type(args_list_obj)) == "list":
        fail("ctx.args_list should not be list type")
        
    print("Reserved attributes work correctly")

cli(name="TestReservedAttributes")
command(
    name="test",
    action=test_action,
    flags=[string_flag('flags', default="should_not_override")],
    args=[string_arg("args"), string_arg("args_list")]
)
`)
	})
}

func TestInvalidConfigurations(t *testing.T) {
	t.Run("invalid flag type should fail", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    fail("Should not reach here")

cli(name="TestInvalidConfigurations")
command(
    name="test",
    action=test_action,
    flags=[
        unknown_type_flag('invalid', default="test")
    ]
)
`, sindrtest.ShouldFail())
	})

	t.Run("non-string args should fail", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    fail("Should not reach here")

cli(name="TestInvalidConfigurations")
command(
    name="test",
    action=test_action,
    args=[123, "valid"]
)
`, sindrtest.ShouldFail())
	})

	t.Run("invalid subcommand path should fail", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    fail("Should not reach here")

cli(name="TestInvalidConfigurations") 

sub_command(
    path=["nonexistent", "command"],
    action=test_action
)
command(name="test", action=lambda ctx: print("test executed"))
`, sindrtest.ShouldFail())
	})
}

func TestHelperStructs(t *testing.T) {
	sindrtest.AssertValue(t, internal.NewContext(
		starlark.StringDict{}, starlark.StringDict{}, starlark.NewList([]starlark.Value{}),
	), false)
	sindrtest.AssertValue(t, internal.NewFlagMap(starlark.StringDict{
		"foo": starlark.String("bar"),
	}), false)
	sindrtest.AssertValue(t, &internal.Arg{}, false)
	sindrtest.AssertValue(t, &internal.Flag{}, false)
}
