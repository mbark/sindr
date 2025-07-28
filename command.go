package main

import (
	"log/slog"
	"os"

	"github.com/mbark/devslog"
	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/cli/v2"
	lua "github.com/yuin/gopher-lua"
)

// Shmake global to create the cli instance.
type Shmake struct {
	Commands []*cli.Command
	Runtime  *Runtime
}

var _ LuaType = ShmakeType{}

type ShmakeType struct {
	Runtime *Runtime
}

func (s ShmakeType) TypeName() string {
	return "shmake"
}

func (s ShmakeType) GlobalName() string {
	return "shmake"
}

func (s ShmakeType) New(L *lua.LState) int {
	shmake := &Shmake{Runtime: s.Runtime}
	ud := L.NewUserData()
	ud.Value = shmake
	L.SetMetatable(ud, L.GetTypeMetatable(s.TypeName()))
	L.Push(ud)
	return 1
}

func (s ShmakeType) Funcs() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"command": s.Command,
		"run":     s.Run,
	}
}

type commandOptions struct {
	Usage string
	Flags []Flag
}

func (s ShmakeType) Command(L *lua.LState) int {
	shmake := IsUserData[*Shmake](L)
	name := L.CheckString(2)
	action := L.CheckFunction(3)

	var options commandOptions
	err := MapTable(4, L.Get(4), &options)
	if err != nil {
		L.RaiseError("invalid options: %v", err)
	}

	shmake.Commands = append(shmake.Commands, &cli.Command{
		Name:  name,
		Usage: options.Usage,
		Action: func(ctx *cli.Context) error {
			L.SetContext(ctx.Context)
			return L.CallByParam(lua.P{Fn: action, NRet: 1, Protect: true})
		},
	})

	return 0
}

func (s ShmakeType) Run(L *lua.LState) int {
	shmake := IsUserData[*Shmake](L)

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

	app := &cli.App{
		Name:     "shmake",
		Usage:    "make shmake",
		Version:  version,
		Flags:    cliFlags,
		Commands: shmake.Commands,
		Before: func(ctx *cli.Context) error {
			if verbose {
				slog.SetLogLoggerLevel(slog.LevelDebug)
				opts := &slog.HandlerOptions{Level: slog.LevelDebug}
				shmake.Runtime.logger = slog.New(slogmulti.Fanout(
					slog.NewJSONHandler(shmake.Runtime.logFile, opts),
					devslog.NewHandler(os.Stdout, &devslog.Options{HandlerOptions: opts}),
				))
			}
			if noCache {
				shmake.Runtime.cache.ForceOutOfDate = noCache
			}
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		L.RaiseError("failed to run shmake: %v", err)
	}

	return 0
}

// Flag struct to use when adding flags to a command.
type Flag struct{}

// TODO: implement
