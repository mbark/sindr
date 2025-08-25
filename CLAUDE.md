# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`shmake` is a replacement for `make` that allows creating CLI tools written in Starlark and run via a single Go binary. It
provides a batteries-included set of functions for common development tasks like shell commands, async operations, and
string templating.

## Architecture

The project consists of:

- **Go Runtime**: Core engine that interprets and executes Starlark scripts (`run_starlark.go`, `command.go`)
- **Starlark Integration**: Uses `go.starlark.net` interpreter with custom builtin functions
- **CLI Framework**: Built on `urfave/cli/v3` for command parsing and help generation
- **Configuration**: Projects use `main.star` files to define commands and actions

## Common Commands

### Building

```bash
go build -o sindr cmd/main.go
```

### Running

```bash
# Run the built binary
./sindr [command] [flags]

# Or run directly with Go
go run cmd/main.go [command] [flags]
```

### Development

```bash
# Build and test with example main.star
go build -o sindr cmd/main.go && ./sindr build

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
golangci-lint fmt --config .golangci.yml
golangci-lint run --config .golangci.yml
```

The linter configuration ensures code quality and security best practices are followed.

## Key Components

### Runtime System

- Starlark interpreter executes `main.star` configuration files
- Global variables defined in Starlark are accessible for template expansion
- Built-in functions are exposed through the `shmake` module

### Command System

- Commands defined in Starlark using `command()` function calls
- Supports nested subcommands via `sub_command()` with path arrays
- Flags, arguments, and actions are defined as function parameters

### Available Functions

- `cli()`: Configure the CLI tool name and usage
- `command()`: Define top-level commands with name, help, action, args, and flags
- `sub_command()`: Define nested subcommands with path arrays
- `shell()`: Execute shell commands with optional prefixes and return a structured result containing stdout, stderr, exit code, and success status
- `exec()`: Execute commands with a specific binary/interpreter by writing the command to a temporary file
- `dotenv()`: Load environment variables from .env files with optional overloading
- `start()`: Run functions concurrently
- `wait()`: Wait for async operations to complete
- `pool()`: Manage groups of concurrent operations
- `string()`: Render string templates with Go template syntax
- `cache()`: Create a cache instance for version management and caching operations

### Cache Functions

The cache system has been refactored to use cache instances created with `cache()`. Each cache instance provides:

- `cache.diff(name, version)`: Check if the current version differs from the stored version
- `cache.get_version(name)`: Retrieve the stored version for a given name
- `cache.set_version(name, version)`: Store a version for a given name
- `cache.with_version(function, name, version)`: Execute a function only if the version has changed

Cache instances can optionally specify a custom cache directory:
```python
c = cache()  # Uses default cache directory
c = cache(cache_dir="/custom/path")  # Uses custom cache directory
```

### Template System

- String templates support Go template syntax with `{{.variable}}`
- Global Starlark variables are automatically available in templates
- `string()` function renders templates with optional additional context

### Shell Command Results

The `shell()` function returns a structured result object with the following attributes:

- `result.stdout`: String containing the command's standard output (trimmed of surrounding whitespace)
- `result.stderr`: String containing the command's standard error output (trimmed of surrounding whitespace)
- `result.exit_code`: Integer exit code returned by the command
- `result.success`: Boolean indicating whether the command succeeded (exit code 0)

The result object can be used directly as a string (returns stdout) or in boolean context (returns success status):

```python
# Access specific outputs
result = shell('echo "hello" && echo "error" >&2')
print("Output:", result.stdout)  # "hello"
print("Error:", result.stderr)   # "error"
print("Success:", result.success) # True
print("Exit code:", result.exit_code) # 0

# Use as string (returns stdout)
output = shell('echo "hello world"')
print(str(output))  # "hello world"

# Use in boolean context (returns success)
if shell('test -f myfile.txt'):
    print("File exists")
else:
    print("File does not exist")
```

