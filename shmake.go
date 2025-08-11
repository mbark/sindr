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

	"github.com/mbark/shmake/loader"
)

var (
	globals starlark.StringDict
	cache   diskCache
)

type runOptions struct {
	cacheDir string
}

type RunOption func(*runOptions)

func WithCacheDir(dir string) RunOption {
	return func(o *runOptions) {
		o.cacheDir = dir
	}
}

func Run(args []string, opts ...RunOption) {
	options := runOptions{
		cacheDir: cacheHome(),
	}
	for _, o := range opts {
		o(&options)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache = NewCache(options.cacheDir)

	logger := slog.New(log.NewWithOptions(os.Stderr, log.Options{}))
	slog.SetDefault(logger)

	dir, err := findPathUpdwards("main.star")
	checkErr(err)

	err = os.Chdir(dir)
	checkErr(err)

	predeclared := createPredeclaredDict(dir)
	loader.Predeclared = predeclared
	thread := &starlark.Thread{
		Name: "cli",
		Load: loader.Load,
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Println(msg)
		},
	}

	g, err := starlark.ExecFileOptions(
		&syntax.FileOptions{},
		thread,
		"main.star",
		nil,
		predeclared,
	)
	checkErr(err)

	globals = g
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
			cache.ForceOutOfDate = noCache
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
		"cache": starlarkstruct.FromStringDict(starlark.String("shmake"), starlark.StringDict{
			"diff":         starlark.NewBuiltin("diff", shmakeDiff),
			"get_version":  starlark.NewBuiltin("get_version", shmakeGetVersion),
			"set_version":  starlark.NewBuiltin("set_version", shmakeSetVersion),
			"with_version": starlark.NewBuiltin("with_version", shmakeWithVersion),
		}),
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
