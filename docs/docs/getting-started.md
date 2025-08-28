---
sidebar_position: 2
title: Getting Started with Sindr
description: Learn how to use Sindr as a replacement for make with Starlark scripts
keywords: [sindr, make, starlark, cli, build tool]
---

# Getting Started with Sindr

Sindr is a modern replacement for `make` that allows you to create CLI tools using Starlark scripts. It provides a batteries-included set of functions for common development tasks.

Before getting started, make sure you have [installed Sindr](./installation.md) on your system.

## Quick Start

1. Create a `sindr.star` file in your project root:

```python
# Define CLI metadata
cli(
    name = "my-project",
    usage = "My awesome project CLI"
)

# Define a build command
def build_action(ctx):
    print("Building project...")
    result = shell('go build .')
    if result.success:
        print("Build completed successfully!")
    else:
        print("Build failed:", result.stderr)

command(
    name = "build",
    help = "Build the project",
    action = build_action
)
```

2. Run your command:

```bash
sindr build
```

## Core Concepts

### Commands
Define commands using the `command()` function:

```python
def my_action(ctx):
    print("Hello from", ctx.args.name)

command(
    name = "greet",
    help = "Greet someone",
    action = my_action,
    args = ["name"]
)
```

### Shell Execution
Execute shell commands with the `shell()` function:

```python
def deploy(ctx):
    result = shell('docker build -t myapp .')
    if result.success:
        shell('docker push myapp')
        print("Deployment successful!")
    else:
        print("Build failed:", result.stderr)
```

### Environment Variables
Load environment variables from `.env` files:

```python
def setup(ctx):
    dotenv(['.env', '.env.local'])
    shell('echo "Using database: $DATABASE_URL"')
```

### Caching
Cache expensive operations:

```python
def expensive_build(ctx):
    c = cache()
    
    def build_task():
        shell('npm run build')
        print('Build completed')
    
    # Only rebuild if version changed
    if c.with_version(build_task, name='build', version='1.2.3'):
        print('Build executed')
    else:
        print('Build skipped - version unchanged')
```

## Available Functions

| Function | Description |
|----------|-------------|
| `cli()` | Configure CLI metadata |
| `command()` | Define commands |
| `sub_command()` | Define nested subcommands |
| `shell()` | Execute shell commands |
| `exec()` | Execute with specific interpreter |
| `dotenv()` | Load environment variables |
| `cache()` | Create cache instance |
| `string()` | Render templates |
| `start()` | Run functions concurrently |
| `wait()` | Wait for async operations |

## Examples

### Building and Testing
```python
def test_action(ctx):
    result = shell('go test ./...')
    if not result.success:
        print("Tests failed!")
        return
    
    shell('go build -o myapp .')
    print("Build and test completed successfully!")

command(name="test", action=test_action, help="Run tests and build")
```

### Multi-stage Deployment
```python
sub_command(
    path = ["deploy", "staging"],
    help = "Deploy to staging environment",
    action = lambda ctx: deploy_to_env("staging")
)

sub_command(
    path = ["deploy", "production"],
    help = "Deploy to production environment", 
    action = lambda ctx: deploy_to_env("production")
)

def deploy_to_env(env):
    dotenv([f'.env.{env}'])
    shell(f'docker build -t myapp:{env} .')
    shell(f'docker push myapp:{env}')
```

## Next Steps

- Explore the [API Reference](./api-reference.md)
- Check out [Examples](./examples.md)
- Learn about [Best Practices](./best-practices.md)