### Exec Function

The `exec()` function allows executing commands with a specific binary or interpreter. It writes the command content to a temporary file and executes it with the specified binary. The function signature is:

```python
exec(bin, command, args=None, no_output=False, prefix="")
```

**Parameters:**
- `bin` (required): The binary/interpreter to use (e.g., "python3", "sh", "awk")
- `command` (required): The command content to execute
- `args` (optional): Additional arguments to pass to the binary (currently unused)
- `no_output` (optional): If True, suppress stdout/stderr capture
- `prefix` (optional): Logging prefix for the command

The function returns the same structured result object as `shell()` with stdout, stderr, exit_code, and success attributes.

```python
# Execute Python code
result = exec(bin='python3', command='print("Hello from Python")')
print(result.stdout)  # "Hello from Python"

# Execute shell script
shell_script = '''
echo "Processing files..."
for file in *.txt; do
    echo "Found: $file"
done
'''
result = exec(bin='sh', command=shell_script)

# Execute with different interpreters
awk_result = exec(bin='awk', command='BEGIN { print "Hello from AWK" }')

# Handle multiline Python scripts
python_code = '''
import sys
import os

def main():
    print(f"Python version: {sys.version}")
    print(f"Current directory: {os.getcwd()}")

if __name__ == "__main__":
    main()
'''
result = exec(bin='python3', command=python_code)

# Use with prefix and no_output
exec(bin='sh', command='echo "Building..."', prefix='BUILD', no_output=True)
```

### Dotenv Function

The `dotenv()` function loads environment variables from .env files into the current process. It uses the `godotenv` library to parse files and can handle multiple files and overloading behavior. The function signature is:

```python
dotenv(files=None, overload=False)
```

**Parameters:**
- `files` (optional): List of .env file paths to load. If not specified, defaults to `[".env"]`
- `overload` (optional): If True, override existing environment variables. If False (default), skip variables that are already set

**Behavior:**
- By default, loads `.env` from the current directory
- Existing environment variables are preserved unless `overload=True`
- Variables are set in the current process and available to subsequent shell commands
- Supports standard .env file formats including comments, quotes, and empty values
- Logs the loading process and provides verbose output about exported, overloaded, and skipped variables

**Return Value:**
- Returns `None` (Starlark equivalent)

```python
# Load default .env file
dotenv()

# Load specific files
dotenv(['.env.local', '.env.production'])

# Load with overloading (override existing variables)
dotenv(overload=True)

# Load multiple files with overloading
dotenv(['.env', '.env.local'], overload=True)

# Example usage in a command
def deploy(ctx):
    # Load environment configuration
    dotenv(['.env', '.env.production'])
    
    # Use environment variables in shell commands
    result = shell('echo "Deploying to $DEPLOY_ENV"')
    shell('docker build -t $IMAGE_NAME:$VERSION .')
```

**Supported .env File Format:**
```bash
# Comments are ignored
DATABASE_URL=postgres://localhost:5432/mydb
API_KEY=secret-key-here
DEBUG=true

# Quoted values
APP_NAME="My Application"
DESCRIPTION='A sample app'

# Empty values
OPTIONAL_VAR=

# Values with equals signs
CONFIG_JSON={"key":"value","nested":{"setting":"enabled"}}
```

## Project Structure

- `cmd/main.go`: Entry point that calls `Run()`
- `internal/run.go`: Main runtime and Starlark integration
- `internal/command.go`: CLI command building and Starlark integration
- `internal/shell.go`: Shell execution with async support
- `internal/exec.go`: Binary/interpreter execution with temporary files
- `internal/dotenv.go`: Environment variable loading from .env files
- `internal/strings.go`: String templating system
- `internal/helpers_test.go`: Test helper functions for consistent test setup
- `cache/cache.go`: Caching system for expensive operations
- `main.star`: Example/development configuration file

## Development Notes

