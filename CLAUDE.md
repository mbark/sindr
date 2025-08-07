# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`shmake` is a replacement for `make` that allows creating CLI tools written in Lua and run via a single Go binary. It
provides a batteries-included set of functions for common development tasks like file operations, shell commands,
watching files, and async operations.

## Architecture

The project consists of:

- **Go Runtime**: Core engine that interprets and executes Lua scripts (`shmake.go`, `types.go`)
- **Lua Modules**: Exposed Go functions available to Lua scripts
    - `shmake.main`: Core functions (commands, shell, async, templates, caching)
    - `shmake.files`: File operations (write, delete, copy, mkdir, chdir)
- **CLI Framework**: Built on `urfave/cli/v3` for command parsing and help generation
- **Lua Integration**: Uses `gopher-lua` interpreter with custom module system
- **Configuration**: Projects use `main.lua` files to define commands and actions

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
# Build and test with example main.lua
go build -o shmake cmd/main.go && ./shmake run

# Format code
go fmt ./...

# Download dependencies
go mod download
go mod tidy
```

### Testing

No test files exist currently. Tests would use standard Go testing:

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

- `Runtime` struct manages Lua state, modules, async operations, and caching
- Modules are registered with the Lua state via `PreloadModule`
- Global variables in Lua are accessible for template expansion

### Command System

- Commands defined in Lua using `shmake.command()` and chained methods
- Supports nested subcommands via `:command()` or `:sub_command()`
- Flags, arguments, and actions are defined fluently

### Module Functions

- `shell()`: Execute shell commands with optional prefixes and capture output
- `async()` and `wait()`: Run commands concurrently
- `watch()`: Monitor file changes and trigger actions
- `pool()`: Manage groups of concurrent operations
- `with_version()`: Cache operations based on file modification times
- File operations: write, delete, copy, mkdir, chdir, popdir

### Template System

- String templates support Go template syntax with `{{.variable}}`
- Global Lua variables are automatically available in templates
- `shmake.string()` function renders templates with optional additional context

## Project Structure

- `cmd/main.go`: Entry point that calls `shmake.Run()`
- `shmake.go`: Main runtime and module registration
- `command.go`: CLI command building and Lua integration
- `files.go`: File operation module functions
- `shell.go`: Shell execution with async support
- `watch.go`: File watching functionality
- `template.go`: String templating system
- `cache.go`: Caching system for expensive operations
- `types.go`: Lua type definitions and utilities
- `main.lua`: Example/development configuration file

## Development Notes

- The project looks for `main.lua` in the current directory or parent directories
- Logging goes to both stderr and a log file using structured JSON logging
- GoReleaser configuration exists for multi-platform builds
- Uses disk-based caching in user cache directory for `with_version` operations
