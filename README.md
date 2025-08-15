# make shmake

`shmake` is a replacement for `make`, allowing the creation of a cli-tool be written using `Starlark` and run via a
single binary (built with `go`).

## Motivation

`make` often ends up being used as a way to write simple scripts that developers want to run often locally. This is not
ideal as `Makefiles` are weird to write, don't really play nice with configuration variables, nesting of commands, etc.
Additionally, `Makefiles` have their own syntax which most developers are not very familiar with.

## Goal

The goal of `shmake` is to allow writing a `.star` file with all script-targets with a batteries included set of
functions provided by `shmake` as modules. These `.star` files are then converted into a cli that allows targets,
sub-targets and the passing of flags and args.

## Design

Using a `Starlark` interpreter written in `go` we can ship a single binary which handles both interpreting and running
the `.star` file.

## TODOs and ideas

- [x] Allow "importing" package.json files and adding their scripts as commands
- [x] Add functionality for setting categories
- [x] Remove the watch stuff
- [ ] lib:ify the `shmake` package to allow people to use and extend it through Go.
- [ ] Add the newest_ts and oldest_ts functions back
- [ ] When running a command, show it in some nice way.
- [] Improve the tests for `packagejson` and command to actually check the commands are added (e.g., via --help)