- The project looks for `main.star` in the current directory or parent directories
- Logging goes to stderr using structured JSON logging
- GoReleaser configuration exists for multi-platform builds
- Uses disk-based caching in user cache directory for cache operations

## Starlark Configuration Example

```python
# Define CLI metadata
cli(
    name = "shmake",
    usage = "A sample CLI tool"
)

# Define a command with arguments and flags
def build(ctx):
    print("building", ctx.args.target, "with flag", ctx.flags.some_flag)
    shell('echo "Building..."')

command(
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
sub_command(
    path = ["deploy", "staging"],
    action = lambda ctx: print("Deploying to staging")
)

# Use cache for version management
def deploy(ctx):
    c = cache()
    def build_task():
        shell('go build .')
        print('Built successfully')
    
    # Only rebuild if version changed
    if c.with_version(build_task, name='build', version='1.2.3'):
        print('Build executed')
    else:
        print('Build skipped - version unchanged')

# Use exec function for running scripts with specific interpreters
def setup(ctx):
    # Run Python setup script
    python_setup = '''
import json
import os

config = {
    "project": "shmake",
    "version": "1.0.0",
    "author": "developer"
}

with open("config.json", "w") as f:
    json.dump(config, f, indent=2)

print("Configuration file created")
'''
    result = exec(bin='python3', command=python_setup)
    if result.success:
        print("Setup completed successfully")

command(
    name = "setup", 
    help = "Set up project configuration",
    action = setup
)

# Use dotenv for environment variable management
def deploy(ctx):
    # Load environment configuration
    dotenv(['.env', '.env.production'])
    
    # Use loaded environment variables
    result = shell('echo "Deploying $PROJECT_NAME to $DEPLOY_ENV"')
    if result.success:
        shell('docker build -t $IMAGE_NAME:$VERSION .')
        shell('docker push $IMAGE_NAME:$VERSION')
        print("Deployment completed successfully")

command(
    name = "deploy",
    help = "Deploy application with environment configuration", 
    action = deploy
)
```

## Linting Guidelines

- You're not allowed to modify the golangci-lint to remove linters unless specifically requested to do so.

## Linting Best Practices

- Before running golangci-lint run, always run golangci-lint fmt to ensure the files are formatted.

## Testing Guidelines

All tests must follow this consistent pattern:

### Test Structure Pattern

1. **Use helper functions from `internal/helpers_test.go`**:
   - `sindrtest.SetupStarlarkRuntime(t)` - Creates runtime with temp directory and returns a run function
   - `sindrtest.WithMainStar(t, starlarkCode)` - Creates main.star file with Starlark test code

2. **Test function structure**:
   ```go
   func TestFunctionName(t *testing.T) {
       t.Run("test case description", func(t *testing.T) {
           run := sindrtest.SetupStarlarkRuntime(t)
           sindrtest.WithMainStar(t, `
   def test_action(ctx):
       result = function_name('args')
       if result != 'expected':
           fail('expected "expected", got: ' + str(result))
   
   cli(name="TestName")
   command(name="test", action=test_action)
   `)
           run()
       })
   }
   ```

3. **Required elements**:
   - Package: `package internal_test` for tests in the internal package
   - Use `sindrtest.SetupStarlarkRuntime(t)` for consistent test environment
   - Use `sindrtest.WithMainStar(t, starlarkCode)` for Starlark script setup
   - Call the returned `run()` function to execute the test
   - Test command name should match the Go test function
   - Use descriptive sub-test names with `t.Run()`

4. **Starlark test patterns**:
   - Define test functions that use `shmake` module functions
   - Create CLI with `cli(name="TestName")`
   - Add test command with `command(name="test", action=test_function)`
   - Use Starlark `fail()` for test failures with descriptive messages
   - Access context via `ctx` parameter for args and flags

This pattern ensures consistent test structure, proper cleanup, and reliable test execution across all test files.
