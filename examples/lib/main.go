package main

import (
	"context"
	_ "embed"
	"os"

	"go.starlark.net/starlark"

	"github.com/mbark/shmake"
)

// This shows an example of using shmake as a library that can be extended with custom functions that you want to use.

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shmake.Run(ctx, os.Args,
		shmake.WithBuiltin("custom_function", CustomFunction),
		shmake.WithFileName("lib.star"),
	)
}

var CustomFunction shmake.StarlarkBuiltin = func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.MakeInt(42), nil
}
