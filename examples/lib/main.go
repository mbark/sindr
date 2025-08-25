package main

import (
	"context"
	_ "embed"
	"os"

	"go.starlark.net/starlark"

	"github.com/mbark/sindr"
	"github.com/mbark/sindr/internal/logger"
)

// This shows an example of using sindr as a library that can be extended with custom functions that you want to use.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := sindr.Run(ctx, os.Args,
		sindr.WithBuiltin("custom_function", CustomFunction),
		sindr.WithFileName("lib.star"),
	)
	if err != nil {
		logger.LogErr("error running sindr", err)
	}
}

var CustomFunction = func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.MakeInt(42), nil
}
