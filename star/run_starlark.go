package star

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
)

func checkErr(err error) {
	if err == nil {
		return
	}

	slog.Error(err.Error())
	os.Exit(1)
}

var globals starlark.StringDict

func RunStar(args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(log.NewWithOptions(os.Stderr, log.Options{}))

	slog.SetDefault(logger)
	// Ensure we have a context before the app starts
	predeclared := starlark.StringDict{
		"shmake": starlarkstruct.FromStringDict(starlark.String("shmake"), starlark.StringDict{
			"cli":         starlark.NewBuiltin("cli", shmakeCLI),
			"command":     starlark.NewBuiltin("command", shmakeCommand),
			"sub_command": starlark.NewBuiltin("sub_command", shmakeSubCommand),
			"shell":       starlark.NewBuiltin("shell", shmakeShell),
			"string":      starlark.NewBuiltin("string", shmakeString),
		}),
		"print": starlark.NewBuiltin("print", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
			ar := make([]any, args.Len())
			for i := 0; i < args.Len(); i++ {
				if s, ok := args.Index(i).(starlark.String); ok {
					ar[i] = string(s) // unquoted, raw string
				} else {
					ar[i] = args.Index(i).String() // fallback to default
				}
			}
			fmt.Println(ar...)
			return starlark.None, nil
		}),
	}
	g, err := starlark.ExecFileOptions(
		&syntax.FileOptions{},
		&starlark.Thread{Name: "cli"},
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
			forceOutOfDate = noCache
		}
		return ctx, nil
	}

	err := cmd.Run(ctx, args)
	checkErr(err)

	wg.Wait()
}
