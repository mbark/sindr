package shmake

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"

	"github.com/mbark/shmake/cache"
	"github.com/mbark/shmake/internal"
	"github.com/mbark/shmake/internal/logger"
	"github.com/mbark/shmake/loader"
)

// StarlarkBuiltin exposes the expected function signature for a starlark builtin function. It's just added here to
// simplify adding additional Globals.
type StarlarkBuiltin = func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)

type runOptions struct {
	cacheDir string
	fileName string
	globals  starlark.StringDict
}

type RunOption func(*runOptions)

func WithCacheDir(dir string) RunOption {
	return func(o *runOptions) {
		o.cacheDir = dir
	}
}

func WithFileName(name string) RunOption {
	return func(o *runOptions) {
		o.fileName = name
	}
}

func WithGlobalValue(name string, value starlark.Value) RunOption {
	return func(o *runOptions) {
		o.globals[name] = value
	}
}

func WithBuiltin(name string, builtin StarlarkBuiltin) RunOption {
	return func(o *runOptions) {
		o.globals[name] = starlark.NewBuiltin(name, builtin)
	}
}

func Run(ctx context.Context, args []string, opts ...RunOption) error {
	options := runOptions{
		cacheDir: cacheHome(),
		fileName: "main.star",
		globals:  starlark.StringDict{},
	}
	for _, o := range opts {
		o(&options)
	}

	cache.SetCache(options.cacheDir)

	dir, err := findPathUpdwards(options.fileName)
	if err != nil {
		return err
	}

	err = os.Chdir(dir)
	if err != nil {
		return err
	}

	predeclared := createPredeclaredDict(dir)
	for name, value := range options.globals {
		predeclared[name] = value
	}

	loader.Predeclared = predeclared
	thread := &starlark.Thread{
		Name: "cli",
		Load: loader.Load,
		Print: func(thread *starlark.Thread, msg string) {
			logger.WithStack(thread.CallStack()).Log(msg)
		},
	}
	_, err = starlark.ExecFileOptions(
		&syntax.FileOptions{},
		thread,
		options.fileName,
		nil,
		predeclared,
	)
	if err != nil {
		return err
	}

	return runCLI(ctx, args)
}

func runCLI(ctx context.Context, args []string) error {
	var verbose, noCache, withLineNumbers bool
	cliFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:        "verbose",
			Usage:       "print logs to stdout",
			Destination: &verbose,
		},
		&cli.BoolFlag{
			Name:        "no-cache",
			Usage:       "ignore stored values in the cache",
			Destination: &noCache,
		},
		&cli.BoolFlag{
			Name:        "with-line-numbers",
			Usage:       "print logs with Starlark line numbers if possible",
			Destination: &withLineNumbers,
		},
	}

	cmd := internal.GlobalCLI.Command.Command
	cmd.Flags = append(cmd.Flags, cliFlags...)
	cmd.Before = func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		logger.DoLogVerbose = verbose
		logger.WithLineNumbers = withLineNumbers
		cache.GlobalCache.ForceOutOfDate = noCache
		return ctx, nil
	}

	err := cmd.Run(ctx, args)
	if err != nil {
		return err
	}

	internal.WaitGroup.Wait()
	return nil
}

// createPredeclaredDict creates the predeclared dictionary for Starlark execution.
func createPredeclaredDict(dir string) starlark.StringDict {
	return starlark.StringDict{
		"shmake": starlarkstruct.FromStringDict(starlark.String("shmake"), starlark.StringDict{
			"cli":         starlark.NewBuiltin("cli", internal.ShmakeCLI),
			"command":     starlark.NewBuiltin("command", internal.ShmakeCommand),
			"sub_command": starlark.NewBuiltin("sub_command", internal.ShmakeSubCommand),

			"dotenv": starlark.NewBuiltin("dotenv", internal.ShmakeDotenv),

			"shell": starlark.NewBuiltin("shell", internal.ShmakeShell),
			"exec":  starlark.NewBuiltin("exec", internal.ShmakeExec),

			"string": starlark.NewBuiltin("string", internal.ShmakeString),

			"start": starlark.NewBuiltin("start", internal.ShmakeStart),
			"wait":  starlark.NewBuiltin("wait", internal.ShmakeWait),
			"pool":  starlark.NewBuiltin("pool", internal.ShmakePool),

			"newest_ts": starlark.NewBuiltin("newest_ts", internal.ShmakeNewestTS),
			"oldest_ts": starlark.NewBuiltin("oldest_ts", internal.ShmakeOldestTS),
			"glob":      starlark.NewBuiltin("glob", internal.ShmakeGlob),

			"load_package_json": starlark.NewBuiltin(
				"load_package_json",
				internal.ShmakeLoadPackageJson,
			),
		}),
		"cache":       starlark.NewBuiltin("cache", cache.NewCacheValue),
		"current_dir": starlark.String("current_dir"),
	}
}
