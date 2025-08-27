<h1 align=center><code>🔨✨ sindr</code></h1>

[![tag](https://img.shields.io/github/tag/mbark/sindr.svg)](https://github.com/mbark/sindr/releases)
[![GoDoc](https://godoc.org/github.com/mbark/sindr?status.svg)](https://pkg.go.dev/github.com/mbark/sindr)
[![Build Status](https://github.com/mbark/sindr/actions/workflows/lint.yml/badge.svg)](https://github.com/mbark/sindr/actions/workflows/lint.yml)
[![Go report](https://goreportcard.com/badge/github.com/mbark/sindr)](https://goreportcard.com/report/github.com/mbark/sindr)
[![Coverage](https://img.shields.io/codecov/c/github/mbark/sindr)](https://codecov.io/gh/mbark/sindr)
[![Contributors](https://img.shields.io/github/contributors/mbark/sindr)](https://github.com/mbark/sindr/graphs/contributors)
[![License](https://img.shields.io/github/license/mbark/sindr)](./LICENSE)

`sindr` ([from Old Norse: "slag or dross from a forge"](https://cleasby-vigfusson-dictionary.vercel.app/word/sindr)) is
a simple way to create a CLI to save and run project-specific commands.

Using [Starlark](https://github.com/bazelbuild/starlark), a Python-subset, you can create a CLI with flags, arguments
and autocompletion to save and run project-specific commands.

<img src="carbon.png">

Configure your CLI in `Starlark`, using builtin functions like `command` and `shell` to easily create a CLI for your
project. Then run `sindr` to get a CLI for your project-specific commands.

```
sindr test -- -race
test
  Flags  
    short: true    
  Named arguments  
    args: '-race'    
$ go test -short -race ./...
[...]
```

`sindr` has several useful features:

- `sindr` generates a fully-fledged CLI giving your developers a familiar interface for how to run scripts.
- `Starlark` is a Python-subset which makes configuration simple and familiar for developers.
- Building with `go` gives us a single binary with no external dependencies to run.
- Error messages are clear and if relevant, point to the exact line and column in `Starlark`.
- `sindr` can be invoked from any subdirectory, not just the one with `sindr.star`.
- `sindr` can load `.env` files, making it easy to populate environment variables.
- allows string-expansion using `golang` templates.
- has a builtin file-based cache system that can be used to check if some command should be run.
- several other builtin functions to do common tasks like checking if any files have been updated.
- allows executing arbitrary languages, like Python or NodeJS.

## Installation

### Installation Script

You can also install `sindr` using the installation script:

```bash
curl -sSL https://github.com/mbark/sindr/raw/master/install.sh | sh
```

### Homebrew

For users of Homebrew, it's probably easiest to install via `sindr` it:

```bash
brew tap mbark/tap
brew install sindr
```

### Pre-built Binaries

You can find pre-built binaries for your platform at [GitHub Releases](https://github.com/mbark/sindr/releases/latest).

#### macOS

```bash
# Intel Mac
curl -sSL https://api.github.com/repos/mbark/sindr/releases/latest | \
  grep "browser_download_url.*darwin_amd64.tar.gz" | \
  cut -d '"' -f 4 | \
  xargs curl -L | tar xz && chmod +x sindr && sudo mv sindr /usr/local/bin/

# Apple Silicon Mac
curl -sSL https://api.github.com/repos/mbark/sindr/releases/latest | \
  grep "browser_download_url.*darwin_arm64.tar.gz" | \
  cut -d '"' -f 4 | \
  xargs curl -L | tar xz && chmod +x sindr && sudo mv sindr /usr/local/bin/
```

#### Linux

```bash
# 64-bit
curl -sSL https://api.github.com/repos/mbark/sindr/releases/latest | \
  grep "browser_download_url.*linux_amd64.tar.gz" | \
  cut -d '"' -f 4 | \
  xargs curl -L | tar xz && chmod +x sindr && sudo mv sindr /usr/local/bin/

# ARM64
curl -sSL https://api.github.com/repos/mbark/sindr/releases/latest | \
  grep "browser_download_url.*linux_arm64.tar.gz" | \
  cut -d '"' -f 4 | \
  xargs curl -L | tar xz && chmod +x sindr && sudo mv sindr /usr/local/bin/
```

### Using Go

If you have Go installed and prefer using that you can either `go install` with:

```bash
go install github.com/mbark/sindr@latest
```

When using Go 1.24+ you can also add it as a tool for your project:

```bash
go get -tool github.com/mbark/sindr
go mod tidy
go tool sindr
```

## Quick start

Create a file named `sindr.star` in the root of your project.

```starlark
cli(
    name = "cli_name",
    usage = "some usage text"
)

def a_command(ctx):
    res = shell(string('echo "{{.text}}"',text=ctx.flags.text))
    print(res.stdout)

command(
    name = "a_command",
    action = a_command,
    flags = {
        "text": {
            "type": "string",
            "default": "hello from sindr",
            "help": "text to echo"
        },
    },
)
```

Then invoke `sindr`, it will look in the current directory and upwards for a `sindr.star` file.

When invoking `sindr` with no arguments it will by default show the equivalent of running with `--help`.

```console
$ sindr
NAME:
   cli_name - some usage text

USAGE:
   cli_name [global options] [command [command options]]

COMMANDS:
   a_command  
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --verbose            print logs to stdout (default: false)
   --no-cache           ignore stored values in the cache (default: false)
   --with-line-numbers  print logs with Starlark line numbers if possible (default: false)
   --help, -h           show help
```

This will behave exactly as a CLI, meaning to run the command `a_command` just invoke it with that as the argument.

```console
sindr a_command
a_command
  Flags  
    text: 'hello from sindr'    
$ echo "hello from sindr"
  "hello from sindr"  
"hello from sindr"
```

When running a command `sindr` will log the name of the command, with flags and arguments if any are defined. In this
case none are so it will log simply `a_command`.

## Examples

A variety of examples can be found in the [examples directory](https://github.com/mbark/sindr/tree/master/examples).

### Adding commands and subcommands

* `cli`
* `command`
* `sub_command`

### Shell commands

* `shell`

### String templating

* `string`

### Working with asynchronous commands

* `start`
* `wait`
* `pool`

### Working with files

* `newest_ts`
* `oldest_ts`
* `glob`

### Using the cache

* `cache`

### Running other programming languages

* `exec`

### Importing scripts from `pacage.json`

* `load_package_json`

### Sourcing .env files with `dotenv`

* `dotenv`

### Working with `sindr` as a Go-library

## Comparison to other popular tools

`sindr` was developed out of frustration with existing project-specific commands focusing not being great for the task (
`make`) or building on the somewhat arcane syntax (`just`). However, that doesn't mean those are great tools. This
section attempts to outline how `sindr` compares to other popular tools for running commands.

### make

`make` and `Makefiles` are primarily built for (as the name indicates) building code. The `Makefile` format specifies
what files should be built and what files they depend on. This makes it great for building code but not as good when you
primarily use it for running commands. When using `Makefiles` for commands you can't really require passing arguments or
flags and if you want some simple scripting logic like `if X then Y` the syntax looks arcane.

Consider using `make` if you want to build code (typically C or C++) and maybe have only a few (`.PHONY`)
command-targets, otherwise a tool specifically for scripts is a better idea.

## Just

`just` is more or less `make` but optimized for project-specific commands. `just` is a really nice tool but also suffers
from having borrowed much syntax from `Makefiles`. In many ways it feels like how you wish `make` worked but it doesn't
really re-think the paradigm.

If you're already familiar with `Makefiles` or `Just` it's probably the better tool to choose. However, if you're not
either `sindr` allows writing your commands with a subset of Python and you get a CLI – a tool that most developers are
already very familiar with how it should be used.

## package.json

When working with web or `node` you can use the `script` section in your `package.json` to store common commands to run.

Scripts in `package.json` files however, are primarily built around having simple one-liner calls, not writing more
complex things. If you only have some simple things to do like `prettier --write' or similar, it works great. However,
if you need to do more complex logic, you have to create a separate file and write the script there.

`package.json` is good when you're already working with `npm` (or similar), have simple one-liners or are fine with
creating new files for each command.

## Building your own CLI

Pretty much every programming language has some great library for writing your CLI and this is definitely a nice way to
do it. It has several upsides: you can use the language the project is already in, you don't need to add any new tooling
and you can even re-use some of your other parts of your code base.

However, it can easily feel quite heavyhanded to have to create your own CLI just to run scripts. Additionally, some
programming languages are not ideal for scripting and the scripts can become quite cumbersome.

Using a tool developer specifically for running commands – like `sindr` – can be a good way to allow creating a CLI
a bit more easily. Should you also want to add more complex commands, you can use your own file for those. And finally,
it's not going to be too hard to migrate from a simple command-runner to your own CLI should you choose to do so.

## Inspiration

- [`make`](https://www.gnu.org/software/make/manual/html_node/index.html) the original.
- [`just`](https://github.com/casey/just) like `make` but explicitly for running commands.

## Credit

`sindr` is primarily an idea built on the shoulder of giants, this is a callout to the great libraries that made this
possible.

- [`urfave/cli`](https://github.com/urfave/cli) does all the heavy lifting to build a modern CLI.
- [`charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss) colorizes and pads logs to make it look good.
- [`google/starlark-go`](https://github.com/google/starlark-go) parses and runs the `Starlark`.
