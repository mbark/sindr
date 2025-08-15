package shmake

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/log"
	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/cli/v3"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"

	"github.com/mbark/shmake/cache"
	"github.com/mbark/shmake/loader"
)

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

func WithGlobal(name string, value starlark.Value) RunOption {
	return func(o *runOptions) {
		o.globals[name] = value
	}
}

func Run(ctx context.Context, args []string, opts ...RunOption) {
	options := runOptions{
		cacheDir: cacheHome(),
		fileName: "main.star",
		globals:  starlark.StringDict{},
	}
	for _, o := range opts {
		o(&options)
	}

	cache.SetCache(options.cacheDir)

	logger := slog.New(log.NewWithOptions(os.Stderr, log.Options{}))
	slog.SetDefault(logger)

	dir, err := findPathUpdwards("main.star")
	checkErr(err)

	err = os.Chdir(dir)
	checkErr(err)

	predeclared := createPredeclaredDict(dir)
	for name, value := range options.globals {
		predeclared[name] = value
	}

	loader.Predeclared = predeclared
	thread := &starlark.Thread{
		Name: "cli",
		Load: loader.Load,
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Println(msg)
		},
	}
	_, err = starlark.ExecFileOptions(
		&syntax.FileOptions{},
		thread,
		options.fileName,
		nil,
		predeclared,
	)
	checkErr(err)

	runCLI(ctx, args)
}

func runCLI(ctx context.Context, args []string) {
	var verbose, noCache bool
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
	}

	cmd := gCLI.Command.Command
	cmd.Flags = append(cmd.Flags, cliFlags...)
	cmd.Before = func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		if verbose {
			slog.SetLogLoggerLevel(slog.LevelDebug)
			slog.SetDefault(slog.New(slogmulti.Fanout(
				log.NewWithOptions(os.Stderr, log.Options{Level: log.DebugLevel}),
			)))
		}
		if noCache {
			cache.GlobalCache.ForceOutOfDate = noCache
		}
		return ctx, nil
	}

	err := cmd.Run(ctx, args)
	checkErr(err)

	wg.Wait()
}

// createPredeclaredDict creates the predeclared dictionary for Starlark execution.
func createPredeclaredDict(dir string) starlark.StringDict {
	return starlark.StringDict{
		"shmake": starlarkstruct.FromStringDict(starlark.String("shmake"), starlark.StringDict{
			"cli":         starlark.NewBuiltin("cli", shmakeCLI),
			"command":     starlark.NewBuiltin("command", shmakeCommand),
			"sub_command": starlark.NewBuiltin("sub_command", shmakeSubCommand),

			"shell": starlark.NewBuiltin("shell", shmakeShell),

			"string": starlark.NewBuiltin("string", shmakeString),

			"start": starlark.NewBuiltin("start", shmakeStart),
			"wait":  starlark.NewBuiltin("wait", shmakeWait),
			"pool":  starlark.NewBuiltin("pool", shmakePool),

			"load_package_json": starlark.NewBuiltin("load_package_json", shmakeLoadPackageJson),
		}),
		"cache": starlark.NewBuiltin("cache", cache.NewCacheValue),
		"current_dir": starlark.NewBuiltin("current_dir", func(
			thread *starlark.Thread,
			fn *starlark.Builtin,
			args starlark.Tuple,
			kwargs []starlark.Tuple,
		) (starlark.Value, error) {
			return starlark.String(dir), nil
		}),
	}
}

func checkErr(err error) {
	if err == nil {
		return
	}

	slog.Error(err.Error())
	os.Exit(1)
}
