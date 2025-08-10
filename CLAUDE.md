# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`shmake` is a replacement for `make` that allows creating CLI tools written in Starlark and run via a single Go binary. It
provides a batteries-included set of functions for common development tasks like shell commands,
watching files, async operations, and string templating.

## Architecture

The project consists of:

- **Go Runtime**: Core engine that interprets and executes Starlark scripts (`run_starlark.go`, `command.go`)
- **Starlark Integration**: Uses `go.starlark.net` interpreter with custom builtin functions
- **CLI Framework**: Built on `urfave/cli/v3` for command parsing and help generation
- **Configuration**: Projects use `main.star` files to define commands and actions

## Common Commands

### Building

```bash
go build -o shmake cmd/main.go
```

### Running

```bash
# Run the built binary
./shmake [command] [flags]

# Or run directly with Go
go run cmd/main.go [command] [flags]
```

### Development

```bash
# Build and test with example main.star
go build -o shmake cmd/main.go && ./shmake build

# Format code
go fmt ./...

# Download dependencies
go mod download
go mod tidy
```

### Testing

The project includes unit tests for core functionality. Tests should be run to ensure code quality:

```bash
go test ./...
```

### Linting

The project uses `golangci-lint` for code quality checks. All code must pass linting before being committed:

```bash
golangci-lint run
```

The linter configuration ensures code quality and security best practices are followed.

## Key Components

### Runtime System

- Starlark interpreter executes `main.star` configuration files
- Global variables defined in Starlark are accessible for template expansion
- Built-in functions are exposed through the `shmake` module

### Command System

- Commands defined in Starlark using `shmake.command()` function calls
- Supports nested subcommands via `shmake.sub_command()` with path arrays
- Flags, arguments, and actions are defined as function parameters

### Available Functions

- `shmake.cli()`: Configure the CLI tool name and usage
- `shmake.command()`: Define top-level commands with name, help, action, args, and flags
- `shmake.sub_command()`: Define nested subcommands with path arrays
- `shmake.shell()`: Execute shell commands with optional prefixes and capture output
- `shmake.run_async()`: Run functions concurrently
- `shmake.wait()`: Wait for async operations to complete
- `shmake.watch()`: Monitor file changes and trigger actions
- `shmake.pool()`: Manage groups of concurrent operations
- `shmake.string()`: Render string templates with Go template syntax
- `shmake.with_version()`: Cache operations based on versions
- `shmake.store()` and `shmake.get_version()`: Version storage and retrieval
- `shmake.diff()`: Compare values for changes

### Template System

- String templates support Go template syntax with `{{.variable}}`
- Global Starlark variables are automatically available in templates
- `shmake.string()` function renders templates with optional additional context

## Project Structure

- `cmd/main.go`: Entry point that calls `shmake.RunStar()`
- `run_starlark.go`: Main runtime and Starlark integration
- `command.go`: CLI command building and Starlark integration
- `shell.go`: Shell execution with async support
- `cache.go`: Caching system for expensive operations
- `strings.go`: String templating system
- `main.star`: Example/development configuration file

## Development Notes

- The project looks for `main.star` in the current directory or parent directories
- Logging goes to stderr using structured JSON logging
- GoReleaser configuration exists for multi-platform builds
- Uses disk-based caching in user cache directory for `with_version` operations

## Starlark Configuration Example

```python
# Define CLI metadata
shmake.cli(
    name = "shmake",
    usage = "A sample CLI tool"
)

# Define a command with arguments and flags
def build(ctx):
    print("building", ctx.args.target, "with flag", ctx.flags.some_flag)
    shmake.shell('echo "Building..."')

shmake.command(
    name = "build",
    help = "Build the project", 
    action = build,
    args = ["target"],
    flags = {
        "some-flag": {
            "type": "bool",
            "default": False,
        }
    }
)

# Define subcommands using path arrays
shmake.sub_command(
    path = ["deploy", "staging"],
    action = lambda ctx: print("Deploying to staging")
)
```

## Linting Guidelines

- You're not allowed to modify the golangci-lint to remove linters unless specifically requested to do so.

## Linting Best Practices

- Before running golangci-lint run, always run golangci-lint fmt to ensure the files are formatted.

## Testing Guidelines

All tests must follow this consistent pattern:

### Test Structure Pattern

1. **Use helper functions from `helpers_test.go`**:
   - `setupStarlarkRuntime(t)` - Creates runtime with temp directory and returns a run function
   - `withMainStar(t, starlarkCode)` - Creates main.star file with Starlark test code

2. **Test function structure**:
   ```go
   func TestFunctionName(t *testing.T) {
       t.Run("test case description", func(t *testing.T) {
           run := setupStarlarkRuntime(t)
           withMainStar(t, `
   def test_action(ctx):
       result = shmake.function_name('args')
       if result != 'expected':
           fail('expected "expected", got: ' + str(result))
   
   shmake.cli(name="TestName")
   shmake.command(name="test", action=test_action)
   `)
           run()
       })
   }
   ```

3. **Required elements**:
   - Package: `package shmake_test`  
   - Use `setupStarlarkRuntime(t)` for consistent test environment
   - Use `withMainStar(t, starlarkCode)` for Starlark script setup
   - Call the returned `run()` function to execute the test
   - Test command name should match the Go test function
   - Use descriptive sub-test names with `t.Run()`

4. **Starlark test patterns**:
   - Define test functions that use `shmake` module functions
   - Create CLI with `shmake.cli(name="TestName")`
   - Add test command with `shmake.command(name="test", action=test_function)`
   - Use Starlark `fail()` for test failures with descriptive messages
   - Access context via `ctx` parameter for args and flags

This pattern ensures consistent test structure, proper cleanup, and reliable test execution across all test files.