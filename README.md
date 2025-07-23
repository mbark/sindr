# make shmake

`shmake` is a replacement for `make`, allowing the creation of a cli-tool be written using `lua` and run via a single
binary (built with `go`).

## Motivation

`make` often ends up being used as a way to write simple scripts that developers want to run often locally. This is not
ideal as `Makefiles` are weird to write, don't really play nice with configuration variables, nesting of commands, etc.
Additionally, `Makefiles` have their own syntax which most developers are not very familiar with.

## Goal

The goal of `shmake` is to allow writing a `.lua` file with all script-targets with a batteries included set of
functions provided by `shmake` as modules. These `.lua` files are then converted into a cli that allows targets,
sub-targets and the passing of flags.

## Design

Using a `lua` interpreter written in `go` we can ship a single binary which handles both interpreting and running the
`lua` file.

With a set of modules provided to allow doing what you typically want to do e.g., watching files, copying and creating
files or running a command if something has changed since last time.
